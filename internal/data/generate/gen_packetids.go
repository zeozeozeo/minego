package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// generatePacketIds generates packet ID constants organized by phase and bound
func generatePacketIds(packets PacketsJSON, outDir string) {
	// first, find which packet names need phase suffixes (appear in multiple phases with same bound)
	needsSuffix := findPacketsNeedingSuffix(packets)

	// phase name to file suffix mapping
	phaseSuffixes := map[string]string{
		"configuration": "Configuration",
		"handshake":     "Handshaking",
		"login":         "Login",
		"play":          "Play",
		"status":        "Status",
	}

	// bound to prefix mapping
	boundPrefixes := map[string]string{
		"serverbound": "C2S",
		"clientbound": "S2C",
	}

	// generate files for each phase and bound
	for phase, bounds := range packets {
		phaseSuffix := phaseSuffixes[phase]
		for bound, pktMap := range bounds {
			prefix := boundPrefixes[bound]
			filename := strings.ToLower(prefix) + "_" + strings.ToLower(phaseSuffix) + "_gen.go"
			if phase == "handshake" {
				filename = "c2s_handshaking_gen.go" // handshake only has serverbound
			}

			generatePacketFile(pktMap, prefix, phaseSuffix, needsSuffix, bound, filepath.Join(outDir, filename))
		}
	}
}

// findPacketsNeedingSuffix returns a map of packet_name+bound that appear in multiple phases
func findPacketsNeedingSuffix(packets PacketsJSON) map[string]bool {
	// count occurrences of each packet name + bound combination
	counts := make(map[string]int)
	for _, bounds := range packets {
		for bound, pktMap := range bounds {
			for name := range pktMap {
				key := name + ":" + bound
				counts[key]++
			}
		}
	}

	// return those that appear more than once
	needsSuffix := make(map[string]bool)
	for key, count := range counts {
		if count > 1 {
			needsSuffix[key] = true
		}
	}
	return needsSuffix
}

func generatePacketFile(pktMap map[string]PacketEntryJSON, prefix, phaseSuffix string, needsSuffix map[string]bool, bound, outPath string) {
	// sort packets by protocol_id for consistent ordering
	type packetInfo struct {
		name       string
		protocolID int32
	}
	var pkts []packetInfo
	for name, entry := range pktMap {
		pkts = append(pkts, packetInfo{name, entry.ProtocolID})
	}
	sort.Slice(pkts, func(i, j int) bool {
		return pkts[i].protocolID < pkts[j].protocolID
	})

	var sb strings.Builder
	sb.WriteString(generatedFileHeader("packet_ids"))
	sb.WriteString("// Packet IDs for " + strings.ToLower(phaseSuffix) + " phase (" + bound + ")\n")
	sb.WriteString("const (\n")

	for _, pkt := range pkts {
		constName := packetConstName(pkt.name, prefix, phaseSuffix, needsSuffix, bound)
		sb.WriteString(fmt.Sprintf("\t%s = %d\n", constName, pkt.protocolID))
	}

	sb.WriteString(")\n")

	writeFile(outPath, sb.String())
}

func packetConstName(name, prefix, phaseSuffix string, needsSuffix map[string]bool, bound string) string {
	// minecraft:packet_name -> PacketName
	name = strings.TrimPrefix(name, "minecraft:")
	goName := toGoName(name)

	// determine if suffix is needed
	key := "minecraft:" + name + ":" + bound
	suffix := ""
	if needsSuffix[key] {
		suffix = phaseSuffix
	}

	return prefix + goName + suffix + "ID"
}
