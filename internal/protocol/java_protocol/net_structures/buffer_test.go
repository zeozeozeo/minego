package net_structures_test

import (
	"bytes"
	"io"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func TestBufferReaderModes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}

	t.Run("NewReader from bytes", func(t *testing.T) {
		buf := ns.NewReader(data)
		b, err := buf.ReadByte()
		if err != nil {
			t.Fatalf("ReadByte() error = %v", err)
		}
		if b != 0x01 {
			t.Errorf("ReadByte() = %v, want %v", b, 0x01)
		}
	})
}

func TestBufferWriterModes(t *testing.T) {
	t.Run("NewWriter to internal buffer", func(t *testing.T) {
		buf := ns.NewWriter()
		if err := buf.WriteByte(0x42); err != nil {
			t.Fatalf("WriteByte() error = %v", err)
		}
		if !bytes.Equal(buf.Bytes(), []byte{0x42}) {
			t.Errorf("Bytes() = %v, want %v", buf.Bytes(), []byte{0x42})
		}
	})
}

func TestBufferReadExact(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	buf := ns.NewReader(data)

	// read exactly 2 bytes
	p := make([]byte, 2)
	n, err := buf.Read(p)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != 2 {
		t.Errorf("Read() n = %d, want 2", n)
	}
	if !bytes.Equal(p, []byte{0x01, 0x02}) {
		t.Errorf("Read() = %v, want %v", p, []byte{0x01, 0x02})
	}

	// read remaining
	p = make([]byte, 1)
	n, err = buf.Read(p)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if n != 1 || p[0] != 0x03 {
		t.Errorf("Read() = %v, want %v", p, []byte{0x03})
	}
}

func TestBufferReadEOF(t *testing.T) {
	data := []byte{0x01}
	buf := ns.NewReader(data)

	// read 1 byte, should succeed
	_, err := buf.ReadByte()
	if err != nil {
		t.Fatalf("ReadByte() error = %v", err)
	}

	// read another, should fail
	_, err = buf.ReadByte()
	if err != io.EOF {
		t.Errorf("ReadByte() error = %v, want io.EOF", err)
	}
}

func TestBufferByteArrayReadWrite(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	// write
	buf := ns.NewWriter()
	if err := buf.WriteByteArray(data); err != nil {
		t.Fatalf("WriteByteArray() error = %v", err)
	}

	// should be: VarInt(5) + data = 0x05 + data
	expected := append([]byte{0x05}, data...)
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("WriteByteArray() = %v, want %v", buf.Bytes(), expected)
	}

	// read
	reader := ns.NewReader(buf.Bytes())
	got, err := reader.ReadByteArray(0)
	if err != nil {
		t.Fatalf("ReadByteArray() error = %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("ReadByteArray() = %v, want %v", got, data)
	}
}

func TestBufferByteArrayMaxLen(t *testing.T) {
	// create byte array with 10 bytes
	data := make([]byte, 10)
	buf := ns.NewWriter()
	_ = buf.WriteByteArray(data)

	// try to read with max 5 bytes - should fail
	reader := ns.NewReader(buf.Bytes())
	_, err := reader.ReadByteArray(5)
	if err == nil {
		t.Error("ReadByteArray() should error when exceeding max length")
	}
}

func TestBufferFixedByteArray(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}

	// write
	buf := ns.NewWriter()
	if err := buf.WriteFixedByteArray(data); err != nil {
		t.Fatalf("WriteFixedByteArray() error = %v", err)
	}

	// should be raw bytes, no length prefix
	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("WriteFixedByteArray() = %v, want %v", buf.Bytes(), data)
	}

	// read
	reader := ns.NewReader(buf.Bytes())
	got, err := reader.ReadFixedByteArray(3)
	if err != nil {
		t.Fatalf("ReadFixedByteArray() error = %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("ReadFixedByteArray() = %v, want %v", got, data)
	}
}

func TestBufferReset(t *testing.T) {
	buf := ns.NewWriter()
	_ = buf.WriteByte(0x01)
	_ = buf.WriteByte(0x02)

	if buf.Len() != 2 {
		t.Errorf("Len() = %d, want 2", buf.Len())
	}

	buf.Reset()

	if buf.Len() != 0 {
		t.Errorf("Len() after Reset() = %d, want 0", buf.Len())
	}
	if len(buf.Bytes()) != 0 {
		t.Errorf("Bytes() after Reset() = %v, want empty", buf.Bytes())
	}
}

func TestBufferModeErrors(t *testing.T) {
	t.Run("write on reader", func(t *testing.T) {
		buf := ns.NewReader([]byte{0x01})
		_, err := buf.Write([]byte{0x02})
		if err == nil {
			t.Error("Write() on reader should error")
		}
	})

	t.Run("read on writer", func(t *testing.T) {
		buf := ns.NewWriter()
		_, err := buf.Read(make([]byte, 1))
		if err == nil {
			t.Error("Read() on writer should error")
		}
	})
}
