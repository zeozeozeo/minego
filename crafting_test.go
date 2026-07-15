package minego

import "testing"

func TestVanillaWoodCraftingRecipes(t *testing.T) {
	b := syntheticBot(t)
	planks := b.Crafter.RecipesFor("oak_planks")
	if len(planks) == 0 || planks[0].Output.Count != 4 || planks[0].Width != 1 {
		t.Fatalf("unexpected plank recipes: %#v", planks)
	}
	doors := b.Crafter.RecipesFor("minecraft:warped_door")
	if len(doors) == 0 || doors[0].Width != 2 || doors[0].Height != 3 || doors[0].Output.Count != 3 {
		t.Fatalf("unexpected door recipes: %#v", doors)
	}
	if got := b.Crafter.RecipesFor("crafting_table"); len(got) < 10 || len(got[0].Ingredients) != 4 {
		t.Fatalf("unexpected table recipe: %#v", got)
	}
}

func TestWindowSnapshotIsDeepCopied(t *testing.T) {
	b := syntheticBot(t)
	b.Inventory.window = WindowSnapshot{ID: 2, StateID: 7, Slots: []ItemStack{{Name: "minecraft:stone", Count: 1, Components: map[int32][]byte{1: {2, 3}}}}}
	w := b.Inventory.Window()
	w.Slots[0].Count = 9
	w.Slots[0].Components[1][0] = 8
	again := b.Inventory.Window()
	if again.Slots[0].Count != 1 || again.Slots[0].Components[1][0] != 2 {
		t.Fatal("window snapshot aliases internal state")
	}
}

func TestRecipeRegistrationCopiesAndRejectsCycles(t *testing.T) {
	b := syntheticBot(t)
	r := Recipe{ID: "a", Output: ItemStack{Name: "test:a", Count: 1}, Width: 1, Height: 1, Ingredients: []RecipeIngredient{{Alternatives: []string{"test:b"}}}}
	b.Crafter.RegisterRecipe(r)
	r.Ingredients[0].Alternatives[0] = "test:changed"
	b.Crafter.RegisterRecipe(Recipe{ID: "b", Output: ItemStack{Name: "test:b", Count: 1}, Width: 1, Height: 1, Ingredients: []RecipeIngredient{{Alternatives: []string{"test:a"}}}})
	if got := b.Crafter.RecipesFor("test:a"); got[0].Ingredients[0].Alternatives[0] != "test:b" {
		t.Fatal("registered recipe aliases caller memory")
	}
	if err := b.Crafter.validateAcyclic("test:a", map[string]bool{}, map[string]bool{}); err == nil {
		t.Fatal("expected recipe cycle")
	}
}
