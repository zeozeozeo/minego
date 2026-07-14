package versions

import (
	"github.com/zeozeozeo/minego/version"
	"testing"
)

func TestV262(t *testing.T) {
	if V26_2.Protocol() != 776 {
		t.Fatalf("protocol = %d", V26_2.Protocol())
	}
	id, ok := V26_2.StateID("minecraft:oak_log", map[string]string{"axis": "y"})
	if !ok {
		t.Fatal("oak log state missing")
	}
	b, ok := V26_2.BlockByState(id)
	if !ok || b.Name != "minecraft:oak_log" || b.Properties["axis"] != "y" {
		t.Fatalf("bad round trip: %#v", b)
	}
	if _, ok := V26_2.Packet(version.Login, version.Clientbound, 0); !ok {
		t.Fatal("login packet 0 missing")
	}
}
