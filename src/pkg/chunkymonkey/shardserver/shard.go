package shardserver

import (
	"fmt"
	"log"
	"time"

	"chunkymonkey/chunkstore"
	. "chunkymonkey/types"
)

const chunksPerShard = ShardSize * ShardSize

// chunkXzToChunkIndex assumes that locDelta is offset relative to the shard
// origin.
func chunkXzToChunkIndex(locDelta *ChunkXz) int {
	return int(locDelta.X)*ShardSize + int(locDelta.Z)
}

// ChunkShard represents a square shard of chunks that share a master
// goroutine.
type ChunkShard struct {
	mgr              *LocalShardManager
	loc              ShardXz
	originChunkLoc   ChunkXz // The lowest X and Z located chunk in the shard.
	chunks           [chunksPerShard]*Chunk
	requests         chan iShardRequest
	ticksSinceUpdate int

	newActiveBlocks []BlockXyz
	newActiveShards map[uint64]*destActiveShard
}

func NewChunkShard(mgr *LocalShardManager, loc ShardXz) (shard *ChunkShard) {
	shard = &ChunkShard{
		mgr:              mgr,
		loc:              loc,
		originChunkLoc:   loc.ToChunkXz(),
		requests:         make(chan iShardRequest, 256),
		ticksSinceUpdate: 0,

		newActiveShards: make(map[uint64]*destActiveShard),
	}

	return
}

// serve services shard requests in the foreground.
func (shard *ChunkShard) serve() {
	ticker := time.NewTicker(NanosecondsInSecond / TicksPerSecond)

	for {
		select {
		case <-ticker.C:
			shard.tick()

		case request := <-shard.requests:
			request.perform(shard)
		}
	}
}

// tick runs the shard for a single tick.
func (shard *ChunkShard) tick() {
	shard.ticksSinceUpdate++

	for _, chunk := range shard.chunks {
		if chunk != nil {
			chunk.tick()
		}
	}

	if shard.ticksSinceUpdate >= TicksPerSecond {
		for _, chunk := range shard.chunks {
			if chunk != nil {
				chunk.sendUpdate()
			}
		}
		shard.ticksSinceUpdate = 0
	}

	shard.transferActiveBlocks()
}

// transferActiveBlocks takes blocks marked as newly active by addActiveBlock,
// and informs the chunk in the destination shards.
func (shard *ChunkShard) transferActiveBlocks() {
	if len(shard.newActiveBlocks) == 0 {
		return
	}

	thisShardKey := shard.loc.Key()
	for shardKey, activeShard := range shard.newActiveShards {
		if shardKey == thisShardKey {
			shard.reqSetBlocksActive(activeShard.blocks)
		} else {
			// TODO For now we only open a connection per request and then discard
			// it. It will likely be worth holding open connections at least to
			// neighbouring shards at some point.
			destShardClient := shard.mgr.ShardShardConnect(activeShard.loc)
			defer destShardClient.Disconnect()
			destShardClient.ReqSetActiveBlocks(activeShard.blocks)
		}
	}
}

// reqSetBlocksActive sets each block in the given slice to be active within
// the chunk. Note: if a block is within a different shard, it is discarded.
func (shard *ChunkShard) reqSetBlocksActive(blocks []BlockXyz) {
	for _, block := range blocks {
		chunkXz := block.ToChunkXz()
		chunkIndex, _, _, isThisShard := shard.chunkIndexAndRelLoc(chunkXz)
		if isThisShard {
			chunk := shard.chunks[chunkIndex]
			if chunk == nil {
				continue
			}
			chunk.AddActiveBlock(&block)
		}
	}
}

// addActiveBlock sets the given block to be active on the next tick. This
// works even if the block is not within the shard - it will be made active
// provided that the chunk that the block is within is loaded.
func (shard *ChunkShard) addActiveBlock(block *BlockXyz) {
	chunkXz := block.ToChunkXz()
	shardXz := chunkXz.ToShardXz()
	shardKey := shardXz.Key()
	activeShard, ok := shard.newActiveShards[shardKey]
	if ok {
		activeShard = &destActiveShard{
			loc:    shardXz,
			blocks: []BlockXyz{*block},
		}
		shard.newActiveShards[shardKey] = activeShard
	} else {
		activeShard.blocks = append(activeShard.blocks, *block)
	}
}

func (shard *ChunkShard) String() string {
	return fmt.Sprintf("ChunkShard[%#v/%#v]", shard.loc, shard.originChunkLoc)
}

