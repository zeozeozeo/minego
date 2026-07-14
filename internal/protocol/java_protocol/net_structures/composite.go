// These types handle common patterns like length-prefixed arrays,
// boolean-prefixed optionals, and bit sets.
//
// See https://minecraft.wiki/w/Java_Edition_protocol/Data_types
package net_structures

import (
	"fmt"
)

// ElementEncoder is a function that encodes an element to a buffer.
type ElementEncoder[T any] func(buf *PacketBuffer, v T) error

// ElementDecoder is a function that decodes an element from a buffer.
type ElementDecoder[T any] func(buf *PacketBuffer) (T, error)

// -----------------------------------------------------------------------------
// Prefixed Array
// -----------------------------------------------------------------------------

// PrefixedArray is a VarInt length-prefixed array of elements.
//
// Wire format:
//
//	┌───────────────────────┬───────────────────────────────┐
//	│  Length (VarInt)      │  Elements (T × Length)        │
//	└───────────────────────┴───────────────────────────────┘
//
// Example usage:
//
//	type MyPacket struct {
//	    Names PrefixedArray[String]
//	}
//
//	// in Read:
//	p.Names.DecodeWith(buf, func(b *PacketBuffer) (String, error) {
//	    return b.ReadString(32767)
//	})
//
//	// in Write:
//	p.Names.EncodeWith(buf, func(b *PacketBuffer, v String) error {
//	    return b.WriteString(v)
//	})
type PrefixedArray[T any] []T

// DecodeWith reads a length-prefixed array using the provided decoder function.
func (a *PrefixedArray[T]) DecodeWith(buf *PacketBuffer, decode ElementDecoder[T]) error {
	length, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read array length: %w", err)
	}
	if length < 0 {
		return fmt.Errorf("negative array length: %d", length)
	}

	*a = make([]T, length)
	for i := range *a {
		(*a)[i], err = decode(buf)
		if err != nil {
			return fmt.Errorf("failed to read array element %d: %w", i, err)
		}
	}
	return nil
}

// EncodeWith writes a length-prefixed array using the provided encoder function.
func (a PrefixedArray[T]) EncodeWith(buf *PacketBuffer, encode ElementEncoder[T]) error {
	if err := buf.WriteVarInt(VarInt(len(a))); err != nil {
		return fmt.Errorf("failed to write array length: %w", err)
	}
	for i, v := range a {
		if err := encode(buf, v); err != nil {
			return fmt.Errorf("failed to write array element %d: %w", i, err)
		}
	}
	return nil
}

// Len returns the number of elements in the array.
func (a PrefixedArray[T]) Len() int {
	return len(a)
}

// -----------------------------------------------------------------------------
// Prefixed Optional
// -----------------------------------------------------------------------------

// PrefixedOptional is a Boolean-prefixed optional value.
//
// Wire format:
//
//	┌──────────────────┬─────────────────────────────────────┐
//	│  Present (Bool)  │  Value (T, only if Present=true)    │
//	└──────────────────┴─────────────────────────────────────┘
//
// Example usage:
//
//	type MyPacket struct {
//	    Title PrefixedOptional[String]
//	}
//
//	// In Read:
//	p.Title.DecodeWith(buf, func(b *PacketBuffer) (String, error) {
//	    return b.ReadString(32767)
//	})
//
//	// In Write:
//	p.Title.EncodeWith(buf, func(b *PacketBuffer, v String) error {
//	    return b.WriteString(v)
//	})
type PrefixedOptional[T any] struct {
	Present bool
	Value   T
}

// Some creates a PrefixedOptional with a value.
func Some[T any](value T) PrefixedOptional[T] {
	return PrefixedOptional[T]{Present: true, Value: value}
}

// None creates an empty PrefixedOptional.
func None[T any]() PrefixedOptional[T] {
	return PrefixedOptional[T]{Present: false}
}

// DecodeWith reads a boolean-prefixed optional using the provided decoder.
func (o *PrefixedOptional[T]) DecodeWith(buf *PacketBuffer, decode ElementDecoder[T]) error {
	present, err := buf.ReadBool()
	if err != nil {
		return fmt.Errorf("failed to read optional presence: %w", err)
	}
	o.Present = bool(present)

	if o.Present {
		o.Value, err = decode(buf)
		if err != nil {
			return fmt.Errorf("failed to read optional value: %w", err)
		}
	}
	return nil
}

