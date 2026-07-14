package nbt_test

import (
	"bytes"
	"testing"

	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

func TestEncodeDecodePrimitives(t *testing.T) {
	tests := []struct {
		name string
		tag  nbt.Tag
	}{
		{"byte", nbt.Byte(42)},
		{"byte negative", nbt.Byte(-1)},
		{"short", nbt.Short(12345)},
		{"short negative", nbt.Short(-12345)},
		{"int", nbt.Int(123456789)},
		{"int negative", nbt.Int(-123456789)},
		{"long", nbt.Long(9223372036854775807)},
		{"long negative", nbt.Long(-9223372036854775808)},
		{"float", nbt.Float(3.14159)},
		{"double", nbt.Double(3.141592653589793)},
		{"string", nbt.String("Hello, NBT!")},
		{"string unicode", nbt.String("日本語テスト")},
		{"byte array", nbt.ByteArray{1, 2, 3, 4, 5}},
		{"int array", nbt.IntArray{1, 2, 3, 4, 5}},
		{"long array", nbt.LongArray{1, 2, 3, 4, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name+" network", func(t *testing.T) {
			// wrap in compound for valid NBT
			compound := nbt.Compound{"value": tt.tag}

			data, err := nbt.EncodeNetwork(compound)
			if err != nil {
				t.Fatalf("EncodeNetwork() error = %v", err)
			}

			decoded, err := nbt.DecodeNetwork(data)
			if err != nil {
				t.Fatalf("DecodeNetwork() error = %v", err)
			}

			c, ok := decoded.(nbt.Compound)
			if !ok {
				t.Fatalf("expected Compound, got %T", decoded)
			}

			// compare string representation
			got := c["value"]
			if got.ID() != tt.tag.ID() {
				t.Errorf("tag type = %d, want %d", got.ID(), tt.tag.ID())
			}
		})

		t.Run(tt.name+" file", func(t *testing.T) {
			compound := nbt.Compound{"value": tt.tag}

			data, err := nbt.EncodeFile(compound, "test")
			if err != nil {
				t.Fatalf("EncodeFile() error = %v", err)
			}

			decoded, rootName, err := nbt.DecodeFile(data)
			if err != nil {
				t.Fatalf("DecodeFile() error = %v", err)
			}

			if rootName != "test" {
				t.Errorf("rootName = %q, want %q", rootName, "test")
			}

			c, ok := decoded.(nbt.Compound)
			if !ok {
				t.Fatalf("expected Compound, got %T", decoded)
			}

			got := c["value"]
			if got.ID() != tt.tag.ID() {
				t.Errorf("tag type = %d, want %d", got.ID(), tt.tag.ID())
			}
		})
	}
}

func TestEncodeDecodeCompound(t *testing.T) {
	original := nbt.Compound{
		"name":  nbt.String("Steve"),
		"x":     nbt.Double(100.5),
		"y":     nbt.Double(64.0),
		"z":     nbt.Double(-200.5),
		"level": nbt.Int(42),
		"items": nbt.List{
			ElementType: nbt.TagCompound,
			Elements: []nbt.Tag{
				nbt.Compound{
					"id":    nbt.String("minecraft:diamond"),
					"count": nbt.Byte(64),
				},
				nbt.Compound{
					"id":    nbt.String("minecraft:stick"),
					"count": nbt.Byte(32),
				},
			},
		},
	}

	data, err := nbt.EncodeNetwork(original)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := nbt.DecodeNetwork(data)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	c := decoded.(nbt.Compound)

	if c.GetString("name") != "Steve" {
		t.Errorf("name = %q, want %q", c.GetString("name"), "Steve")
	}
	if c.GetDouble("x") != 100.5 {
		t.Errorf("x = %v, want %v", c.GetDouble("x"), 100.5)
	}
	if c.GetInt("level") != 42 {
		t.Errorf("level = %v, want %v", c.GetInt("level"), 42)
	}

	items := c.GetList("items")
	if items.Len() != 2 {
		t.Errorf("items length = %d, want 2", items.Len())
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	type Item struct {
		ID    string `nbt:"id"`
		Count int8   `nbt:"count"`
	}

	type Player struct {
		Name  string  `nbt:"name"`
		X     float64 `nbt:"x"`
		Y     float64 `nbt:"y"`
		Z     float64 `nbt:"z"`
		Level int32   `nbt:"level"`
		Items []Item  `nbt:"items"`
	}

	original := Player{
		Name:  "Steve",
		X:     100.5,
		Y:     64.0,
		Z:     -200.5,
		Level: 42,
		Items: []Item{
			{ID: "minecraft:diamond", Count: 64},
			{ID: "minecraft:stick", Count: 32},
		},
	}

	data, err := nbt.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Player
	if err := nbt.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.X != original.X {
		t.Errorf("X = %v, want %v", decoded.X, original.X)
	}
	if decoded.Level != original.Level {
		t.Errorf("Level = %v, want %v", decoded.Level, original.Level)
	}
	if len(decoded.Items) != len(original.Items) {
		t.Errorf("Items length = %d, want %d", len(decoded.Items), len(original.Items))
	}
	if decoded.Items[0].ID != original.Items[0].ID {
		t.Errorf("Items[0].ID = %q, want %q", decoded.Items[0].ID, original.Items[0].ID)
	}
}

func TestMarshalOmitEmpty(t *testing.T) {
	type Config struct {
		Name    string `nbt:"name"`
		Debug   bool   `nbt:"debug,omitempty"`
		Timeout int32  `nbt:"timeout,omitempty"`
	}

	c := Config{Name: "test"}

	data, err := nbt.MarshalNetwork(c)
	if err != nil {
		t.Fatalf("MarshalNetwork() error = %v", err)
	}

	tag, err := nbt.DecodeNetwork(data)
	if err != nil {
		t.Fatalf("DecodeNetwork() error = %v", err)
	}

	compound := tag.(nbt.Compound)

	if _, ok := compound["debug"]; ok {
		t.Error("debug should be omitted")
	}
	if _, ok := compound["timeout"]; ok {
		t.Error("timeout should be omitted")
	}
	if _, ok := compound["name"]; !ok {
		t.Error("name should be present")
	}
}

func TestNetworkVsFileFormat(t *testing.T) {
	compound := nbt.Compound{"test": nbt.Int(42)}

	// network format: tag type (1) + payload
	networkData, _ := nbt.EncodeNetwork(compound)

	// file format: tag type (1) + name length (2) + name + payload
	fileData, _ := nbt.EncodeFile(compound, "root")

	// file format should be longer (has name field)
	if len(fileData) <= len(networkData) {
		t.Errorf("file format (%d bytes) should be longer than network format (%d bytes)",
			len(fileData), len(networkData))
	}

	// both should start with TagCompound (0x0A)
	if networkData[0] != nbt.TagCompound {
		t.Errorf("network format first byte = 0x%02X, want 0x%02X", networkData[0], nbt.TagCompound)
	}
	if fileData[0] != nbt.TagCompound {
		t.Errorf("file format first byte = 0x%02X, want 0x%02X", fileData[0], nbt.TagCompound)
	}

	// File format should have name "root" at bytes 1-6 (2 byte length + 4 chars)
	if fileData[1] != 0 || fileData[2] != 4 { // length = 4
		t.Errorf("file format name length = %d, want 4", int(fileData[1])<<8|int(fileData[2]))
	}
	if string(fileData[3:7]) != "root" {
		t.Errorf("file format name = %q, want %q", string(fileData[3:7]), "root")
	}
}

func TestDepthLimit(t *testing.T) {
	// create deeply nested structure
	var compound nbt.Tag = nbt.Compound{"end": nbt.Byte(1)}
	for range 600 {
		compound = nbt.Compound{"nested": compound}
	}

	data, err := nbt.EncodeNetwork(compound)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// should fail with default depth limit (512)
	_, err = nbt.DecodeNetwork(data)
	if err == nil {
		t.Error("DecodeNetwork() should fail with depth > 512")
	}

	// should succeed with higher limit
	_, err = nbt.DecodeNetwork(data, nbt.WithMaxDepth(700))
	if err != nil {
		t.Errorf("DecodeNetwork() with higher limit error = %v", err)
	}
}

func TestKnownBytes(t *testing.T) {
	// test against known NBT bytes
	// this is a simple compound with one byte value
	// network format: 0x0A (compound) + payload
	// payload: 0x01 (byte) + 0x00 0x04 "test" + 0x2A (42) + 0x00 (end)
	knownBytes := []byte{
		0x0A,       // TAG_Compound
		0x01,       // TAG_Byte
		0x00, 0x04, // name length = 4
		't', 'e', 's', 't', // name = "test"
		0x2A, // value = 42
		0x00, // TAG_End
	}

	tag, err := nbt.DecodeNetwork(knownBytes)
	if err != nil {
		t.Fatalf("DecodeNetwork() error = %v", err)
	}

	compound, ok := tag.(nbt.Compound)
	if !ok {
		t.Fatalf("expected Compound, got %T", tag)
	}

	if compound.GetByte("test") != 42 {
		t.Errorf("test = %d, want 42", compound.GetByte("test"))
	}

	// Re-encode and compare
	reencoded, err := nbt.EncodeNetwork(compound)
	if err != nil {
		t.Fatalf("EncodeNetwork() error = %v", err)
	}

	if !bytes.Equal(reencoded, knownBytes) {
		t.Errorf("re-encoded bytes = %v, want %v", reencoded, knownBytes)
	}
}

func TestEmptyCompound(t *testing.T) {
	compound := nbt.Compound{}

	data, err := nbt.EncodeNetwork(compound)
	if err != nil {
		t.Fatalf("EncodeNetwork() error = %v", err)
	}

	// Should be: 0x0A (compound) + 0x00 (end)
	expected := []byte{0x0A, 0x00}
	if !bytes.Equal(data, expected) {
		t.Errorf("empty compound = %v, want %v", data, expected)
	}

	decoded, err := nbt.DecodeNetwork(data)
	if err != nil {
		t.Fatalf("DecodeNetwork() error = %v", err)
	}

	if len(decoded.(nbt.Compound)) != 0 {
		t.Errorf("decoded compound length = %d, want 0", len(decoded.(nbt.Compound)))
	}
}

func TestEmptyList(t *testing.T) {
	list := nbt.List{ElementType: nbt.TagInt, Elements: nil}
	compound := nbt.Compound{"list": list}

	data, err := nbt.EncodeNetwork(compound)
	if err != nil {
		t.Fatalf("EncodeNetwork() error = %v", err)
	}

	decoded, err := nbt.DecodeNetwork(data)
	if err != nil {
		t.Fatalf("DecodeNetwork() error = %v", err)
	}

	decodedList := decoded.(nbt.Compound).GetList("list")
	if decodedList.Len() != 0 {
		t.Errorf("list length = %d, want 0", decodedList.Len())
	}
}
