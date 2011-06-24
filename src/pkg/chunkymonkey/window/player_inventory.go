package window

import (
	"io"
	"os"

	"chunkymonkey/inventory"
	"chunkymonkey/recipe"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

const (
	playerInvArmorNum   = 4
	playerInvMainNum    = 3 * 9
	playerInvHoldingNum = 9
)

type PlayerInventory struct {
	Window
	entityId     EntityId
	crafting     inventory.CraftingInventory
	armor        inventory.Inventory
	main         inventory.Inventory
	holding      inventory.Inventory
	holdingIndex SlotId
}

// Init initializes PlayerInventory.
// entityId - The EntityId of the player who holds the inventory.
func (w *PlayerInventory) Init(entityId EntityId, viewer IWindowViewer, recipes *recipe.RecipeSet) {
	w.entityId = entityId

	w.crafting.InitPlayerCraftingInventory(recipes)
	w.armor.Init(playerInvArmorNum)
	w.main.Init(playerInvMainNum)
	w.holding.Init(playerInvHoldingNum)
	w.Window.Init(
		WindowIdInventory,
		// Note that we have no known value for invTypeId - but it's only used
		// in WriteWindowOpen which isn't used for PlayerInventory.
		-1,
		viewer,
		"Inventory",
		&w.crafting,
		// TODO Create and use special inventory type for armor slots only.
		&w.armor,
		&w.main,
		&w.holding,
	)
	w.holdingIndex = 0
}

// Resubscribe should be called when another window has potentially been
// subscribed to the player's inventory, and subsequently closed.
func (w *PlayerInventory) Resubscribe() {
	for i := range w.Window.views {
		w.Window.views[i].Resubscribe()
	}
}

// NewWindow creates a new window for the player that shares its player
// inventory sections with `w`. Returns nil for unrecognized inventory types.
// TODO implement more inventory types.
func (w *PlayerInventory) NewWindow(invTypeId InvTypeId, windowId WindowId, inv IInventory) IWindow {
	switch invTypeId {
	case InvTypeIdWorkbench:
		return NewWindow(
			windowId, invTypeId, w.viewer, "Crafting",
			inv, &w.main, &w.holding)
	case InvTypeIdChest:
		return NewWindow(
			windowId, invTypeId, w.viewer, "Chest",
			inv, &w.main, &w.holding)
	case InvTypeIdFurnace:
		return NewWindow(
			windowId, invTypeId, w.viewer, "Furnace",
			inv, &w.main, &w.holding)
	}
	return nil
}

// SetHolding chooses the held item (0-8). Out of range values have no effect.
func (w *PlayerInventory) SetHolding(holding SlotId) {
	if holding >= 0 && holding < SlotId(playerInvHoldingNum) {
		w.holdingIndex = holding
	}
}

// HeldItem returns the slot that is the current "held" item.
// TODO need any changes to the held item slot to create notifications to
// players.
func (w *PlayerInventory) HeldItem() (slot slot.Slot, slotId SlotId) {
	return w.holding.Slot(w.holdingIndex), w.holdingIndex
}

// TakeOneHeldItem takes one item from the stack of items the player is holding
// and puts it in `into`. It does nothing if the player is holding no items, or
// if `into` cannot take any items of that type.
func (w *PlayerInventory) TakeOneHeldItem(into *slot.Slot) {
	w.holding.TakeOneItem(w.holdingIndex, into)
}

// Writes packets for other players to see the equipped items.
func (w *PlayerInventory) SendFullEquipmentUpdate(writer io.Writer) (err os.Error) {
	slot, _ := w.HeldItem()
	err = slot.SendEquipmentUpdate(writer, w.entityId, 0)
	if err != nil {
		return
	}

	numArmor := w.armor.NumSlots()
	for i := SlotId(0); i < numArmor; i++ {
		slot := w.armor.Slot(i)
		err = slot.SendEquipmentUpdate(writer, w.entityId, SlotId(i+1))
		if err != nil {
			return
		}
	}
	return
}

// PutItem attempts to put the item stack into the player's inventory. The item
// will be modified as a result.
func (w *PlayerInventory) PutItem(item *slot.Slot) {
	w.holding.PutItem(item)
	w.main.PutItem(item)
	return
}

// CanTakeItem returns true if it can take at least one item from the passed
// Slot.
func (w *PlayerInventory) CanTakeItem(item *slot.Slot) bool {
	return w.holding.CanTakeItem(item) || w.main.CanTakeItem(item)
}