// EncodeWith writes a boolean-prefixed optional using the provided encoder.
func (o PrefixedOptional[T]) EncodeWith(buf *PacketBuffer, encode ElementEncoder[T]) error {
	if err := buf.WriteBool(Boolean(o.Present)); err != nil {
		return fmt.Errorf("failed to write optional presence: %w", err)
	}
	if o.Present {
		if err := encode(buf, o.Value); err != nil {
			return fmt.Errorf("failed to write optional value: %w", err)
		}
	}
	return nil
}

// Get returns the value and whether it's present.
func (o PrefixedOptional[T]) Get() (T, bool) {
	return o.Value, o.Present
}

// GetOrDefault returns the value if present, otherwise returns the default.
func (o PrefixedOptional[T]) GetOrDefault(defaultValue T) T {
	if o.Present {
		return o.Value
	}
	return defaultValue
}

// -----------------------------------------------------------------------------
// BitSet
// -----------------------------------------------------------------------------

// BitSet is a dynamically-sized bit set, prefixed by its length in longs.
//
// Wire format:
//
//	┌─────────────────┬─────────────────────────────────────┐
//	│  Length (VarInt)│  Longs (Int64 × Length)             │
//	└─────────────────┴─────────────────────────────────────┘
//
// The ith bit is set when (Data[i/64] & (1 << (i % 64))) != 0.
type BitSet struct {
	data []int64
}

// NewBitSet creates a BitSet with the given capacity in bits.
func NewBitSet(capacity int) *BitSet {
	numLongs := (capacity + 63) / 64
	return &BitSet{data: make([]int64, numLongs)}
}

// Decode reads a BitSet from the buffer.
func (b *BitSet) Decode(buf *PacketBuffer) error {
	length, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read bitset length: %w", err)
	}
	if length < 0 {
		return fmt.Errorf("negative bitset length: %d", length)
	}

	b.data = make([]int64, length)
	for i := range b.data {
		val, err := buf.ReadInt64()
		if err != nil {
			return fmt.Errorf("failed to read bitset long %d: %w", i, err)
		}
		b.data[i] = int64(val)
	}
	return nil
}

// Encode writes a BitSet to the buffer.
func (b *BitSet) Encode(buf *PacketBuffer) error {
	if err := buf.WriteVarInt(VarInt(len(b.data))); err != nil {
		return fmt.Errorf("failed to write bitset length: %w", err)
	}
	for i, v := range b.data {
		if err := buf.WriteInt64(Int64(v)); err != nil {
			return fmt.Errorf("failed to write bitset long %d: %w", i, err)
		}
	}
	return nil
}

// Get returns whether the bit at index i is set.
func (b *BitSet) Get(i int) bool {
	if i < 0 || i/64 >= len(b.data) {
		return false
	}
	return (b.data[i/64] & (1 << (i % 64))) != 0
}

// Set sets the bit at index i.
func (b *BitSet) Set(i int) {
	if i < 0 {
		return
	}
	idx := i / 64
	for len(b.data) <= idx {
		b.data = append(b.data, 0)
	}
	b.data[idx] |= 1 << (i % 64)
}

// Clear clears the bit at index i.
func (b *BitSet) Clear(i int) {
	if i < 0 || i/64 >= len(b.data) {
		return
	}
	b.data[i/64] &^= 1 << (i % 64)
}

// Longs returns the underlying long array.
func (b *BitSet) Longs() []int64 {
	return b.data
}

// -----------------------------------------------------------------------------
// Fixed BitSet
// -----------------------------------------------------------------------------

// FixedBitSet is a fixed-size bit set encoded as ceil(n/8) bytes.
//
// Wire format:
//
//	┌─────────────────────────────────────────────────────────┐
//	│  Bytes (ceil(n/8) bytes, no length prefix)              │
//	└─────────────────────────────────────────────────────────┘
//
// The ith bit is set when (Data[i/8] & (1 << (i % 8))) != 0.
type FixedBitSet struct {
	data []byte
	size int // number of bits
}

// NewFixedBitSet creates a FixedBitSet with the given size in bits.
func NewFixedBitSet(size int) *FixedBitSet {
	numBytes := (size + 7) / 8
	return &FixedBitSet{data: make([]byte, numBytes), size: size}
}

// FixedBitSetFromBytes creates a FixedBitSet from raw bytes.
func FixedBitSetFromBytes(data []byte, size int) *FixedBitSet {
	d := make([]byte, len(data))
	copy(d, data)
	return &FixedBitSet{data: d, size: size}
}

// Decode reads a FixedBitSet of the configured size from the buffer.
func (b *FixedBitSet) Decode(buf *PacketBuffer) error {
	numBytes := (b.size + 7) / 8
	data, err := buf.ReadFixedByteArray(numBytes)
	if err != nil {
		return fmt.Errorf("failed to read fixed bitset: %w", err)
	}
	b.data = data
	return nil
}

