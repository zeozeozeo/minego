package nbt_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

var (
	fixturePath = filepath.Join("fixture", "real.dat")
	expected    = nbt.Compound{"Data": nbt.Compound{
		"test":        nbt.String("abc"),
		"DataVersion": nbt.Int(4671),
		"Difficulty":  nbt.Byte(2),
		"LastPlayed":  nbt.Long(1769167696260),
		"ServerBrands": nbt.List{
			ElementType: nbt.TagString,
			Elements: []nbt.Tag{
				nbt.String("fabric"),
			},
		},
		"Time": nbt.Long(56600),
		"Version": nbt.Compound{
			"Id":       nbt.Int(4671),
			"Name":     nbt.String("1.21.11"),
			"Series":   nbt.String("main"),
			"Snapshot": nbt.Byte(0),
		},
		"WanderingTraderSpawnChance": nbt.Int(50),
		"version":                    nbt.Int(19133),
		"TestFloat":                  nbt.Float(1.234567890),
	}}
)

func TestRealDecode(t *testing.T) {
	// read original
	compressed, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// decompress gzip (Minecraft .dat files are gzip compressed)
	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer gr.Close()

	// decode
	data, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	decoded, _, err := nbt.Decode(data, false)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// compare
	if !reflect.DeepEqual(decoded, expected) {
		t.Fatalf("decoded = %v, want %v", decoded, expected)
	}
}

func TestRealEncode(t *testing.T) {
	// read and decompress fixture
	compressed, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("gzip.NewReader() error = %v", err)
	}
	defer gr.Close()
	fixtureNBT, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}

	// encode our expected compound
	encoded, err := nbt.EncodeFile(expected, "")
	if err != nil {
		t.Fatalf("EncodeFile() error = %v", err)
	}

	// compare bytes directly
	if !bytes.Equal(encoded, fixtureNBT) {
		t.Fatalf("encoded bytes differ from fixture")
	}
}
