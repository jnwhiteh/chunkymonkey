package player

import (
	"chunkymonkey/inventory"
	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

type RemoteInventory struct {
	blockLoc   BlockXyz
	chunkSubs  *chunkSubscriptions
	slots      []proto.WindowSlot
	subscriber inventory.IInventorySubscriber
}

func NewRemoteInventory(block *BlockXyz, chunkSubs *chunkSubscriptions, slots []proto.WindowSlot) *RemoteInventory {
	return &RemoteInventory{
		blockLoc:   *block,
		chunkSubs:  chunkSubs,
		slots:      slots,
		subscriber: nil,
	}
}

func (inv *RemoteInventory) IsForBlock(block *BlockXyz) bool {
	return inv.blockLoc.X == block.X && inv.blockLoc.Y == block.Y && inv.blockLoc.Z == block.Z
}

func (inv *RemoteInventory) slotUpdate(slot *slot.Slot, slotId SlotId) {
	if inv.subscriber != nil {
		inv.subscriber.SlotUpdate(slot, slotId)
	}
}

func (inv *RemoteInventory) progressUpdate(prgBarId PrgBarId, value PrgBarValue) {
	if inv.subscriber != nil {
		inv.subscriber.ProgressUpdate(prgBarId, value)
	}
}

func (inv *RemoteInventory) Close() {
	shard, _, ok := inv.chunkSubs.ShardClientForBlockXyz(&inv.blockLoc)

	if ok {
		shard.ReqInventoryUnsubscribed(inv.blockLoc)
	}
}

// The following methods are to implement window.IInventory.

func (inv *RemoteInventory) NumSlots() SlotId {
	return SlotId(len(inv.slots))
}

func (inv *RemoteInventory) Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) (txState TxState) {
	shard, _, ok := inv.chunkSubs.ShardClientForBlockXyz(&inv.blockLoc)

	if ok {
		shard.ReqInventoryClick(inv.blockLoc, slotId, *cursor, rightClick, shiftClick, txId, *expectedSlot)
	}

	return TxStateDeferred
}

func (inv *RemoteInventory) SetSubscriber(subscriber inventory.IInventorySubscriber) {
	inv.subscriber = subscriber
}

func (inv *RemoteInventory) WriteProtoSlots(slots []proto.WindowSlot) {
	// Note that this only produces accurate results before any slot updates come
	// through. inv.slots contains only a snapshot of the state when the window
	// opened.
	copy(slots, inv.slots)
}