// Encode writes a FixedBitSet to the buffer.
func (b *FixedBitSet) Encode(buf *PacketBuffer) error {
	return buf.WriteFixedByteArray(b.data)
}

// Get returns whether the bit at index i is set.
func (b *FixedBitSet) Get(i int) bool {
	if i < 0 || i >= b.size {
		return false
	}
	return (b.data[i/8] & (1 << (i % 8))) != 0
}

// Set sets the bit at index i.
func (b *FixedBitSet) Set(i int) {
	if i < 0 || i >= b.size {
		return
	}
	b.data[i/8] |= 1 << (i % 8)
}

// Clear clears the bit at index i.
func (b *FixedBitSet) Clear(i int) {
	if i < 0 || i >= b.size {
		return
	}
	b.data[i/8] &^= 1 << (i % 8)
}

// Size returns the number of bits in the set.
func (b *FixedBitSet) Size() int {
	return b.size
}

// Bytes returns the underlying byte array.
func (b *FixedBitSet) Bytes() []byte {
	return b.data
}

// -----------------------------------------------------------------------------
// ID Set
// -----------------------------------------------------------------------------

// IDSet represents a registry ID set, which can be either a tag reference
// or an inline list of IDs.
//
// Wire format:
//
//	┌─────────────────┬─────────────────────────────────────┐
//	│  Type (VarInt)  │  Data (depends on Type)             │
//	└─────────────────┴─────────────────────────────────────┘
//
// If Type = 0: followed by an Identifier (tag name)
// If Type > 0: followed by (Type - 1) VarInt IDs
type IDSet struct {
	// IsTag indicates whether this is a tag reference (true) or inline IDs (false).
	IsTag bool
	// TagName is the tag identifier (only valid if IsTag is true).
	TagName Identifier
	// IDs is the list of registry IDs (only valid if IsTag is false).
	IDs []VarInt
}

// NewTagIDSet creates an IDSet that references a tag.
func NewTagIDSet(tagName Identifier) *IDSet {
	return &IDSet{IsTag: true, TagName: tagName}
}

// NewInlineIDSet creates an IDSet with inline IDs.
func NewInlineIDSet(ids []VarInt) *IDSet {
	return &IDSet{IsTag: false, IDs: ids}
}

// Decode reads an IDSet from the buffer.
func (s *IDSet) Decode(buf *PacketBuffer) error {
	typeVal, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read id set type: %w", err)
	}

	if typeVal == 0 {
		// tag reference
		s.IsTag = true
		s.TagName, err = buf.ReadIdentifier()
		if err != nil {
			return fmt.Errorf("failed to read id set tag name: %w", err)
		}
	} else {
		// inline IDs
		s.IsTag = false
		count := int(typeVal - 1)
		s.IDs = make([]VarInt, count)
		for i := range s.IDs {
			s.IDs[i], err = buf.ReadVarInt()
			if err != nil {
				return fmt.Errorf("failed to read id set id %d: %w", i, err)
			}
		}
	}
	return nil
}

