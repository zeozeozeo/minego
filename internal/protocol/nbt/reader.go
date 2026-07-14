package nbt

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// Reader decodes NBT data from binary format.
type Reader struct {
	r         io.Reader
	depth     int
	maxDepth  int
	bytesRead int64
	maxBytes  int64
}

// ReaderOption configures a Reader.
type ReaderOption func(*Reader)

// WithMaxDepth sets the maximum nesting depth.
func WithMaxDepth(depth int) ReaderOption {
	return func(r *Reader) {
		r.maxDepth = depth
	}
}

// WithMaxBytes sets the maximum bytes that can be read.
// Set to 0 for unlimited.
func WithMaxBytes(n int64) ReaderOption {
	return func(r *Reader) {
		r.maxBytes = n
	}
}

// NewReader creates a Reader from a byte slice.
func NewReader(data []byte, opts ...ReaderOption) *Reader {
	return NewReaderFrom(&byteReader{data: data}, opts...)
}

// NewReaderFrom creates a Reader from an io.Reader.
func NewReaderFrom(r io.Reader, opts ...ReaderOption) *Reader {
	reader := &Reader{
		r:        r,
		maxDepth: MaxDepth,
		maxBytes: MaxBytes,
	}
	for _, opt := range opts {
		opt(reader)
	}
	return reader
}

// byteReader wraps a byte slice as an io.Reader.
type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// ReadTag reads a complete NBT structure.
//
// If network is true, expects network format (no root name).
// If network is false, expects file format (with root name).
//
// Returns the tag, root name (empty for network format), and any error.
func (r *Reader) ReadTag(network bool) (Tag, string, error) {
	// Read tag type
	tagType, err := r.readByte()
	if err != nil {
		return nil, "", fmt.Errorf("failed to read tag type: %w", err)
	}

	if tagType == TagEnd {
		return End{}, "", nil
	}

	// Read root name (only for file format)
	var rootName string
	if !network {
		rootName, err = r.readString()
		if err != nil {
			return nil, "", fmt.Errorf("failed to read root name: %w", err)
		}
	}

	// Read tag payload
	tag, err := r.readTagPayload(tagType)
	if err != nil {
		return nil, "", err
	}

	return tag, rootName, nil
}

// readTagPayload reads the payload for a tag of the given type.
func (r *Reader) readTagPayload(tagType byte) (Tag, error) {
	switch tagType {
	case TagEnd:
		return End{}, nil

	case TagByte:
		v, err := r.readByte()
		return Byte(int8(v)), err

	case TagShort:
		v, err := r.readShort()
		return Short(v), err

	case TagInt:
		v, err := r.readInt()
		return Int(v), err

	case TagLong:
		v, err := r.readLong()
		return Long(v), err

	case TagFloat:
		v, err := r.readFloat()
		return Float(v), err

	case TagDouble:
		v, err := r.readDouble()
		return Double(v), err

	case TagByteArray:
		return r.readByteArray()

	case TagString:
		v, err := r.readString()
		return String(v), err

	case TagList:
		return r.readList()

	case TagCompound:
		return r.readCompound()

	case TagIntArray:
		return r.readIntArray()

	case TagLongArray:
		return r.readLongArray()

	default:
		return nil, fmt.Errorf("unknown tag type: %d", tagType)
	}
}

func (r *Reader) readByteArray() (ByteArray, error) {
	length, err := r.readInt()
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, fmt.Errorf("negative byte array length: %d", length)
	}

	data := make([]byte, length)
	if err := r.readFull(data); err != nil {
		return nil, err
	}
	return ByteArray(data), nil
}

func (r *Reader) readList() (List, error) {
	if err := r.pushDepth(); err != nil {
		return List{}, err
	}
	defer r.popDepth()

	elemType, err := r.readByte()
	if err != nil {
		return List{}, err
	}

	length, err := r.readInt()
	if err != nil {
		return List{}, err
	}
	if length < 0 {
		return List{}, fmt.Errorf("negative list length: %d", length)
	}

	elements := make([]Tag, length)
	for i := range length {
		elem, err := r.readTagPayload(elemType)
		if err != nil {
			return List{}, fmt.Errorf("failed to read list element %d: %w", i, err)
		}
		elements[i] = elem
	}

	return List{ElementType: elemType, Elements: elements}, nil
}

