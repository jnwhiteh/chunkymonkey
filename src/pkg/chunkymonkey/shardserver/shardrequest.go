package shardserver

import (
	. "chunkymonkey/types"
)

// iShardRequest represents a request for the ChunkShard to perform.
type iShardRequest interface {
	// Executes the request.
	perform(shard *ChunkShard)
}

// Various types of iShardRequest types follow.

// runOnChunk runs a function on a specific chunk.
type runOnChunk struct {
	loc ChunkXz
	fn  func(chunk *Chunk)
}

func (req *runOnChunk) perform(shard *ChunkShard) {
	chunk := shard.Get(&req.loc)
	if chunk != nil {
		req.fn(chunk)
	}
}

// runOnChunk runs a function on all loaded chunks in a shard.
type runOnAllChunks struct {
	fn func(chunk *Chunk)
}

func (req *runOnAllChunks) perform(shard *ChunkShard) {
	for _, chunk := range shard.chunks {
		if chunk != nil {
			req.fn(chunk)
		}
	}
}

// runGeneric runs a function.
type runGeneric struct {
	fn func()
}

func (req *runGeneric) perform(shard *ChunkShard) {
	req.fn()
}
