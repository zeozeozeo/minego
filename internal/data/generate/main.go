package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	baseDir := filepath.Dir(os.Args[0])
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}

	// load JSON data
	registries := loadJSON[map[string]RegistryJSON](filepath.Join(baseDir, "registries.json"))
	blocks := loadJSON[map[string]BlockJSON](filepath.Join(baseDir, "blocks.json"))
	items := loadItems(baseDir)
	packets := loadJSON[PacketsJSON](filepath.Join(baseDir, "packets.json"))
	langPath := filepath.Join(baseDir, "en_us.json")

	outDir := filepath.Dir(baseDir)
	// decompiled sources live in the sibling go-mclib/mcsrc repo (../../../mcsrc relative to pkg/data)
	decompiledDir := filepath.Join(outDir, "..", "..", "..", "mcsrc", "current")
	if len(os.Args) > 2 {
		decompiledDir = os.Args[2]
	}
	itemTagsDir := filepath.Join(decompiledDir, "data", "minecraft", "tags", "item")
	datapackDir := filepath.Join(decompiledDir, "data", "minecraft")

	// game data dumped by the standalone mod (mcdump) — collision shapes, entity
	// hitboxes, hardness, mob categories: straight from the game, no scraping.
	mcdumpDir := filepath.Join(baseDir, "mcdump")
	mcdumpBlocks := filepath.Join(mcdumpDir, "blocks.json")
	mcdumpEntities := filepath.Join(mcdumpDir, "entities.json")

	// generate version info
	generateVersion(filepath.Join(outDir, "version_gen.go"))

	// generate packages
	generateRegistries(registries, datapackDir, filepath.Join(outDir, "registries", "registries_gen.go"))
	generateBlocks(registries, filepath.Join(outDir, "blocks", "blocks_gen.go"))
	generateBlockStates(blocks, registries, filepath.Join(outDir, "blocks", "block_states_gen.go"))
	generateItems(items, registries, filepath.Join(outDir, "items", "items_gen.go"))
	generateComponentTypes(registries, filepath.Join(outDir, "items", "item_components_gen.go"))
	generateComponentCodecs(registries, filepath.Join(baseDir, "component_metadata.include.json"), filepath.Join(outDir, "items", "item_components_codec_gen.go"))
	generatePacketIds(packets, filepath.Join(outDir, "packet_ids"))
	generateLang(langPath, filepath.Join(outDir, "lang", "lang_gen.go"))
	generateEntities(mcdumpEntities, filepath.Join(outDir, "entities", "entities_gen.go"))
	generateEntityMetadata(filepath.Join(baseDir, "entity_metadata.include.json"), filepath.Join(outDir, "entities"))
	generateBlockShapes(mcdumpBlocks, filepath.Join(outDir, "hitboxes", "blocks", "block_shapes_gen.go"))
	generateEntityHitboxes(mcdumpEntities, filepath.Join(outDir, "hitboxes", "entities", "entity_hitboxes_gen.go"))
	generateItemTags(itemTagsDir, registries, filepath.Join(outDir, "items", "tags_gen.go"))
	generateRegistryData(datapackDir, filepath.Join(outDir, "registries", "registry_data_gen.go"))
	generateTagData(filepath.Join(decompiledDir, "data", "minecraft", "tags"), filepath.Join(outDir, "registries", "tag_data_gen.go"))
	generateBlockHardness(mcdumpBlocks, filepath.Join(outDir, "blocks", "block_hardness_gen.go"))

	// fail loudly if any scraper produced implausibly few entries
	reportSanityChecks()

	fmt.Println("generation complete")
}
