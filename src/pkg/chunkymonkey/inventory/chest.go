package inventory

const (
	chestInvWidth  = 9
	chestInvHeight = 3
)

// InitChestInventory initializes inv as a 9x3 chest.
func (inv *Inventory) InitChestInventory() {
	inv.Init(chestInvWidth * chestInvHeight)
}
