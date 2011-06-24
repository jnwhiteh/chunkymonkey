package chunkstore

import (
	"os"

	. "chunkymonkey/types"
)

type request struct {
	chunkLoc     *ChunkXz
	responseChan chan<- ChunkResult
}

type iSerialChunkStore interface {
	loadChunk(chunkLoc *ChunkXz) (reader IChunkReader, err os.Error)
}

// chunkService adapts an iSerialStore (which can only be accessed from one
// goroutine) to an IChunkStore.
type chunkService struct {
	store    iSerialChunkStore
	requests chan request
}

func newChunkService(store iSerialChunkStore) *chunkService {
	return &chunkService{
		store:    store,
		requests: make(chan request),
	}
}

// serve() serves LoadChunk() requests in the foreground.
func (s *chunkService) serve() {
	for {
		request := <-s.requests
		reader, err := s.store.loadChunk(request.chunkLoc)
		request.responseChan <- ChunkResult{reader, err}
	}
}

func (s *chunkService) LoadChunk(chunkLoc *ChunkXz) <-chan ChunkResult {
	responseChan := make(chan ChunkResult)

	s.requests <- request{
		chunkLoc:     chunkLoc,
		responseChan: responseChan,
	}

	return responseChan
}
