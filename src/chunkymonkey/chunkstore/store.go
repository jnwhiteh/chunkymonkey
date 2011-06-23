package chunkstore

import (
	"fmt"
	"os"

	. "chunkymonkey/types"
	"chunkymonkey/nbt"
)

type ChunkResult struct {
	Reader IChunkReader
	Err    os.Error
}

type IChunkStore interface {
	LoadChunk(chunkLoc *ChunkXz) (result <-chan ChunkResult)
}

type IChunkReader interface {
	// Returns the chunk location.
	ChunkLoc() *ChunkXz

	// Returns the block IDs in the chunk.
	Blocks() []byte

	// Returns the block data in the chunk.
	BlockData() []byte

	// Returns the block light data in the chunk.
	BlockLight() []byte

	// Returns the sky light data in the chunk.
	SkyLight() []byte

	// Returns the height map data in the chunk.
	HeightMap() []byte

	// Return a list of the entities (items, mobs) within the chunk.
	Entities() []*nbt.Compound

	// For low-level NBT access. Not for regular use. It's possible that this
	// might return nil if the underlying system doesn't use NBT.
	GetRootTag() *nbt.NamedTag
}

// Given the NamedTag for a level.dat, returns an appropriate ChunkStore.
func ChunkStoreForLevel(worldPath string, levelData *nbt.NamedTag) (store IChunkStore, err os.Error) {
	var serialStore iSerialChunkStore
	versionTag, ok := levelData.Lookup("/Data/version").(*nbt.Int)

	if !ok {
		serialStore = newChunkStoreAlpha(worldPath)
	} else {
		switch version := versionTag.Value; version {
		case 19132:
			serialStore = newChunkStoreBeta(worldPath)
		default:
			err = UnknownLevelVersion(version)
		}
	}

	if serialStore != nil {
		service := newChunkService(serialStore)
		store = service
		go service.serve()
	}

	return
}

type UnknownLevelVersion int32

func (err UnknownLevelVersion) String() string {
	return fmt.Sprintf("Unknown level version %d", err)
}

type NoSuchChunkError bool

func (err NoSuchChunkError) String() string {
	return "Chunk does not exist."
}
