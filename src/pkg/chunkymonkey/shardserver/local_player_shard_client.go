package shardserver

import (
	"chunkymonkey/stub"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// localPlayerShardClient implements IPlayerShardClient for LocalShardManager.
type localPlayerShardClient struct {
	entityId EntityId
	player   stub.IShardPlayerClient
	shard    *ChunkShard
}

func newLocalPlayerShardClient(entityId EntityId, player stub.IShardPlayerClient, shard *ChunkShard) *localPlayerShardClient {
	return &localPlayerShardClient{
		entityId: entityId,
		player:   player,
		shard:    shard,
	}
}

func (conn *localPlayerShardClient) Disconnect() {
	// TODO This inefficiently unsubscribes from all chunks, even if not
	// subscribed to.
	conn.shard.EnqueueAllChunks(func(chunk *Chunk) {
		chunk.reqUnsubscribeChunk(conn.entityId, false)
	})
}

func (conn *localPlayerShardClient) Enqueue(fn func()) {
	conn.shard.Enqueue(fn)
}

func (conn *localPlayerShardClient) ReqSubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqSubscribeChunk(conn.entityId, conn.player)
	})
}

func (conn *localPlayerShardClient) ReqUnsubscribeChunk(chunkLoc ChunkXz) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqUnsubscribeChunk(conn.entityId, true)
	})
}

func (conn *localPlayerShardClient) ReqMulticastPlayers(chunkLoc ChunkXz, exclude EntityId, packet []byte) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqMulticastPlayers(exclude, packet)
	})
}

func (conn *localPlayerShardClient) ReqAddPlayerData(chunkLoc ChunkXz, name string, position AbsXyz, look LookBytes, held ItemTypeId) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqAddPlayerData(conn.entityId, name, position, look, held)
	})
}

func (conn *localPlayerShardClient) ReqRemovePlayerData(chunkLoc ChunkXz, isDisconnect bool) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqRemovePlayerData(conn.entityId, isDisconnect)
	})
}

func (conn *localPlayerShardClient) ReqSetPlayerPositionLook(chunkLoc ChunkXz, position AbsXyz, look LookBytes, moved bool) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqSetPlayerPositionLook(conn.entityId, position, look, moved)
	})
}

func (conn *localPlayerShardClient) ReqHitBlock(held slot.Slot, target BlockXyz, digStatus DigStatus, face Face) {
	chunkLoc := target.ToChunkXz()

	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqHitBlock(conn.player, held, digStatus, &target, face)
	})
}

func (conn *localPlayerShardClient) ReqInteractBlock(held slot.Slot, target BlockXyz, face Face) {
	chunkLoc := target.ToChunkXz()

	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqInteractBlock(conn.player, held, &target, face)
	})
}

func (conn *localPlayerShardClient) ReqPlaceItem(target BlockXyz, slot slot.Slot) {
	chunkLoc, _ := target.ToChunkLocal()

	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqPlaceItem(conn.player, &target, &slot)
	})
}

func (conn *localPlayerShardClient) ReqTakeItem(chunkLoc ChunkXz, entityId EntityId) {
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqTakeItem(conn.player, entityId)
	})
}

func (conn *localPlayerShardClient) ReqDropItem(content slot.Slot, position AbsXyz, velocity AbsVelocity) {
	chunkLoc := position.ToChunkXz()
	conn.shard.EnqueueOnChunk(chunkLoc, func(chunk *Chunk) {
		chunk.reqDropItem(conn.player, &content, &position, &velocity)
	})
}

func (conn *localPlayerShardClient) ReqInventoryClick(block BlockXyz, slotId SlotId, cursor slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot slot.Slot) {
	chunkLoc := block.ToChunkXz()
	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqInventoryClick(conn.player, &block, slotId, &cursor, rightClick, shiftClick, txId, &expectedSlot)
	})
}

func (conn *localPlayerShardClient) ReqInventoryUnsubscribed(block BlockXyz) {
	chunkLoc := block.ToChunkXz()
	conn.shard.EnqueueOnChunk(*chunkLoc, func(chunk *Chunk) {
		chunk.reqInventoryUnsubscribed(conn.player, &block)
	})
}
