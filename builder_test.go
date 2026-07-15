package minego

import (
	"context"
	"errors"
	"testing"
)

func TestBuilderSelectsHotbarBlockAndBottomSupport(t *testing.T) {
	b := syntheticBot(t)
	b.Inventory.slots[38] = ItemStack{Name: "minecraft:cobblestone", ID: 1, Count: 12}
	item, hotbar, err := b.Builder.placementItem("cobblestone")
	if err != nil || hotbar != 2 || item.Name != "minecraft:cobblestone" {
		t.Fatalf("unexpected placement selection: item=%#v hotbar=%d err=%v", item, hotbar, err)
	}
	support, face, cursor, ok := b.Builder.supportFace(BlockPos{1, 64, 1})
	if !ok || support != (BlockPos{1, 63, 1}) || face != 1 || cursor.Y != 1 {
		t.Fatalf("unexpected support face: support=%+v face=%d cursor=%+v ok=%t", support, face, cursor, ok)
	}
}

func TestBuilderRejectsOccupiedPositionBeforeNetwork(t *testing.T) {
	b := syntheticBot(t)
	_, err := b.Builder.Place(context.Background(), BlockPos{1, 63, 1}, PlaceOptions{Item: "stone"})
	if err == nil || errors.Is(err, ErrNotConnected) {
		t.Fatalf("expected local replaceability error, got %v", err)
	}
}

func TestReplaceableBlockTag(t *testing.T) {
	for _, name := range []string{"minecraft:air", "minecraft:water", "minecraft:dead_bush"} {
		if !replaceableBlock(name) {
			t.Fatalf("%s should be replaceable", name)
		}
	}
	if replaceableBlock("minecraft:stone") {
		t.Fatal("stone must not be replaceable")
	}
}

func TestBuilderChoosesReachableLocalBlockBeforeInputOrder(t *testing.T) {
	b := syntheticBot(t)
	b.Self.update(func(s *SelfState) { s.Position = Vec3{1.5, 64, 1.5} })
	blocks := []BlueprintBlock{
		{Offset: BlockPos{X: 10}, Item: "minecraft:cobblestone"},
		{Offset: BlockPos{X: 1}, Item: "minecraft:cobblestone"},
	}
	if got := b.Builder.nextBuildBlock(BlockPos{0, 64, 1}, blocks, false); got != 1 {
		t.Fatalf("next build block = %d, want local supported block 1", got)
	}
}
