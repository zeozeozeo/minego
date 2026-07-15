package minego

import (
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func TestPlayersDecodeIdentityAndLatency(t *testing.T) {
	p := newPlayers()
	buf := ns.NewWriter()
	uuid, _ := ns.UUIDFromString("00112233-4455-6677-8899-aabbccddeeff")
	_ = buf.WriteInt8(1 | 16)
	_ = buf.WriteVarInt(1)
	_ = buf.WriteUUID(uuid)
	_ = buf.WriteString("Steve")
	_ = buf.WriteVarInt(0)
	_ = buf.WriteVarInt(42)
	if err := p.decodeUpdate(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
	v, ok := p.ByName("steve")
	if !ok || v.UUID != uuid.String() || v.Latency != 42 {
		t.Fatalf("unexpected player: %#v %t", v, ok)
	}
	remove := ns.NewWriter()
	_ = remove.WriteVarInt(1)
	_ = remove.WriteUUID(uuid)
	if err := p.decodeRemove(remove.Bytes()); err != nil {
		t.Fatal(err)
	}
	if _, ok = p.Get(uuid.String()); ok {
		t.Fatal("player was not removed")
	}
}
