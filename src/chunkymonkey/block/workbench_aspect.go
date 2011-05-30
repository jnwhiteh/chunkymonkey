package block

import (
	"chunkymonkey/inventory"
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
)

func makeWorkbenchAspect() (aspect IBlockAspect) {
	return &WorkbenchAspect{}
}

// WorkbenchAspect is the behaviour for the workbench block that allows 3x3
// crafting.
type WorkbenchAspect struct {
	StandardAspect
}

func (aspect *WorkbenchAspect) Name() string {
	return "Workbench"
}

func (aspect *WorkbenchAspect) Interact(instance *BlockInstance, player stub.IPlayerConnection) {
	extra := aspect.invWrapper(instance, true)
	if extra != nil {
		extra.AddSubscriber(player)
	}
}

func (aspect *WorkbenchAspect) InventoryClick(instance *BlockInstance, player stub.IPlayerConnection, slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) {
	extra := aspect.invWrapper(instance, false)
	if extra != nil {
		extra.Click(player, slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
	} else {
		// No inventory to act on (shouldn't happen, normally).
		player.ReqInventoryTxState(extra.instance.BlockLoc, txId, false)
		player.ReqInventoryCursorUpdate(instance.BlockLoc, *cursor)
		return
	}
}

func (aspect *WorkbenchAspect) InventoryUnsubscribed(instance *BlockInstance, player stub.IPlayerConnection) {
	extra := aspect.invWrapper(instance, false)
	if extra != nil {
		extra.RemoveSubscriber(player.GetEntityId())
	}
}

func (aspect *WorkbenchAspect) Destroy(instance *BlockInstance) {
	extra := aspect.invWrapper(instance, false)
	if extra != nil {
		extra.EjectItems()
		extra.Destroyed()
	}

	aspect.StandardAspect.Destroy(instance)
}

func (aspect *WorkbenchAspect) invWrapper(instance *BlockInstance, create bool) *workbenchExtra {
	extra, ok := instance.Chunk.BlockExtra(&instance.SubLoc).(*workbenchExtra)
	if !ok && create {
		extra = newWorkbenchExtra(instance)
		instance.Chunk.SetBlockExtra(&instance.SubLoc, extra)
	}

	return extra
}


// workbenchExtra is the data stored in Chunk.SetBlockExtra. It also implements
// IInventorySubscriber to relay events to player(s) subscribed.
type workbenchExtra struct {
	instance    BlockInstance
	inv         inventory.CraftingInventory
	subscribers map[EntityId]stub.IPlayerConnection
}

func newWorkbenchExtra(instance *BlockInstance) *workbenchExtra {
	extra := &workbenchExtra{
		instance:    *instance,
		subscribers: make(map[EntityId]stub.IPlayerConnection),
	}

	extra.inv.InitWorkbenchInventory(instance.Chunk.RecipeSet())
	extra.inv.SetSubscriber(extra)

	return extra
}

func (extra *workbenchExtra) Click(player stub.IPlayerConnection, slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) {
	txState := extra.inv.Click(slotId, cursor, rightClick, shiftClick, txId, expectedSlot)

	player.ReqInventoryCursorUpdate(extra.instance.BlockLoc, *cursor)

	// Inform client of operation status.
	player.ReqInventoryTxState(extra.instance.BlockLoc, txId, txState == TxStateAccepted)
}

func (extra *workbenchExtra) SlotUpdate(slot *slot.Slot, slotId SlotId) {
	for _, subscriber := range extra.subscribers {
		subscriber.ReqInventorySlotUpdate(extra.instance.BlockLoc, *slot, slotId)
	}
}

func (extra *workbenchExtra) AddSubscriber(player stub.IPlayerConnection) {
	entityId := player.GetEntityId()
	extra.subscribers[entityId] = player

	// Register self for automatic removal when IPlayerConnection unsubscribes
	// from the chunk.
	extra.instance.Chunk.AddOnUnsubscribe(entityId, extra)

	slots := extra.inv.MakeProtoSlots()

	player.ReqInventorySubscribed(extra.instance.BlockLoc, InvTypeIdWorkbench, slots)
}

func (extra *workbenchExtra) RemoveSubscriber(entityId EntityId) {
	extra.subscribers[entityId] = nil, false
	extra.instance.Chunk.RemoveOnUnsubscribe(entityId, extra)
	if len(extra.subscribers) == 0 {
		extra.EjectItems()
	}
}

func (extra *workbenchExtra) Destroyed() {
	for _, subscriber := range extra.subscribers {
		subscriber.ReqInventoryUnsubscribed(extra.instance.BlockLoc)
		extra.instance.Chunk.RemoveOnUnsubscribe(subscriber.GetEntityId(), extra)
	}
	extra.subscribers = nil
}

// Unsubscribed implements block.IUnsubscribed. It removes a player's
// subscription to the inventory when they unsubscribe from the chunk.
func (extra *workbenchExtra) Unsubscribed(entityId EntityId) {
	extra.subscribers[entityId] = nil, false
}

// EjectItems removes all items from the inventory and drops them at the
// location of the block.
func (extra *workbenchExtra) EjectItems() {
	items := extra.inv.TakeAllItems()

	for _, slot := range items {
		spawnItemInBlock(&extra.instance, slot.ItemType, slot.Count, slot.Data)
	}
}
