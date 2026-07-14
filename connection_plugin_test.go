package minego

import (
	"github.com/zeozeozeo/minego/internal/data/packets"
	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"testing"
)

func TestLoginPluginRequestCodec(t *testing.T) {
	p := &packets.S2CCustomQuery{MessageId: 7, Channel: "controlify:handshake", Data: []byte{0, 0, 0, 1}}
	wire, err := jp.ToWire(p)
	if err != nil {
		t.Fatal(err)
	}
	var got packets.S2CCustomQuery
	if err = wire.ReadInto(&got); err != nil {
		t.Fatal(err)
	}
	if got.MessageId != ns.VarInt(7) || got.Channel != "controlify:handshake" {
		t.Fatalf("decoded %#v", got)
	}
}
