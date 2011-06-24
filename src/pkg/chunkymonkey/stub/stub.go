package stub

import (
	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// IShardConnecter is used to look up shards and connect to them.
type IShardConnecter interface {
	// PlayerShardConnect makes a connection from a player to a shard.
	PlayerShardConnect(entityId EntityId, player IShardPlayerClient, shardLoc ShardXz) IPlayerShardClient

	// ShardShardConnect makes a connection from one shard to another.
	// TODO Consider making this package-private to shardserver.
	ShardShardConnect(shardLoc ShardXz) IShardShardClient
}

// IShardPlayerClient is the interface by which shards communicate to players on
// the frontend.
type IShardPlayerClient interface {
	GetEntityId() EntityId

	TransmitPacket(packet []byte)

	// ReqInventorySubscribed informs the player that an inventory has been
	// opened. The block position 
	ReqInventorySubscribed(block BlockXyz, invTypeId InvTypeId, slots []proto.WindowSlot)

	// ReqInventorySlotUpdate informs the player of a change to a slot in the
	// open inventory.
	ReqInventorySlotUpdate(block BlockXyz, slot slot.Slot, slotId SlotId)

	// ReqInventoryProgressUpdate informs the player of a change of a progress
	// bar in a window.
	ReqInventoryProgressUpdate(block BlockXyz, prgBarId PrgBarId, value PrgBarValue)

	// ReqInventoryCursorUpdate informs the player of their new cursor contents.
	ReqInventoryCursorUpdate(block BlockXyz, cursor slot.Slot)

	// ReqInventoryTxState requests that the player report the transaction state
	// as accepted or not. This is used by remote inventories when
	// TxStateDeferred is returned from Click.
	ReqInventoryTxState(block BlockXyz, txId TxId, accepted bool)

	// ReqInventorySubscribed informs the player that an inventory has been
	// closed.
	ReqInventoryUnsubscribed(block BlockXyz)

	// ReqPlaceHeldItem requests that the player frontend take one item from the
	// held item stack and send it in a ReqPlaceItem to the target block.  The
	// player code may *not* honour this request (e.g there might be no suitable
	// held item).
	ReqPlaceHeldItem(target BlockXyz, wasHeld slot.Slot)

	// ReqOfferItem requests that the player check if it can take the item.  If
	// it can then it should ReqTakeItem from the chunk.
	ReqOfferItem(fromChunk ChunkXz, entityId EntityId, item slot.Slot)

	// ReqGiveItem requests that the player takes the item contents into their
	// inventory. If they cannot, then the player should drop the item at the
	// given position.
	ReqGiveItem(atPosition AbsXyz, item slot.Slot)
}

// IPlayerShardClient is the interface by which shards can be communicated to by
// player frontend code.
type IPlayerShardClient interface {
	// Removes connection to shard, and removes all subscriptions to chunks in
	// the shard. Note that this does *not* send packets to tell the client to
	// unload the subscribed chunks.
	Disconnect()

	// TODO better method to send events to chunks from player frontend.
	Enqueue(fn func())

	// The following methods are requests upon chunks.

	ReqSubscribeChunk(chunkLoc ChunkXz)

	ReqUnsubscribeChunk(chunkLoc ChunkXz)

	ReqMulticastPlayers(chunkLoc ChunkXz, exclude EntityId, packet []byte)

	ReqAddPlayerData(chunkLoc ChunkXz, name string, position AbsXyz, look LookBytes, held ItemTypeId)

	ReqRemovePlayerData(chunkLoc ChunkXz, isDisconnect bool)

	ReqSetPlayerPositionLook(chunkLoc ChunkXz, position AbsXyz, look LookBytes, moved bool)

	// ReqHitBlock requests that the targetted block be hit.
	ReqHitBlock(held slot.Slot, target BlockXyz, digStatus DigStatus, face Face)

	// ReqHitBlock requests that the targetted block be interacted with.
	ReqInteractBlock(held slot.Slot, target BlockXyz, face Face)

	// ReqPlaceItem requests that the item passed be placed at the given target
	// location. The shard *may* choose not to do this, but if it cannot, then it
	// *must* account for the item in some way (maybe hand it back to the player
	// or just drop it on the ground).
	ReqPlaceItem(target BlockXyz, slot slot.Slot)

	// ReqTakeItem requests that the item with the specified entityId is given to
	// the player. The chunk doesn't have to respect this (particularly if the
	// item no longer exists).
	ReqTakeItem(chunkLoc ChunkXz, entityId EntityId)

	// ReqDropItem requests that an item be created.
	ReqDropItem(content slot.Slot, position AbsXyz, velocity AbsVelocity)

	// ReqInventoryClick requests that the given cursor be "clicked" onto the
	// inventory. The chunk should send a replying ReqInventoryCursorUpdate to
	// reflect the new state of the cursor afterwards - in addition to any
	// ReqInventorySlotUpdate to all subscribers to the inventory.
	ReqInventoryClick(block BlockXyz, slotId SlotId, cursor slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot slot.Slot)

	// ReqInventoryUnsubscribed requests that the inventory for the block be
	// unsubscribed to.
	ReqInventoryUnsubscribed(block BlockXyz)
}

// IShardShardClient provides an interface for shards to make requests against
// another shard.
// TODO Consider making this package-private to shardserver.
type IShardShardClient interface {
	Disconnect()

	ReqSetActiveBlocks(blocks []BlockXyz)
}
