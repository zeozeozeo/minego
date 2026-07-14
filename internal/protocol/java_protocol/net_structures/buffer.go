package net_structures

import (
	"bytes"
	"fmt"
	"io"
)

// PacketBuffer provides methods for reading and writing Minecraft protocol data types.
// It wraps io.Reader and io.Writer interfaces for streaming network communication.
type PacketBuffer struct {
	reader io.Reader
	writer io.Writer

	// For writer mode, we also keep a bytes.Buffer to retrieve written bytes
	buf *bytes.Buffer
}

// NewReader creates a PacketBuffer for reading from data.
func NewReader(data []byte) *PacketBuffer {
	return &PacketBuffer{
		reader: bytes.NewReader(data),
	}
}

// NewReaderFrom creates a PacketBuffer for reading from an io.Reader.
func NewReaderFrom(r io.Reader) *PacketBuffer {
	return &PacketBuffer{
		reader: r,
	}
}

// NewWriter creates a PacketBuffer for writing data.
func NewWriter() *PacketBuffer {
	buf := &bytes.Buffer{}
	return &PacketBuffer{
		writer: buf,
		buf:    buf,
	}
}

// NewWriterTo creates a PacketBuffer that writes directly to an io.Writer.
func NewWriterTo(w io.Writer) *PacketBuffer {
	return &PacketBuffer{
		writer: w,
	}
}

// Bytes returns the written bytes. Only valid for buffers created with NewWriter.
func (pb *PacketBuffer) Bytes() []byte {
	if pb.buf != nil {
		return pb.buf.Bytes()
	}
	return nil
}

// Len returns the number of written bytes. Only valid for buffers created with NewWriter.
func (pb *PacketBuffer) Len() int {
	if pb.buf != nil {
		return pb.buf.Len()
	}
	return 0
}

// Reset resets the buffer for reuse. Only valid for buffers created with NewWriter.
func (pb *PacketBuffer) Reset() {
	if pb.buf != nil {
		pb.buf.Reset()
	}
}

// --- Raw I/O ---

// Read reads exactly len(p) bytes from the buffer.
func (pb *PacketBuffer) Read(p []byte) (int, error) {
	if pb.reader == nil {
		return 0, fmt.Errorf("buffer not in read mode")
	}
	return io.ReadFull(pb.reader, p)
}

// ReadRemaining returns all unread bytes from a packet buffer.
func (pb *PacketBuffer) ReadRemaining() ([]byte, error) {
	if pb.reader == nil {
		return nil, fmt.Errorf("buffer not in read mode")
	}
	return io.ReadAll(pb.reader)
}

// Write writes p to the buffer.
func (pb *PacketBuffer) Write(p []byte) (int, error) {
	if pb.writer == nil {
		return 0, fmt.Errorf("buffer not in write mode")
	}
	return pb.writer.Write(p)
}

// ReadByte reads a single byte.
func (pb *PacketBuffer) ReadByte() (byte, error) {
	var b [1]byte
	_, err := pb.Read(b[:])
	return b[0], err
}

// WriteByte writes a single byte.
func (pb *PacketBuffer) WriteByte(b byte) error {
	_, err := pb.Write([]byte{b})
	return err
}

// Reader returns the underlying io.Reader.
func (pb *PacketBuffer) Reader() io.Reader {
	return pb.reader
}

// Writer returns the underlying io.Writer.
func (pb *PacketBuffer) Writer() io.Writer {
	return pb.writer
}

// --- VarInt ---

// ReadVarInt reads a variable-length 32-bit integer.
func (pb *PacketBuffer) ReadVarInt() (VarInt, error) {
	return DecodeVarInt(pb.reader)
}

// WriteVarInt writes a variable-length 32-bit integer.
func (pb *PacketBuffer) WriteVarInt(v VarInt) error {
	return v.Encode(pb.writer)
}

// --- VarLong ---

// ReadVarLong reads a variable-length 64-bit integer.
func (pb *PacketBuffer) ReadVarLong() (VarLong, error) {
	return DecodeVarLong(pb.reader)
}

// WriteVarLong writes a variable-length 64-bit integer.
func (pb *PacketBuffer) WriteVarLong(v VarLong) error {
	return v.Encode(pb.writer)
}

// --- Fixed-width integers (Big Endian) ---

// ReadBool reads a boolean (1 byte: 0x00 = false, 0x01 = true).
func (pb *PacketBuffer) ReadBool() (Boolean, error) {
	return DecodeBoolean(pb.reader)
}

// WriteBool writes a boolean.
func (pb *PacketBuffer) WriteBool(v Boolean) error {
	return v.Encode(pb.writer)
}

// ReadInt8 reads a signed 8-bit integer.
func (pb *PacketBuffer) ReadInt8() (Int8, error) {
	return DecodeInt8(pb.reader)
}

// WriteInt8 writes a signed 8-bit integer.
func (pb *PacketBuffer) WriteInt8(v Int8) error {
	return v.Encode(pb.writer)
}

