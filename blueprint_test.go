package minego

import "testing"

func TestBlueprintMaterialsAndSite(t *testing.T) {
	b := syntheticBot(t)
	b.Self.update(func(s *SelfState) { s.Position = Vec3{1.5, 64, 1.5} })
	bp := Blueprint{Name: "tiny", Blocks: []BlueprintBlock{{Item: "oak_planks"}, {Offset: BlockPos{X: 1}, Item: "minecraft:oak_planks"}, {Offset: BlockPos{Y: 1}, Item: "oak_door"}}}
	m := b.Builder.Materials(bp)
	if m["minecraft:oak_planks"] != 2 || m["minecraft:oak_door"] != 1 {
		t.Fatalf("wrong materials: %#v", m)
	}
	site, err := b.Builder.FindSite(FindSiteOptions{Width: 2, Depth: 2, Height: 2, Radius: 2})
	if err != nil {
		t.Fatal(err)
	}
	if site.Y != 64 {
		t.Fatalf("unexpected site: %+v", site)
	}
}

func TestElytraFindsSafeLanding(t *testing.T) {
	b := syntheticBot(t)
	p, ok := b.Elytra.safeLanding(BlockPos{X: 4, Y: 64, Z: 4}, 2)
	if !ok || p.Y != 64 {
		t.Fatalf("unexpected landing: %+v %t", p, ok)
	}
}

func TestFindBuildSiteAllowsLevelingDiggableTerrain(t *testing.T) {
	b := syntheticBot(t)
	stone, _ := b.pack.StateID("minecraft:stone", nil)
	b.Self.update(func(s *SelfState) { s.Position = Vec3{1.5, 64, 1.5} })
	b.World.chunks[chunkKey{0, 0}].SetBlockState(2, 64, 2, stone)
	opt := FindSiteOptions{Width: 2, Depth: 2, Height: 2, Radius: 1, AllowClearing: true}
	site, err := b.Builder.FindSite(opt)
	if err != nil {
		t.Fatal(err)
	}
	if site.Y != 64 {
		t.Fatalf("clearable site was selected underground: %+v", site)
	}
}
