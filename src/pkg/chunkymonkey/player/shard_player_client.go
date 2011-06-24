package player

import (
	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

// shardPlayerClient receives events from chunk shards and acts upon them. It
// implements stub.IShardPlayerClient.
type shardPlayerClient struct {
	player *Player
}

func (psr *shardPlayerClient) Init(player *Player) {
	psr.player = player
}

func (psr *shardPlayerClient) GetEntityId() EntityId {
	return psr.player.EntityId
}

func (psr *shardPlayerClient) TransmitPacket(packet []byte) {
	psr.player.TransmitPacket(packet)
}

func (psr *shardPlayerClient) ReqInventorySubscribed(block BlockXyz, invTypeId InvTypeId, slots []proto.WindowSlot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventorySubscribed(&block, invTypeId, slots)
	})
}

func (psr *shardPlayerClient) ReqInventorySlotUpdate(block BlockXyz, slot slot.Slot, slotId SlotId) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventorySlotUpdate(&block, &slot, slotId)
	})
}

func (psr *shardPlayerClient) ReqInventoryProgressUpdate(block BlockXyz, prgBarId PrgBarId, value PrgBarValue) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryProgressUpdate(&block, prgBarId, value)
	})
}

func (psr *shardPlayerClient) ReqInventoryCursorUpdate(block BlockXyz, cursor slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryCursorUpdate(&block, &cursor)
	})
}

func (psr *shardPlayerClient) ReqInventoryTxState(block BlockXyz, txId TxId, accepted bool) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryTxState(&block, txId, accepted)
	})
}

func (psr *shardPlayerClient) ReqInventoryUnsubscribed(block BlockXyz) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqInventoryUnsubscribed(&block)
	})
}

func (psr *shardPlayerClient) ReqPlaceHeldItem(target BlockXyz, wasHeld slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqPlaceHeldItem(&target, &wasHeld)
	})
}

func (psr *shardPlayerClient) ReqOfferItem(fromChunk ChunkXz, entityId EntityId, item slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqOfferItem(&fromChunk, entityId, &item)
	})
}

func (psr *shardPlayerClient) ReqGiveItem(atPosition AbsXyz, item slot.Slot) {
	psr.player.Enqueue(func(_ *Player) {
		psr.player.reqGiveItem(&atPosition, &item)
	})
}