// ReadUint8 reads an unsigned 8-bit integer.
func (pb *PacketBuffer) ReadUint8() (Uint8, error) {
	return DecodeUint8(pb.reader)
}

// WriteUint8 writes an unsigned 8-bit integer.
func (pb *PacketBuffer) WriteUint8(v Uint8) error {
	return v.Encode(pb.writer)
}

// ReadInt16 reads a big-endian signed 16-bit integer.
func (pb *PacketBuffer) ReadInt16() (Int16, error) {
	return DecodeInt16(pb.reader)
}

// WriteInt16 writes a big-endian signed 16-bit integer.
func (pb *PacketBuffer) WriteInt16(v Int16) error {
	return v.Encode(pb.writer)
}

// ReadUint16 reads a big-endian unsigned 16-bit integer.
func (pb *PacketBuffer) ReadUint16() (Uint16, error) {
	return DecodeUint16(pb.reader)
}

// WriteUint16 writes a big-endian unsigned 16-bit integer.
func (pb *PacketBuffer) WriteUint16(v Uint16) error {
	return v.Encode(pb.writer)
}

// ReadInt32 reads a big-endian signed 32-bit integer.
func (pb *PacketBuffer) ReadInt32() (Int32, error) {
	return DecodeInt32(pb.reader)
}

// WriteInt32 writes a big-endian signed 32-bit integer.
func (pb *PacketBuffer) WriteInt32(v Int32) error {
	return v.Encode(pb.writer)
}

// ReadInt64 reads a big-endian signed 64-bit integer.
func (pb *PacketBuffer) ReadInt64() (Int64, error) {
	return DecodeInt64(pb.reader)
}

// WriteInt64 writes a big-endian signed 64-bit integer.
func (pb *PacketBuffer) WriteInt64(v Int64) error {
	return v.Encode(pb.writer)
}

// --- Floating point (Big Endian IEEE 754) ---

// ReadFloat32 reads a big-endian 32-bit IEEE 754 float.
func (pb *PacketBuffer) ReadFloat32() (Float32, error) {
	return DecodeFloat32(pb.reader)
}

// WriteFloat32 writes a big-endian 32-bit IEEE 754 float.
func (pb *PacketBuffer) WriteFloat32(v Float32) error {
	return v.Encode(pb.writer)
}

// ReadFloat64 reads a big-endian 64-bit IEEE 754 double.
func (pb *PacketBuffer) ReadFloat64() (Float64, error) {
	return DecodeFloat64(pb.reader)
}

// WriteFloat64 writes a big-endian 64-bit IEEE 754 double.
func (pb *PacketBuffer) WriteFloat64(v Float64) error {
	return v.Encode(pb.writer)
}

// --- String ---

// ReadString reads a UTF-8 string with VarInt length prefix.
// maxLen is the maximum allowed string length in characters (usually 32767, 0 means no limit).
func (pb *PacketBuffer) ReadString(maxLen int) (String, error) {
	return DecodeString(pb.reader, maxLen)
}

// WriteString writes a UTF-8 string with VarInt length prefix.
func (pb *PacketBuffer) WriteString(v String) error {
	return v.Encode(pb.writer)
}

// --- Identifier ---

// ReadIdentifier reads a Minecraft identifier (namespaced ID).
func (pb *PacketBuffer) ReadIdentifier() (Identifier, error) {
	return DecodeIdentifier(pb.reader)
}

// WriteIdentifier writes a Minecraft identifier.
func (pb *PacketBuffer) WriteIdentifier(v Identifier) error {
	return v.Encode(pb.writer)
}

// --- Byte Array ---

// ReadByteArray reads a byte array with VarInt length prefix.
func (pb *PacketBuffer) ReadByteArray(maxLen int) (ByteArray, error) {
	length, err := pb.ReadVarInt()
	if err != nil {
		return nil, fmt.Errorf("failed to read byte array length: %w", err)
	}

	if length < 0 {
		return nil, fmt.Errorf("negative byte array length: %d", length)
	}

	if maxLen > 0 && int(length) > maxLen {
		return nil, fmt.Errorf("byte array length %d exceeds maximum %d", length, maxLen)
	}

	data := make([]byte, length)
	if _, err := pb.Read(data); err != nil {
		return nil, fmt.Errorf("failed to read byte array data: %w", err)
	}

	return data, nil
}

// WriteByteArray writes a byte array with VarInt length prefix.
func (pb *PacketBuffer) WriteByteArray(v ByteArray) error {
	if err := pb.WriteVarInt(VarInt(len(v))); err != nil {
		return fmt.Errorf("failed to write byte array length: %w", err)
	}
	if _, err := pb.Write(v); err != nil {
		return fmt.Errorf("failed to write byte array data: %w", err)
	}
	return nil
}

