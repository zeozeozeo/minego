import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import net.minecraft.SharedConstants;
import net.minecraft.core.BlockPos;
import net.minecraft.core.Registry;
import net.minecraft.core.registries.BuiltInRegistries;
import net.minecraft.resources.Identifier;
import net.minecraft.server.Bootstrap;
import net.minecraft.world.entity.EntityDimensions;
import net.minecraft.world.entity.EntityType;
import net.minecraft.world.level.EmptyBlockGetter;
import net.minecraft.world.level.block.Block;
import net.minecraft.world.level.block.state.BlockState;
import net.minecraft.world.level.block.state.properties.Property;
import net.minecraft.world.phys.AABB;
import net.minecraft.world.phys.shapes.VoxelShape;

import java.io.Writer;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;

// Standalone dumper: boots Minecraft's registries straight from the server jar
// (real Mojang names — no Fabric, no remapping) and writes the game's own data
// as JSON. The actual game code is the single source of truth.
//
// Covers the data that previously needed PrismarineJS / source scraping
// (collision shapes, entity hitboxes, hardness, mob categories) plus the
// built-in registry id maps. Datapack-bound data (item component values, tags,
// biomes) needs a loaded world and still comes from the vanilla server datagen.
public final class Dump {
	private static final Gson GSON = new GsonBuilder().setPrettyPrinting().disableHtmlEscaping().create();

	public static void main(final String[] args) throws Exception {
		final Path out = Path.of(args.length > 0 ? args[0] : "go-mclib-dump");
		Files.createDirectories(out);

		SharedConstants.tryDetectVersion();
		Bootstrap.bootStrap();

		write(out, "blocks.json", blocks());
		write(out, "entities.json", entities());
		write(out, "registries.json", registries());

		System.out.println("[mcdump] dump written to " + out.toAbsolutePath());
	}

	// blocks: every state's palette id, properties, collision shape (deduped into
	// a shapes table), plus per-block hardness and requires-correct-tool.
	private static JsonObject blocks() {
		final JsonObject root = new JsonObject();
		final JsonObject blocks = new JsonObject();
		final List<JsonArray> shapeTable = new ArrayList<>();
		final Map<String, Integer> shapeIndex = new HashMap<>();

		for (final Block block : sortedByKey(BuiltInRegistries.BLOCK)) {
			final Identifier key = BuiltInRegistries.BLOCK.getKey(block);
			final BlockState def = block.defaultBlockState();

			final JsonObject bobj = new JsonObject();
			bobj.addProperty("protocolId", BuiltInRegistries.BLOCK.getId(block));
			bobj.addProperty("hardness", round(def.getDestroySpeed(EmptyBlockGetter.INSTANCE, BlockPos.ZERO)));
			bobj.addProperty("requiresTool", def.requiresCorrectToolForDrops());

			final JsonArray states = new JsonArray();
			for (final BlockState state : block.getStateDefinition().getPossibleStates()) {
				final JsonObject sobj = new JsonObject();
				sobj.addProperty("id", Block.getId(state)); // global block-state (palette) id
				final JsonObject props = new JsonObject();
				for (final Property<?> prop : state.getProperties()) {
					props.addProperty(prop.getName(), valueName(state, prop));
				}
				sobj.add("properties", props);
				final VoxelShape shape = state.getCollisionShape(EmptyBlockGetter.INSTANCE, BlockPos.ZERO);
				sobj.addProperty("shape", shapeIndexOf(shape, shapeTable, shapeIndex));
				states.add(sobj);
			}
			bobj.add("states", states);
			blocks.add(key.toString(), bobj);
		}

		root.add("blocks", blocks);
		final JsonArray shapes = new JsonArray();
		for (final JsonArray s : shapeTable) {
			shapes.add(s);
		}
		root.add("shapes", shapes);
		return root;
	}

	private static int shapeIndexOf(final VoxelShape shape, final List<JsonArray> table, final Map<String, Integer> index) {
		final JsonArray boxes = new JsonArray();
		final StringBuilder key = new StringBuilder();
		for (final AABB b : shape.toAabbs()) {
			final JsonArray box = new JsonArray();
			for (final double d : new double[]{b.minX, b.minY, b.minZ, b.maxX, b.maxY, b.maxZ}) {
				box.add(round(d));
			}
			boxes.add(box);
			key.append(box).append('|');
		}
		final Integer existing = index.get(key.toString());
		if (existing != null) {
			return existing;
		}
		final int id = table.size();
		table.add(boxes);
		index.put(key.toString(), id);
		return id;
	}

	@SuppressWarnings({"unchecked", "rawtypes"})
	private static String valueName(final BlockState state, final Property prop) {
		return prop.getName(state.getValue(prop));
	}

	// entities: hitbox width/height, eye height, mob category, client tracking range.
	private static JsonObject entities() {
		final JsonObject root = new JsonObject();
		for (final EntityType<?> type : sortedByKey(BuiltInRegistries.ENTITY_TYPE)) {
			final EntityDimensions dim = type.getDimensions();
			final JsonObject obj = new JsonObject();
			obj.addProperty("protocolId", BuiltInRegistries.ENTITY_TYPE.getId(type));
			obj.addProperty("width", round(dim.width()));
			obj.addProperty("height", round(dim.height()));
			obj.addProperty("eyeHeight", round(dim.eyeHeight()));
			obj.addProperty("category", type.getCategory().getName());
			obj.addProperty("trackingRange", type.clientTrackingRange());
			root.add(BuiltInRegistries.ENTITY_TYPE.getKey(type).toString(), obj);
		}
		return root;
	}

	// registries: every built-in registry -> { entry name: protocol id }. (Datapack
	// registries such as biomes need a loaded world and still come from the datagen.)
	private static JsonObject registries() {
		final JsonObject root = new JsonObject();
		final TreeMap<String, JsonObject> byName = new TreeMap<>();
		for (final Registry<?> reg : BuiltInRegistries.REGISTRY) {
			byName.put(reg.key().identifier().toString(), registryEntries(reg));
		}
		byName.forEach(root::add);
		return root;
	}

	@SuppressWarnings({"unchecked", "rawtypes"})
	private static JsonObject registryEntries(final Registry reg) {
		final TreeMap<Integer, String> byId = new TreeMap<>();
		for (final Object value : reg) {
			final Identifier key = reg.getKey(value);
			if (key != null) {
				byId.put(reg.getId(value), key.toString());
			}
		}
		final JsonObject entries = new JsonObject();
		byId.forEach((id, name) -> entries.addProperty(name, id));
		return entries;
	}

	// --- helpers ---

	private static <T> List<T> sortedByKey(final Registry<T> reg) {
		final List<T> list = new ArrayList<>();
		reg.forEach(list::add);
		list.sort(Comparator.comparing(v -> reg.getKey(v).toString()));
		return list;
	}

	private static double round(final double d) {
		return Math.round(d * 1.0e6) / 1.0e6;
	}

	private static void write(final Path dir, final String name, final JsonElement json) throws Exception {
		try (final Writer w = Files.newBufferedWriter(dir.resolve(name), StandardCharsets.UTF_8)) {
			GSON.toJson(json, w);
		}
	}
}