func (shard *ChunkShard) chunkIndexAndRelLoc(loc *ChunkXz) (index int, x, z ChunkCoord, ok bool) {
	x = loc.X - shard.originChunkLoc.X
	z = loc.Z - shard.originChunkLoc.Z

	if x < 0 || z < 0 || x >= ShardSize || z >= ShardSize {
		log.Printf("%v.chunkIndexAndRelLoc(%#v): ChunkXz outside of shard", shard, loc)
		return 0, 0, 0, false
	}

	ok = true

	index = int(x*ShardSize + z)

	return
}

// Get returns the Chunk at at given coordinates, loading it if it is not
// already loaded.
// TODO make this method private?
func (shard *ChunkShard) Get(loc *ChunkXz) *Chunk {
	chunkIndex, dx, dz, ok := shard.chunkIndexAndRelLoc(loc)
	if !ok {
		return nil
	}

	chunk := shard.chunks[chunkIndex]

	// Chunk already loaded.
	if chunk != nil {
		return chunk
	}

	chunk = shard.loadChunk(loc, &ChunkXz{dx, dz})

	if chunk == nil {
		// No chunk available at that location. (Return nil explicitly - interfaces
		// containing nil pointer don't equal nil).
		return nil
	}

	shard.chunks[chunkIndex] = chunk

	return chunk
}

// loadChunk loads the specified chunk from store, and returns it.
// loc - The absolute world position of the chunk.
// locDelta - The relative position of the chunk within the shard.
func (shard *ChunkShard) loadChunk(loc *ChunkXz, locDelta *ChunkXz) *Chunk {
	chunkResult := <-shard.mgr.chunkStore.LoadChunk(loc)
	chunkReader, err := chunkResult.Reader, chunkResult.Err
	if err != nil {
		if _, ok := err.(chunkstore.NoSuchChunkError); !ok {
			log.Printf("%v.load(%#v): chunk loading error: %v", shard, loc, err)
			return nil
		} else {
			// Chunk doesn't exist in store.
			// TODO Generate new chunks.
			return nil
		}
	}

	chunk := newChunkFromReader(chunkReader, shard.mgr, shard)

	// Notify neighbouring chunk(s) (if any) that this chunk is now active, and
	// notify this chunk of its active neighbours
	linkNeighbours := func(from ChunkSideDir) {
		dx, dz := from.GetDxz()
		nLoc := ChunkXz{
			X: locDelta.X + dx,
			Z: locDelta.Z + dz,
		}
		if nLoc.X < 0 || nLoc.Z < 0 || nLoc.X > ShardSize || nLoc.Z > ShardSize {
			// Link to neighbouring chunk outside the shard.
			// TODO This should also link to chunks outside the shard. Although the
			// architecure of this is likely to change radically or go away.
		} else {
			// Link to neighbouring chunk within the shard.
			chunkIndex := chunkXzToChunkIndex(locDelta)
			neighbour := shard.chunks[chunkIndex]
			if neighbour != nil {
				to := from.GetOpposite()
				chunk.sideCacheSetNeighbour(from, neighbour)
				neighbour.sideCacheSetNeighbour(to, chunk)
			}
		}
	}
	// TODO Corresponding unlinking when a chunk is unloaded.
	linkNeighbours(ChunkSideEast)
	linkNeighbours(ChunkSideSouth)
	linkNeighbours(ChunkSideWest)
	linkNeighbours(ChunkSideNorth)

	return chunk
}

// TODO Make all Enqueue* methods private.

// EnqueueAllChunks runs a given function on all loaded chunks in the shard.
func (shard *ChunkShard) EnqueueAllChunks(fn func(chunk *Chunk)) {
	shard.requests <- &runOnAllChunks{fn}
}

// EnqueueOnChunk runs a function on the chunk at the given location. If the
// chunk does not exist, it does nothing.
func (shard *ChunkShard) EnqueueOnChunk(loc ChunkXz, fn func(chunk *Chunk)) {
	shard.requests <- &runOnChunk{loc, fn}
}

func (shard *ChunkShard) Enqueue(fn func()) {
	shard.requests <- &runGeneric{fn}
}

func (shard *ChunkShard) enqueueRequest(req iShardRequest) {
	shard.requests <- req
}

type destActiveShard struct {
	loc    ShardXz
	blocks []BlockXyz
}
