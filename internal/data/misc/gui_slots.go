package misc

// Player inventory slot indices for the survival inventory GUI (window ID 0).
// See https://minecraft.wiki/w/Java_Edition_protocol/Inventory#Player_Inventory
const (
	PlayerCraftingResult = 0
	PlayerCrafting1      = 1 // top-left
	PlayerCrafting2      = 2 // top-right
	PlayerCrafting3      = 3 // bottom-left
	PlayerCrafting4      = 4 // bottom-right
	PlayerArmorHead      = 5
	PlayerArmorChest     = 6
	PlayerArmorLegs      = 7
	PlayerArmorFeet      = 8
	PlayerInventory0     = 9 // top-left of main inventory
	PlayerInventory1     = 10
	PlayerInventory2     = 11
	PlayerInventory3     = 12
	PlayerInventory4     = 13
	PlayerInventory5     = 14
	PlayerInventory6     = 15
	PlayerInventory7     = 16
	PlayerInventory8     = 17
	PlayerInventory9     = 18
	PlayerInventory10    = 19
	PlayerInventory11    = 20
	PlayerInventory12    = 21
	PlayerInventory13    = 22
	PlayerInventory14    = 23
	PlayerInventory15    = 24
	PlayerInventory16    = 25
	PlayerInventory17    = 26
	PlayerInventory18    = 27
	PlayerInventory19    = 28
	PlayerInventory20    = 29
	PlayerInventory21    = 30
	PlayerInventory22    = 31
	PlayerInventory23    = 32
	PlayerInventory24    = 33
	PlayerInventory25    = 34
	PlayerInventory26    = 35
	PlayerHotbar0        = 36
	PlayerHotbar1        = 37
	PlayerHotbar2        = 38
	PlayerHotbar3        = 39
	PlayerHotbar4        = 40
	PlayerHotbar5        = 41
	PlayerHotbar6        = 42
	PlayerHotbar7        = 43
	PlayerHotbar8        = 44
	PlayerOffhand        = 45

	PlayerInvSize = 46
)

// PlayerPickupScanOrder is the slot scan order for item pickup:
// hotbar (36-44) left to right, then main inventory (9-35) left to right.
var PlayerPickupScanOrder = func() [36]int {
	var order [36]int
	// hotbar first
	for i := range 9 {
		order[i] = PlayerHotbar0 + i
	}
	// then main inventory
	for i := range 27 {
		order[9+i] = PlayerInventory0 + i
	}
	return order
}()
