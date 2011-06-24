package block

import (
	"chunkymonkey/inventory"
	. "chunkymonkey/types"
)

func makeChestAspect() (aspect IBlockAspect) {
	return &InventoryAspect{
		name:                 "Chest",
		createBlockInventory: createChestInventory,
	}
}

func createChestInventory(instance *BlockInstance) *blockInventory {
	inv := new(inventory.Inventory)
	inv.InitChestInventory()
	return newBlockInventory(
		instance,
		inv,
		false,
		InvTypeIdChest,
	)
}
