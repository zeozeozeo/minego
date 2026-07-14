package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Mojang publishes 1.21.11's server jar with obfuscated class names but also
// publishes a ProGuard mapping file. remapDumper applies the small, audited
// subset needed by Dump.java before javac sees it. Keeping this transform local
// to the dump program avoids shipping mappings or obfuscated names in MineGo.
func remapDumper(sourcePath, mappingsPath, outputPath string) error {
	if mappingsPath == "" {
		return fmt.Errorf("official server mappings are unavailable")
	}
	mappings, err := readMappings(mappingsPath)
	if err != nil {
		return err
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	s := string(source)

	// Resolve calls while their named owners are still present. The calls are
	// intentionally explicit: a mapping ambiguity must fail generation rather
	// than silently choosing an unrelated obfuscated member.
	replaceCall := func(owner, name string) error {
		member, err := mappings.member(owner, name)
		if err != nil {
			return err
		}
		s = strings.ReplaceAll(s, "."+name+"(", "."+member+"(")
		return nil
	}
	for _, call := range [][2]string{
		{"net.minecraft.SharedConstants", "tryDetectVersion"},
		{"net.minecraft.server.Bootstrap", "bootStrap"},
		{"net.minecraft.world.level.block.Block", "defaultBlockState"},
		{"net.minecraft.world.level.block.Block", "getId"},
		{"net.minecraft.core.Registry", "getKey"},
		{"net.minecraft.core.Registry", "getId"},
		{"net.minecraft.core.Registry", "key"},
		{"net.minecraft.world.level.block.state.BlockBehaviour$BlockStateBase", "getDestroySpeed"},
		{"net.minecraft.world.level.block.state.BlockBehaviour$BlockStateBase", "requiresCorrectToolForDrops"},
		{"net.minecraft.world.level.block.state.StateHolder", "getProperties"},
		{"net.minecraft.world.level.block.state.StateHolder", "getValue"},
		{"net.minecraft.world.level.block.state.BlockBehaviour$BlockStateBase", "getCollisionShape"},
		{"net.minecraft.world.level.block.state.StateDefinition", "getPossibleStates"},
		{"net.minecraft.world.level.block.state.properties.Property", "getName"},
		{"net.minecraft.world.phys.shapes.VoxelShape", "toAabbs"},
		{"net.minecraft.world.entity.EntityType", "getDimensions"},
		{"net.minecraft.world.entity.EntityType", "getCategory"},
		{"net.minecraft.world.entity.EntityType", "clientTrackingRange"},
		{"net.minecraft.world.entity.EntityDimensions", "width"},
		{"net.minecraft.world.entity.EntityDimensions", "height"},
		{"net.minecraft.world.entity.EntityDimensions", "eyeHeight"},
		{"net.minecraft.world.entity.MobCategory", "getName"},
		{"net.minecraft.resources.ResourceKey", "identifier"},
	} {
		if err := replaceCall(call[0], call[1]); err != nil {
			return err
		}
	}

	// Static registry fields and AABB coordinates are fields, not methods.
	for _, field := range [][2]string{
		{"net.minecraft.core.registries.BuiltInRegistries", "BLOCK"},
		{"net.minecraft.core.registries.BuiltInRegistries", "ENTITY_TYPE"},
		{"net.minecraft.core.registries.BuiltInRegistries", "REGISTRY"},
		{"net.minecraft.world.level.EmptyBlockGetter", "INSTANCE"},
		{"net.minecraft.core.BlockPos", "ZERO"},
		{"net.minecraft.world.phys.AABB", "minX"}, {"net.minecraft.world.phys.AABB", "minY"}, {"net.minecraft.world.phys.AABB", "minZ"},
		{"net.minecraft.world.phys.AABB", "maxX"}, {"net.minecraft.world.phys.AABB", "maxY"}, {"net.minecraft.world.phys.AABB", "maxZ"},
	} {
		obfuscated, err := mappings.member(field[0], field[1])
		if err != nil {
			return err
		}
		s = strings.ReplaceAll(s, "."+field[1], "."+obfuscated)
	}

	// Obfuscated Minecraft classes live in the default package. Remove named
	// imports, then replace each imported simple type with its mapped class.
	for _, named := range dumpClasses {
		obfuscated, err := mappings.class(named)
		if err != nil {
			return err
		}
		s = regexp.MustCompile(`(?m)^import `+regexp.QuoteMeta(named)+`;\r?\n`).ReplaceAllString(s, "")
		s = regexp.MustCompile(`\b`+regexp.QuoteMeta(simpleName(named))+`\b`).ReplaceAllString(s, obfuscated)
	}
	// 1.21.11 inherits several methods through differently mapped interfaces.
	// Resolve those call sites explicitly after the generic transform.
	s = strings.NewReplacer(
		"block.d()", "block.m()",
		"block.getStateDefinition().a()", "block.l().a()",
		"mi.e.j(block)", "mi.e.a(block)",
		"mi.g.j(type)", "mi.g.a(type)",
		"reg.j(value)", "reg.a(value)",
		"prop.f(state.c(prop))", "prop.b(state.c(prop))",
		"type.f().f()", "type.f().a()",
		"type.cq()", "type.o()",
	).Replace(s)
	return os.WriteFile(outputPath, []byte(s), 0o644)
}

var dumpClasses = []string{
	"net.minecraft.SharedConstants", "net.minecraft.core.BlockPos", "net.minecraft.core.Registry",
	"net.minecraft.core.registries.BuiltInRegistries", "net.minecraft.resources.Identifier", "net.minecraft.resources.ResourceKey",
	"net.minecraft.server.Bootstrap", "net.minecraft.world.entity.EntityDimensions", "net.minecraft.world.entity.EntityType", "net.minecraft.world.entity.MobCategory",
	"net.minecraft.world.level.EmptyBlockGetter", "net.minecraft.world.level.block.Block", "net.minecraft.world.level.block.state.BlockState",
	"net.minecraft.world.level.block.state.StateDefinition", "net.minecraft.world.level.block.state.properties.Property",
	"net.minecraft.world.phys.AABB", "net.minecraft.world.phys.shapes.VoxelShape",
}

type mappings struct {
	classes map[string]string
	members map[string]map[string]string
}

func readMappings(path string) (mappings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return mappings{}, err
	}
	m := mappings{classes: map[string]string{}, members: map[string]map[string]string{}}
	var current string
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, " ") && strings.Contains(line, " -> ") && strings.HasSuffix(strings.TrimSpace(line), ":") {
			parts := strings.Split(strings.TrimSuffix(strings.TrimSpace(line), ":"), " -> ")
			if len(parts) != 2 {
				continue
			}
			current = parts[0]
			m.classes[current] = parts[1]
			m.members[current] = map[string]string{}
			continue
		}
		if current == "" || !strings.HasPrefix(line, " ") || !strings.Contains(line, " -> ") {
			continue
		}
		parts := strings.Split(strings.TrimSpace(line), " -> ")
		if len(parts) != 2 {
			continue
		}
		left := parts[0]
		if colon := strings.LastIndex(left, ":"); colon >= 0 {
			left = left[colon+1:]
		}
		name := ""
		if open := strings.Index(left, "("); open >= 0 {
			prefix := strings.TrimSpace(left[:open])
			name = prefix[strings.LastIndex(prefix, " ")+1:]
		} else {
			name = strings.TrimSpace(left[strings.LastIndex(left, " ")+1:])
		}
		if old, exists := m.members[current][name]; exists && old != parts[1] {
			continue
		}
		m.members[current][name] = parts[1]
	}
	return m, nil
}
func (m mappings) class(name string) (string, error) {
	value, ok := m.classes[name]
	if !ok {
		return "", fmt.Errorf("mapping for class %s not found", name)
	}
	return value, nil
}
func (m mappings) member(owner, name string) (string, error) {
	value, ok := m.members[owner][name]
	if !ok {
		return "", fmt.Errorf("mapping for %s.%s not found", owner, name)
	}
	return value, nil
}
func simpleName(name string) string { return name[strings.LastIndex(name, ".")+1:] }