func (r *Reader) readCompound() (Compound, error) {
	if err := r.pushDepth(); err != nil {
		return nil, err
	}
	defer r.popDepth()

	compound := make(Compound)

	for {
		tagType, err := r.readByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read tag type in compound: %w", err)
		}

		if tagType == TagEnd {
			break
		}

		name, err := r.readString()
		if err != nil {
			return nil, fmt.Errorf("failed to read tag name: %w", err)
		}

		tag, err := r.readTagPayload(tagType)
		if err != nil {
			return nil, fmt.Errorf("failed to read tag %q: %w", name, err)
		}

		compound[name] = tag
	}

	return compound, nil
}

func (r *Reader) readIntArray() (IntArray, error) {
	length, err := r.readInt()
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, fmt.Errorf("negative int array length: %d", length)
	}

	data := make(IntArray, length)
	for i := range length {
		v, err := r.readInt()
		if err != nil {
			return nil, err
		}
		data[i] = v
	}
	return data, nil
}

func (r *Reader) readLongArray() (LongArray, error) {
	length, err := r.readInt()
	if err != nil {
		return nil, err
	}
	if length < 0 {
		return nil, fmt.Errorf("negative long array length: %d", length)
	}

	data := make(LongArray, length)
	for i := range length {
		v, err := r.readLong()
		if err != nil {
			return nil, err
		}
		data[i] = v
	}
	return data, nil
}

// --- Internal read methods ---

func (r *Reader) readFull(p []byte) error {
	if err := r.accountBytes(int64(len(p))); err != nil {
		return err
	}
	_, err := io.ReadFull(r.r, p)
	return err
}

func (r *Reader) readByte() (byte, error) {
	if err := r.accountBytes(1); err != nil {
		return 0, err
	}
	var buf [1]byte
	_, err := io.ReadFull(r.r, buf[:])
	return buf[0], err
}

func (r *Reader) readShort() (int16, error) {
	if err := r.accountBytes(2); err != nil {
		return 0, err
	}
	var buf [2]byte
	if _, err := io.ReadFull(r.r, buf[:]); err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(buf[:])), nil
}

func (r *Reader) readInt() (int32, error) {
	if err := r.accountBytes(4); err != nil {
		return 0, err
	}
	var buf [4]byte
	if _, err := io.ReadFull(r.r, buf[:]); err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf[:])), nil
}

func (r *Reader) readLong() (int64, error) {
	if err := r.accountBytes(8); err != nil {
		return 0, err
	}
	var buf [8]byte
	if _, err := io.ReadFull(r.r, buf[:]); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf[:])), nil
}

func (r *Reader) readFloat() (float32, error) {
	if err := r.accountBytes(4); err != nil {
		return 0, err
	}
	var buf [4]byte
	if _, err := io.ReadFull(r.r, buf[:]); err != nil {
		return 0, err
	}
	return math.Float32frombits(binary.BigEndian.Uint32(buf[:])), nil
}

func (r *Reader) readDouble() (float64, error) {
	if err := r.accountBytes(8); err != nil {
		return 0, err
	}
	var buf [8]byte
	if _, err := io.ReadFull(r.r, buf[:]); err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(buf[:])), nil
}

// readString reads a Java modified UTF-8 string.
// Format: 2-byte unsigned length prefix + UTF-8 bytes.
func (r *Reader) readString() (string, error) {
	if err := r.accountBytes(2); err != nil {
		return "", err
	}
	var buf [2]byte
	if _, err := io.ReadFull(r.r, buf[:]); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint16(buf[:])

	data := make([]byte, length)
	if err := r.readFull(data); err != nil {
		return "", err
	}

	return string(data), nil
}

// --- Depth and byte accounting ---

func (r *Reader) pushDepth() error {
	r.depth++
	if r.maxDepth > 0 && r.depth > r.maxDepth {
		return fmt.Errorf("NBT depth exceeds maximum of %d", r.maxDepth)
	}
	return nil
}

func (r *Reader) popDepth() {
	r.depth--
}

func (r *Reader) accountBytes(n int64) error {
	r.bytesRead += n
	if r.maxBytes > 0 && r.bytesRead > r.maxBytes {
		return errors.New("NBT data exceeds maximum byte limit")
	}
	return nil
}

// Decode reads NBT from a byte slice and returns the tag.
func Decode(data []byte, network bool, opts ...ReaderOption) (Tag, string, error) {
	r := NewReader(data, opts...)
	return r.ReadTag(network)
}

// DecodeNetwork reads NBT in network format (nameless root).
func DecodeNetwork(data []byte, opts ...ReaderOption) (Tag, error) {
	tag, _, err := Decode(data, true, opts...)
	return tag, err
}

// DecodeFile reads NBT in file format (with root name).
func DecodeFile(data []byte, opts ...ReaderOption) (Tag, string, error) {
	return Decode(data, false, opts...)
}
