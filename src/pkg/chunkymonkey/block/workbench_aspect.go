package block

import (
	"chunkymonkey/inventory"
	. "chunkymonkey/types"
)

func makeWorkbenchAspect() (aspect IBlockAspect) {
	return &InventoryAspect{
		name:                 "Workbench",
		createBlockInventory: createWorkbenchInventory,
	}
}

func createWorkbenchInventory(instance *BlockInstance) *blockInventory {
	inv := inventory.NewWorkbenchInventory(instance.Chunk.RecipeSet())
	return newBlockInventory(
		instance,
		inv,
		true,
		InvTypeIdWorkbench,
	)
}
