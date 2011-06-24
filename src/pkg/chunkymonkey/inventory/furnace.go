package inventory

import (
	"log"

	"chunkymonkey/itemtype"
	"chunkymonkey/recipe"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

const (
	furnaceSlotReagent = SlotId(0)
	furnaceSlotFuel    = SlotId(1)
	furnaceSlotOutput  = SlotId(2)
	furnaceNumSlots    = 3

	reactionDuration = Ticks(185)
	maxFuelPrg       = 255
)

type FurnaceInventory struct {
	Inventory
	furnaceData       *recipe.FurnaceData
	itemTypes         itemtype.ItemTypeMap
	maxFuel           Ticks
	curFuel           Ticks
	reactionRemaining Ticks

	lastCurFuel           PrgBarValue
	lastReactionRemaining PrgBarValue
	ticksSinceUpdate      int
}

// NewFurnaceInventory creates a furnace inventory.
func NewFurnaceInventory(furnaceData *recipe.FurnaceData, itemTypes itemtype.ItemTypeMap) (inv *FurnaceInventory) {
	inv = &FurnaceInventory{
		furnaceData:       furnaceData,
		itemTypes:         itemTypes,
		maxFuel:           0,
		curFuel:           0,
		reactionRemaining: reactionDuration,
	}
	inv.Inventory.Init(furnaceNumSlots)
	return
}

func (inv *FurnaceInventory) Click(slotId SlotId, cursor *slot.Slot, rightClick bool, shiftClick bool, txId TxId, expectedSlot *slot.Slot) (txState TxState) {

	switch slotId {
	case furnaceSlotReagent:
		slotBefore := inv.slots[furnaceSlotReagent]

		txState = inv.Inventory.Click(
			slotId, cursor, rightClick, shiftClick, txId, expectedSlot)

		slotAfter := &inv.slots[furnaceSlotReagent]

		// If the reagent type changes, the reaction restarts.
		if slotBefore.ItemType != slotAfter.ItemType || slotBefore.Data != slotAfter.Data {
			inv.reactionRemaining = reactionDuration
		}
	case furnaceSlotFuel:
		cursorItemId := cursor.GetItemTypeId()
		_, cursorIsFuel := inv.furnaceData.Fuels[cursorItemId]
		if cursorIsFuel || cursor.IsEmpty() {
			txState = inv.Inventory.Click(
				slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
		}
	case furnaceSlotOutput:
		// Player may only *take* the *whole* stack from the output slot.
		txState = inv.Inventory.TakeOnlyClick(
			slotId, cursor, rightClick, shiftClick, txId, expectedSlot)
	}

	inv.stateCheck()

	inv.sendProgressUpdates()

	return
}

func (inv *FurnaceInventory) stateCheck() {
	reagentSlot := &inv.slots[furnaceSlotReagent]
	fuelSlot := &inv.slots[furnaceSlotFuel]
	outputSlot := &inv.slots[furnaceSlotOutput]

	reaction, haveReagent := inv.furnaceData.Reactions[reagentSlot.GetItemTypeId()]
	fuelTicks, haveFuel := inv.furnaceData.Fuels[fuelSlot.GetItemTypeId()]

	// Work out if the output slot is ready for items to be produced from the
	// reaction.
	var outputReady bool
	if outputSlot.ItemType != nil {
		// Output has items in.
		if !haveReagent {
			outputReady = false
		} else if outputSlot.Count >= outputSlot.ItemType.MaxStack {
			// Output is full.
			outputReady = false
		} else if outputSlot.GetItemTypeId() != reaction.Output || outputSlot.Data != reaction.OutputData {
			// Output has a different type from the reaction.
			outputReady = false
		} else {
			// Output contains compatible items and is not full.
			outputReady = true
		}
	} else {
		// Output is empty.
		outputReady = true
	}

	if inv.curFuel > 0 {
		// Furnace is lit.
		if !outputReady {
			inv.reactionRemaining = reactionDuration
		} else if haveReagent && inv.reactionRemaining == 0 {
			// One reaction complete.
			if itemType, ok := inv.itemTypes[reaction.Output]; !ok {
				log.Printf("Furnace encountered unknown output type in reaction %#v", reaction)
			} else {
				itemCreated := slot.Slot{
					ItemType: itemType,
					Count:    1,
					Data:     reaction.OutputData,
				}
				inv.reactionRemaining = reactionDuration

				outputSlot.AddOne(&itemCreated)
				inv.slotUpdate(outputSlot, furnaceSlotOutput)
				reagentSlot.Decrement()
				inv.slotUpdate(reagentSlot, furnaceSlotReagent)
			}
		}
	} else {
		inv.reactionRemaining = reactionDuration

		// Furnace is unlit.
		if haveReagent && haveFuel && outputReady {
			// Everything is in place, light the furnace by consuming one unit of
			// fuel.
			fuelSlot.Decrement()
			inv.maxFuel = fuelTicks
			inv.curFuel = fuelTicks
			inv.slotUpdate(fuelSlot, furnaceSlotFuel)
		} else {
			inv.reactionRemaining = reactionDuration
		}
	}
}

// sendProgressUpdates sends an update to the subscriber. Not every time,
// however - to cut down on unnecessary communication.
func (inv *FurnaceInventory) sendProgressUpdates() {
	inv.ticksSinceUpdate++
	if inv.ticksSinceUpdate > 5 || !inv.IsLit() {
		inv.ticksSinceUpdate = 0

		curFuelPrg := PrgBarValue(0)
		if inv.maxFuel != 0 {
			curFuelPrg = PrgBarValue((maxFuelPrg * inv.curFuel) / inv.maxFuel)
		}
		if inv.lastCurFuel != curFuelPrg {
			inv.lastCurFuel = curFuelPrg
			inv.subscriber.ProgressUpdate(PrgBarIdFurnaceFire, curFuelPrg)
		}

		curReactionRemaining := PrgBarValue(reactionDuration - inv.reactionRemaining)
		if inv.lastReactionRemaining != curReactionRemaining {
			inv.lastReactionRemaining = curReactionRemaining
			inv.subscriber.ProgressUpdate(PrgBarIdFurnaceProgress, curReactionRemaining)
		}
	}
}

func (inv *FurnaceInventory) IsLit() bool {
	return inv.curFuel > 0
}

// Tick runs the furnace for a single tick.
func (inv *FurnaceInventory) Tick() {
	if inv.curFuel > 0 {
		inv.curFuel--

		if inv.reactionRemaining > 0 {
			inv.reactionRemaining--
		}

		inv.stateCheck()

		inv.sendProgressUpdates()
	}
}
