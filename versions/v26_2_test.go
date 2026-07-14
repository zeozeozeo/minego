package versions

import (
	"github.com/zeozeozeo/minego/internal/data/packet_ids"
	"github.com/zeozeozeo/minego/version"
	"reflect"
	"strings"
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

func TestConnectionPacketIDsArePackOwned(t *testing.T) {
	tests := []struct {
		state version.State
		bound version.Bound
		name  version.PacketName
		id    int32
	}{
		{version.Handshake, version.Serverbound, version.PacketIntention, packet_ids.C2SIntentionID},
		{version.Status, version.Serverbound, version.PacketStatusRequest, packet_ids.C2SStatusRequestID},
		{version.Status, version.Clientbound, version.PacketStatusResponse, packet_ids.S2CStatusResponseID},
		{version.Login, version.Serverbound, version.PacketLoginHello, packet_ids.C2SHelloID},
		{version.Login, version.Clientbound, version.PacketLoginSuccess, packet_ids.S2CLoginFinishedID},
	}
	for _, test := range tests {
		for _, pack := range []version.Pack{V26_2, V26_1Data} {
			got, ok := pack.PacketID(test.state, test.bound, test.name)
			if !ok || got != test.id {
				t.Fatalf("%s %v %s: got %d, %t; want %d, true", pack.Name(), test.state, test.name, got, ok, test.id)
			}
		}
	}
	p, ok := V26_1Data.Packet(version.Login, version.Clientbound, packet_ids.S2CLoginFinishedID)
	if !ok {
		t.Fatal("26.1 login packet factory is missing")
	}
	if path := reflect.TypeOf(p).Elem().PkgPath(); !strings.Contains(path, "/internal/data/versions/v26_1/packets") {
		t.Fatalf("26.1 factory used a shared packet type: %s", path)
	}
}