// Encode writes an IDSet to the buffer.
func (s *IDSet) Encode(buf *PacketBuffer) error {
	if s.IsTag {
		if err := buf.WriteVarInt(0); err != nil {
			return fmt.Errorf("failed to write id set type: %w", err)
		}
		if err := buf.WriteIdentifier(s.TagName); err != nil {
			return fmt.Errorf("failed to write id set tag name: %w", err)
		}
	} else {
		if err := buf.WriteVarInt(VarInt(len(s.IDs) + 1)); err != nil {
			return fmt.Errorf("failed to write id set type: %w", err)
		}
		for i, id := range s.IDs {
			if err := buf.WriteVarInt(id); err != nil {
				return fmt.Errorf("failed to write id set id %d: %w", i, err)
			}
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// X or Y (Boolean-Selected Variant)
// -----------------------------------------------------------------------------

// XOrY represents a boolean-selected variant between two types.
// When IsX is true, the X value is encoded; otherwise, Y is encoded.
//
// Wire format:
//
//	┌──────────────────┬─────────────────────────────────────┐
//	│  IsX (Boolean)   │  X or Y (depending on IsX)          │
//	└──────────────────┴─────────────────────────────────────┘
//
// This pattern is relatively rare and is typically handled inline in packet
// definitions since both X and Y types must be known at compile time.
type XOrY[X, Y any] struct {
	IsX bool
	X   X
	Y   Y
}

// NewX creates an XOrY with an X value.
func NewX[X, Y any](value X) XOrY[X, Y] {
	return XOrY[X, Y]{IsX: true, X: value}
}

// NewY creates an XOrY with a Y value.
func NewY[X, Y any](value Y) XOrY[X, Y] {
	return XOrY[X, Y]{IsX: false, Y: value}
}

// DecodeWith reads an XOrY using the provided decoders.
func (v *XOrY[X, Y]) DecodeWith(buf *PacketBuffer, decodeX ElementDecoder[X], decodeY ElementDecoder[Y]) error {
	isX, err := buf.ReadBool()
	if err != nil {
		return fmt.Errorf("failed to read x-or-y selector: %w", err)
	}
	v.IsX = bool(isX)

	if v.IsX {
		v.X, err = decodeX(buf)
		if err != nil {
			return fmt.Errorf("failed to read x-or-y x value: %w", err)
		}
	} else {
		v.Y, err = decodeY(buf)
		if err != nil {
			return fmt.Errorf("failed to read x-or-y y value: %w", err)
		}
	}
	return nil
}

// EncodeWith writes an XOrY using the provided encoders.
func (v XOrY[X, Y]) EncodeWith(buf *PacketBuffer, encodeX ElementEncoder[X], encodeY ElementEncoder[Y]) error {
	if err := buf.WriteBool(Boolean(v.IsX)); err != nil {
		return fmt.Errorf("failed to write x-or-y selector: %w", err)
	}
	if v.IsX {
		if err := encodeX(buf, v.X); err != nil {
			return fmt.Errorf("failed to write x-or-y x value: %w", err)
		}
	} else {
		if err := encodeY(buf, v.Y); err != nil {
			return fmt.Errorf("failed to write x-or-y y value: %w", err)
		}
	}
	return nil
}

// Get returns either the X value (if IsX) or Y value, along with the selector.
func (v XOrY[X, Y]) Get() (x X, y Y, isX bool) {
	return v.X, v.Y, v.IsX
}

// IDOrX represents a registry ID or an inline value.
// Used when a field can reference a registry entry by ID or define a value inline.
//
// Wire format:
//
//	┌─────────────────┬─────────────────────────────────────┐
//	│  ID (VarInt)    │  Value (X, only if ID = 0)          │
//	└─────────────────┴─────────────────────────────────────┘
//
// If ID = 0, the inline value follows.
// If ID > 0, it represents registry ID + 1 (actual ID is ID - 1).
type IDOrX[T any] struct {
	// IsInline indicates whether this contains an inline value (true) or a registry ID (false).
	IsInline bool
	// ID is the registry ID (only valid if IsInline is false).
	// Note: wire format uses ID+1, but this field stores the actual ID.
	ID VarInt
	// Value is the inline value (only valid if IsInline is true).
	Value T
}

// NewIDRef creates an IDOrX that references a registry entry.
func NewIDRef[T any](id VarInt) IDOrX[T] {
	return IDOrX[T]{IsInline: false, ID: id}
}

// NewInlineValue creates an IDOrX with an inline value.
func NewInlineValue[T any](value T) IDOrX[T] {
	return IDOrX[T]{IsInline: true, Value: value}
}

// DecodeWith reads an IDOrX using the provided decoder for inline values.
func (x *IDOrX[T]) DecodeWith(buf *PacketBuffer, decode ElementDecoder[T]) error {
	id, err := buf.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read id-or-x id: %w", err)
	}

	if id == 0 {
		x.IsInline = true
		x.Value, err = decode(buf)
		if err != nil {
			return fmt.Errorf("failed to read id-or-x inline value: %w", err)
		}
	} else {
		x.IsInline = false
		x.ID = id - 1 // wire format is ID+1
	}
	return nil
}

// EncodeWith writes an IDOrX using the provided encoder for inline values.
func (x IDOrX[T]) EncodeWith(buf *PacketBuffer, encode ElementEncoder[T]) error {
	if x.IsInline {
		if err := buf.WriteVarInt(0); err != nil {
			return fmt.Errorf("failed to write id-or-x id: %w", err)
		}
		if err := encode(buf, x.Value); err != nil {
			return fmt.Errorf("failed to write id-or-x inline value: %w", err)
		}
	} else {
		if err := buf.WriteVarInt(x.ID + 1); err != nil {
			return fmt.Errorf("failed to write id-or-x id: %w", err)
		}
	}
	return nil
}

// Get returns the ID (if reference) or -1 (if inline), and the inline value if present.
func (x IDOrX[T]) Get() (id VarInt, value T, isInline bool) {
	if x.IsInline {
		return -1, x.Value, true
	}
	return x.ID, x.Value, false
}
