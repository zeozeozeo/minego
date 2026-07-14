// go test -bench=. -benchmem
package nbt_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/zeozeozeo/minego/internal/protocol/nbt"
)

var (
	realFixtureNBT  []byte
	realFixtureGzip []byte
)

func init() {
	var err error
	realFixtureGzip, err = os.ReadFile(fixturePath)
	if err != nil {
		panic(err)
	}

	gr, err := gzip.NewReader(bytes.NewReader(realFixtureGzip))
	if err != nil {
		panic(err)
	}
	defer gr.Close()

	realFixtureNBT, err = io.ReadAll(gr)
	if err != nil {
		panic(err)
	}
}

func BenchmarkRealDecode(b *testing.B) {
	for b.Loop() {
		_, _, err := nbt.DecodeFile(realFixtureNBT)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealEncode(b *testing.B) {
	for b.Loop() {
		_, err := nbt.EncodeFile(expected, "")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealRoundTrip(b *testing.B) {
	for b.Loop() {
		decoded, rootName, err := nbt.DecodeFile(realFixtureNBT)
		if err != nil {
			b.Fatal(err)
		}
		_, err = nbt.EncodeFile(decoded, rootName)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealGzipDecode(b *testing.B) {
	for b.Loop() {
		gr, err := gzip.NewReader(bytes.NewReader(realFixtureGzip))
		if err != nil {
			b.Fatal(err)
		}
		data, err := io.ReadAll(gr)
		if err != nil {
			b.Fatal(err)
		}
		_ = gr.Close()

		_, _, err = nbt.DecodeFile(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealGzipEncode(b *testing.B) {
	for b.Loop() {
		data, err := nbt.EncodeFile(expected, "")
		if err != nil {
			b.Fatal(err)
		}

		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(data); err != nil {
			b.Fatal(err)
		}
		if err := gw.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