// ReadFixedByteArray reads exactly n bytes.
func (pb *PacketBuffer) ReadFixedByteArray(n int) (ByteArray, error) {
	data := make([]byte, n)
	if _, err := pb.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

// WriteFixedByteArray writes bytes without length prefix.
func (pb *PacketBuffer) WriteFixedByteArray(v ByteArray) error {
	_, err := pb.Write(v)
	return err
}

// --- Position (BlockPos) ---

// ReadPosition reads a block position packed into a 64-bit integer.
func (pb *PacketBuffer) ReadPosition() (Position, error) {
	return DecodePosition(pb.reader)
}

// WritePosition writes a block position packed into a 64-bit integer.
func (pb *PacketBuffer) WritePosition(v Position) error {
	return v.Encode(pb.writer)
}

// --- GlobalPos ---

// ReadGlobalPos reads a global position (dimension + block position).
func (pb *PacketBuffer) ReadGlobalPos() (GlobalPos, error) {
	return DecodeGlobalPos(pb.reader)
}

// WriteGlobalPos writes a global position (dimension + block position).
func (pb *PacketBuffer) WriteGlobalPos(v GlobalPos) error {
	return v.Encode(pb.writer)
}

// --- UUID ---

// ReadUUID reads a 128-bit UUID (two 64-bit integers).
func (pb *PacketBuffer) ReadUUID() (UUID, error) {
	return DecodeUUID(pb.reader)
}

// WriteUUID writes a 128-bit UUID.
func (pb *PacketBuffer) WriteUUID(v UUID) error {
	return v.Encode(pb.writer)
}

// --- Angle ---

// ReadAngle reads a rotation angle (1 byte = 1/256 of a full turn).
func (pb *PacketBuffer) ReadAngle() (Angle, error) {
	return DecodeAngle(pb.reader)
}

// WriteAngle writes a rotation angle.
func (pb *PacketBuffer) WriteAngle(v Angle) error {
	return v.Encode(pb.writer)
}

// --- Copy methods for primitives (read from source, write to this buffer) ---

// CopyVarInt copies a VarInt from src to this buffer.
func (pb *PacketBuffer) CopyVarInt(src *PacketBuffer) error {
	v, err := src.ReadVarInt()
	if err != nil {
		return err
	}
	return pb.WriteVarInt(v)
}

// CopyVarLong copies a VarLong from src to this buffer.
func (pb *PacketBuffer) CopyVarLong(src *PacketBuffer) error {
	v, err := src.ReadVarLong()
	if err != nil {
		return err
	}
	return pb.WriteVarLong(v)
}

// CopyBool copies a Boolean from src to this buffer.
func (pb *PacketBuffer) CopyBool(src *PacketBuffer) error {
	v, err := src.ReadBool()
	if err != nil {
		return err
	}
	return pb.WriteBool(v)
}

// CopyInt8 copies an Int8 from src to this buffer.
func (pb *PacketBuffer) CopyInt8(src *PacketBuffer) error {
	v, err := src.ReadInt8()
	if err != nil {
		return err
	}
	return pb.WriteInt8(v)
}

// CopyInt16 copies an Int16 from src to this buffer.
func (pb *PacketBuffer) CopyInt16(src *PacketBuffer) error {
	v, err := src.ReadInt16()
	if err != nil {
		return err
	}
	return pb.WriteInt16(v)
}

// CopyInt32 copies an Int32 from src to this buffer.
func (pb *PacketBuffer) CopyInt32(src *PacketBuffer) error {
	v, err := src.ReadInt32()
	if err != nil {
		return err
	}
	return pb.WriteInt32(v)
}

// CopyInt64 copies an Int64 from src to this buffer.
func (pb *PacketBuffer) CopyInt64(src *PacketBuffer) error {
	v, err := src.ReadInt64()
	if err != nil {
		return err
	}
	return pb.WriteInt64(v)
}

// CopyFloat32 copies a Float32 from src to this buffer.
func (pb *PacketBuffer) CopyFloat32(src *PacketBuffer) error {
	v, err := src.ReadFloat32()
	if err != nil {
		return err
	}
	return pb.WriteFloat32(v)
}

// CopyFloat64 copies a Float64 from src to this buffer.
func (pb *PacketBuffer) CopyFloat64(src *PacketBuffer) error {
	v, err := src.ReadFloat64()
	if err != nil {
		return err
	}
	return pb.WriteFloat64(v)
}

// CopyString copies a String from src to this buffer.
func (pb *PacketBuffer) CopyString(src *PacketBuffer, maxLen int) error {
	v, err := src.ReadString(maxLen)
	if err != nil {
		return err
	}
	return pb.WriteString(v)
}

// CopyUUID copies a UUID from src to this buffer.
func (pb *PacketBuffer) CopyUUID(src *PacketBuffer) error {
	v, err := src.ReadUUID()
	if err != nil {
		return err
	}
	return pb.WriteUUID(v)
}

// CopyPosition copies a Position from src to this buffer.
func (pb *PacketBuffer) CopyPosition(src *PacketBuffer) error {
	v, err := src.ReadPosition()
	if err != nil {
		return err
	}
	return pb.WritePosition(v)
}
