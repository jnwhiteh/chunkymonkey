package block

import (
	"chunkymonkey/inventory"
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
)


// blockInventory is the data stored in Chunk.SetBlockExtra by some block
// aspects that contain inventories. It also implements IInventorySubscriber to
// relay events to player(s) subscribed to the inventories.
type blockInventory struct {
	instance           BlockInstance
	inv                inventory.IInventory
	subscribers        map[EntityId]stub.IShardPlayerClient
	ejectOnUnsubscribe bool
	invTypeId          InvTypeId
}

// newBlockInventory creates a new blockInventory.
func newBlockInventory(instance *BlockInstance, inv inventory.IInventory, ejectOnUnsubscribe bool, invTypeId InvTypeId) *blockInventory {
	blkInv := &blockInventory{
		instance:           *instance,
		inv:                inv,
		subscribers:        make(map[EntityId]stub.IShardPlayerClient),
		ejectOnUnsubscribe: ejectOnUnsubscribe,
		invTypeId:          invTypeId,
	}

	blkInv.inv.SetSubscriber(blkInv)

	return blkInv
}

func (blkInv *blockInventory) Click(player stub.IShardPlayerClient, slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) {
	txState := blkInv.inv.Click(slotId, cursor, rightClick, shiftClick, txId, expectedSlot)

	player.ReqInventoryCursorUpdate(blkInv.instance.BlockLoc, *cursor)

	// Inform client of operation status.
	player.ReqInventoryTxState(blkInv.instance.BlockLoc, txId, txState == TxStateAccepted)
}

func (blkInv *blockInventory) SlotUpdate(slot *slot.Slot, slotId SlotId) {
	for _, subscriber := range blkInv.subscribers {
		subscriber.ReqInventorySlotUpdate(blkInv.instance.BlockLoc, *slot, slotId)
	}
}

func (blkInv *blockInventory) ProgressUpdate(prgBarId PrgBarId, value PrgBarValue) {
	for _, subscriber := range blkInv.subscribers {
		subscriber.ReqInventoryProgressUpdate(blkInv.instance.BlockLoc, prgBarId, value)
	}
}

func (blkInv *blockInventory) AddSubscriber(player stub.IShardPlayerClient) {
	entityId := player.GetEntityId()
	blkInv.subscribers[entityId] = player

	// Register self for automatic removal when IShardPlayerClient unsubscribes
	// from the chunk.
	blkInv.instance.Chunk.AddOnUnsubscribe(entityId, blkInv)

	slots := blkInv.inv.MakeProtoSlots()

	player.ReqInventorySubscribed(blkInv.instance.BlockLoc, blkInv.invTypeId, slots)
}

func (blkInv *blockInventory) RemoveSubscriber(entityId EntityId) {
	blkInv.subscribers[entityId] = nil, false
	blkInv.instance.Chunk.RemoveOnUnsubscribe(entityId, blkInv)
	if blkInv.ejectOnUnsubscribe && len(blkInv.subscribers) == 0 {
		blkInv.EjectItems()
	}
}

func (blkInv *blockInventory) Destroyed() {
	for _, subscriber := range blkInv.subscribers {
		subscriber.ReqInventoryUnsubscribed(blkInv.instance.BlockLoc)
		blkInv.instance.Chunk.RemoveOnUnsubscribe(subscriber.GetEntityId(), blkInv)
	}
	blkInv.subscribers = nil
}

// Unsubscribed implements block.IUnsubscribed. It removes a player's
// subscription to the inventory when they unsubscribe from the chunk.
func (blkInv *blockInventory) Unsubscribed(entityId EntityId) {
	blkInv.subscribers[entityId] = nil, false
}

// EjectItems removes all items from the inventory and drops them at the
// location of the block.
func (blkInv *blockInventory) EjectItems() {
	items := blkInv.inv.TakeAllItems()

	for _, slot := range items {
		spawnItemInBlock(&blkInv.instance, slot.ItemType, slot.Count, slot.Data)
	}
}
