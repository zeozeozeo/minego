package nbt

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

// Writer encodes NBT data to binary format.
type Writer struct {
	w   io.Writer
	buf *bytes.Buffer // only set if we own the buffer
}

// NewWriter creates a Writer that writes to an internal buffer.
// Use Bytes() to retrieve the written data.
func NewWriter() *Writer {
	buf := &bytes.Buffer{}
	return &Writer{w: buf, buf: buf}
}

// NewWriterTo creates a Writer that writes to the given io.Writer.
func NewWriterTo(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Bytes returns the written bytes. Only valid if created with NewWriter.
func (w *Writer) Bytes() []byte {
	if w.buf != nil {
		return w.buf.Bytes()
	}
	return nil
}

// Reset resets the internal buffer. Only valid if created with NewWriter.
func (w *Writer) Reset() {
	if w.buf != nil {
		w.buf.Reset()
	}
}

// WriteTag writes a complete NBT structure with root tag.
//
// If network is true, writes in network format (no root name).
// If network is false, writes in file format (with root name, typically empty).
func (w *Writer) WriteTag(tag Tag, rootName string, network bool) error {
	// Write tag type
	if err := w.writeByte(tag.ID()); err != nil {
		return err
	}

	// Write root name (only for file format)
	if !network {
		if err := w.writeString(rootName); err != nil {
			return err
		}
	}

	// Write tag payload
	return tag.write(w)
}

// --- Internal write methods ---

func (w *Writer) writeByte(v byte) error {
	_, err := w.w.Write([]byte{v})
	return err
}

func (w *Writer) writeBytes(v []byte) error {
	_, err := w.w.Write(v)
	return err
}

func (w *Writer) writeShort(v int16) error {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(v))
	_, err := w.w.Write(buf[:])
	return err
}

func (w *Writer) writeInt(v int32) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(v))
	_, err := w.w.Write(buf[:])
	return err
}

func (w *Writer) writeLong(v int64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(v))
	_, err := w.w.Write(buf[:])
	return err
}

func (w *Writer) writeFloat(v float32) error {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(v))
	_, err := w.w.Write(buf[:])
	return err
}

func (w *Writer) writeDouble(v float64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(v))
	_, err := w.w.Write(buf[:])
	return err
}

// writeString writes a Java modified UTF-8 string.
// Format: 2-byte unsigned length prefix + UTF-8 bytes.
//
// Note: This is a simplified implementation that writes standard UTF-8.
// True Java modified UTF-8 handles null bytes and supplementary characters
// differently, but for most Minecraft data this is equivalent.
func (w *Writer) writeString(s string) error {
	data := []byte(s)
	if len(data) > 65535 {
		data = data[:65535]
	}

	// Write 2-byte length (unsigned short, big-endian)
	if err := w.writeShort(int16(len(data))); err != nil {
		return err
	}

	// Write UTF-8 bytes
	_, err := w.w.Write(data)
	return err
}

// Encode writes the given tag as a complete NBT structure.
// This is a convenience method that creates a new Writer and returns the bytes.
func Encode(tag Tag, rootName string, network bool) ([]byte, error) {
	w := NewWriter()
	if err := w.WriteTag(tag, rootName, network); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// EncodeNetwork writes the given tag in network format (nameless root).
func EncodeNetwork(tag Tag) ([]byte, error) {
	return Encode(tag, "", true)
}

// EncodeFile writes the given tag in file format (with root name).
func EncodeFile(tag Tag, rootName string) ([]byte, error) {
	return Encode(tag, rootName, false)
}

// Copy reads an NBT tag from src and writes it to dst.
// This is useful for passthrough scenarios where you don't need to inspect the data.
func Copy(dst io.Writer, src io.Reader, network bool) error {
	reader := NewReaderFrom(src)
	tag, rootName, err := reader.ReadTag(network)
	if err != nil {
		return err
	}

	writer := NewWriterTo(dst)
	return writer.WriteTag(tag, rootName, network)
}
