package versions

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

var packetRef = regexp.MustCompile(`packet_ids\.([A-Za-z0-9_]+)`)
var packetConst = regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_]+)\s*=\s*([0-9]+)\s*$`)

// TestV261RuntimePacketIDCompatibility protects the temporary adapter layer:
// while root handlers are being migrated to pack-owned codecs, every shared
// packet-id reference used by the runtime must mean the exact same wire ID in
// 26.1. A changed ID makes this test fail instead of silently selecting 26.2.
func TestV261RuntimePacketIDCompatibility(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate test source")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), ".."))
	current := readPacketConstants(t, filepath.Join(root, "internal", "data", "packet_ids"))
	older := readPacketConstants(t, filepath.Join(root, "internal", "data", "versions", "v26_1", "packet_ids"))
	for name, want := range current {
		if name == "C2SSpectatorActionID" { // renamed to C2SSpectateEntityID in 26.1
			continue
		}
		got, exists := older[name]
		if !exists || got != want {
			t.Fatalf("packet table differs for %s: 26.1=%d,%t 26.2=%d", name, got, exists, want)
		}
	}
	if got := older["C2SSpectateEntityID"]; got != current["C2SSpectatorActionID"] {
		t.Fatalf("renamed spectator packet ID = %d; want %d", got, current["C2SSpectatorActionID"])
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range packetRef.FindAllStringSubmatch(string(data), -1) {
			name := match[1]
			want, exists := current[name]
			if !exists {
				t.Fatalf("%s references non-generated packet ID %s", entry.Name(), name)
			}
			got, exists := older[name]
			if !exists {
				t.Fatalf("26.1 lacks runtime packet ID %s (referenced by %s)", name, entry.Name())
			}
			if got != want {
				t.Fatalf("26.1 packet %s = %d; runtime's 26.2 ID is %d", name, got, want)
			}
		}
	}
}

func readPacketConstants(t *testing.T, dir string) map[string]int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]int{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range packetConst.FindAllStringSubmatch(string(data), -1) {
			id, err := strconv.Atoi(match[2])
			if err != nil {
				t.Fatal(err)
			}
			result[match[1]] = id
		}
	}
	return result
}
