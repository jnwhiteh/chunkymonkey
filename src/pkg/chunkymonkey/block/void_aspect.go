package block

import (
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
)

func makeVoidAspect() (aspect IBlockAspect) {
	return &VoidAspect{}
}

// Behaviour of a "void" block which has no behaviour.
type VoidAspect struct{}

func (aspect *VoidAspect) setAttrs(blockAttrs *BlockAttrs) {
}

func (aspect *VoidAspect) Name() string {
	return "Void"
}

func (aspect *VoidAspect) Hit(instance *BlockInstance, player stub.IShardPlayerClient, digStatus DigStatus) (destroyed bool) {
	destroyed = false
	return
}

func (aspect *VoidAspect) Interact(instance *BlockInstance, player stub.IShardPlayerClient) {
}

func (aspect *VoidAspect) InventoryClick(instance *BlockInstance, player stub.IShardPlayerClient, slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) {
}

func (aspect *VoidAspect) InventoryUnsubscribed(instance *BlockInstance, player stub.IShardPlayerClient) {
}

func (aspect *VoidAspect) Destroy(instance *BlockInstance) {
}

func (aspect *VoidAspect) Tick(instance *BlockInstance) bool {
	return false
}
