package minego

import (
	"encoding/hex"
	"github.com/zeozeozeo/minego/internal/data/packet_ids"
	"github.com/zeozeozeo/minego/internal/data/packets"
	jp "github.com/zeozeozeo/minego/internal/protocol/java_protocol"
	"testing"
)

func TestCapturedFabricRegisterPayload(t *testing.T) {
	raw, err := hex.DecodeString("126d696e6563726166743a7265676973746572633a76657273696f6e00666f726765636f6e666967617069706f72743a636f6e6669675f66696c65006661627269633a7265636970655f73796e632f737570706f727465645f73657269616c697a657273006661627269633a61636365707465645f6174746163686d656e74735f763100633a7265676973746572006661627269633a637573746f6d5f696e6772656469656e745f73796e63006661627269633a72656769737472792f73796e632f636f6d706c657465")
	if err != nil {
		t.Fatal(err)
	}
	wire := &jp.WirePacket{PacketID: packet_ids.S2CCustomPayloadConfigurationID, Data: raw}
	var p packets.S2CCustomPayloadConfiguration
	if err = wire.ReadInto(&p); err != nil {
		t.Fatal(err)
	}
	if p.Channel != "minecraft:register" || len(p.Data) == 0 {
		t.Fatalf("decoded channel=%q bytes=%d", p.Channel, len(p.Data))
	}
	if _, err = jp.ToWire(&packets.C2SCustomPayloadConfiguration{Channel: p.Channel, Data: p.Data}); err != nil {
		t.Fatal(err)
	}
}
