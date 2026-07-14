package nbt

import (
	"fmt"
	"sort"
)

// Byte represents a TAG_Byte (signed 8-bit integer).
type Byte int8

func (Byte) ID() byte { return TagByte }
func (b Byte) write(w *Writer) error {
	return w.writeByte(byte(b))
}

// Short represents a TAG_Short (signed 16-bit integer, big-endian).
type Short int16

func (Short) ID() byte { return TagShort }
func (s Short) write(w *Writer) error {
	return w.writeShort(int16(s))
}

// Int represents a TAG_Int (signed 32-bit integer, big-endian).
type Int int32

func (Int) ID() byte { return TagInt }
func (i Int) write(w *Writer) error {
	return w.writeInt(int32(i))
}

// Long represents a TAG_Long (signed 64-bit integer, big-endian).
type Long int64

func (Long) ID() byte { return TagLong }
func (l Long) write(w *Writer) error {
	return w.writeLong(int64(l))
}

// Float represents a TAG_Float (32-bit IEEE 754 float, big-endian).
type Float float32

func (Float) ID() byte { return TagFloat }
func (f Float) write(w *Writer) error {
	return w.writeFloat(float32(f))
}

// Double represents a TAG_Double (64-bit IEEE 754 double, big-endian).
type Double float64

func (Double) ID() byte { return TagDouble }
func (d Double) write(w *Writer) error {
	return w.writeDouble(float64(d))
}

// ByteArray represents a TAG_Byte_Array.
type ByteArray []byte

func (ByteArray) ID() byte { return TagByteArray }
func (b ByteArray) write(w *Writer) error {
	if err := w.writeInt(int32(len(b))); err != nil {
		return err
	}
	return w.writeBytes(b)
}

// String represents a TAG_String (modified UTF-8 with 2-byte length prefix).
type String string

func (String) ID() byte { return TagString }
func (s String) write(w *Writer) error {
	return w.writeString(string(s))
}

// List represents a TAG_List (homogeneous list of unnamed tags).
type List struct {
	ElementType byte
	Elements    []Tag
}

func (List) ID() byte { return TagList }
func (l List) write(w *Writer) error {
	elemType := l.ElementType
	if len(l.Elements) == 0 {
		elemType = TagEnd
	}

	if err := w.writeByte(elemType); err != nil {
		return err
	}
	if err := w.writeInt(int32(len(l.Elements))); err != nil {
		return err
	}

	for i, elem := range l.Elements {
		if elem.ID() != elemType {
			return fmt.Errorf("list element %d has type %s, expected %s",
				i, TagName(elem.ID()), TagName(elemType))
		}
		if err := elem.write(w); err != nil {
			return err
		}
	}

	return nil
}

// Len returns the number of elements in the list.
func (l List) Len() int {
	return len(l.Elements)
}

// Get returns the element at index i, or nil if out of bounds.
func (l List) Get(i int) Tag {
	if i < 0 || i >= len(l.Elements) {
		return nil
	}
	return l.Elements[i]
}

// Compound represents a TAG_Compound (map of named tags).
type Compound map[string]Tag

func (Compound) ID() byte { return TagCompound }
func (c Compound) write(w *Writer) error {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		tag := c[name]
		// Write tag type
		if err := w.writeByte(tag.ID()); err != nil {
			return err
		}
		// Write tag name
		if err := w.writeString(name); err != nil {
			return err
		}
		// Write tag payload
		if err := tag.write(w); err != nil {
			return err
		}
	}

	// Write end tag
	return w.writeByte(TagEnd)
}

// Get returns the tag with the given name, or nil if not found.
func (c Compound) Get(name string) Tag {
	return c[name]
}

// GetByte returns the byte value for the given name, or 0 if not found or wrong type.
func (c Compound) GetByte(name string) int8 {
	if v, ok := c[name].(Byte); ok {
		return int8(v)
	}
	return 0
}

// GetShort returns the short value for the given name, or 0 if not found or wrong type.
func (c Compound) GetShort(name string) int16 {
	if v, ok := c[name].(Short); ok {
		return int16(v)
	}
	return 0
}

// GetInt returns the int value for the given name, or 0 if not found or wrong type.
func (c Compound) GetInt(name string) int32 {
	if v, ok := c[name].(Int); ok {
		return int32(v)
	}
	return 0
}

// GetLong returns the long value for the given name, or 0 if not found or wrong type.
func (c Compound) GetLong(name string) int64 {
	if v, ok := c[name].(Long); ok {
		return int64(v)
	}
	return 0
}

// GetFloat returns the float value for the given name, or 0 if not found or wrong type.
func (c Compound) GetFloat(name string) float32 {
	if v, ok := c[name].(Float); ok {
		return float32(v)
	}
	return 0
}

// GetDouble returns the double value for the given name, or 0 if not found or wrong type.
func (c Compound) GetDouble(name string) float64 {
	if v, ok := c[name].(Double); ok {
		return float64(v)
	}
	return 0
}

// GetString returns the string value for the given name, or "" if not found or wrong type.
func (c Compound) GetString(name string) string {
	if v, ok := c[name].(String); ok {
		return string(v)
	}
	return ""
}

// GetCompound returns the compound value for the given name, or nil if not found or wrong type.
func (c Compound) GetCompound(name string) Compound {
	if v, ok := c[name].(Compound); ok {
		return v
	}
	return nil
}

// GetList returns the list value for the given name, or empty list if not found or wrong type.
func (c Compound) GetList(name string) List {
	if v, ok := c[name].(List); ok {
		return v
	}
	return List{}
}

// GetByteArray returns the byte array for the given name, or nil if not found or wrong type.
func (c Compound) GetByteArray(name string) []byte {
	if v, ok := c[name].(ByteArray); ok {
		return []byte(v)
	}
	return nil
}

// GetIntArray returns the int array for the given name, or nil if not found or wrong type.
func (c Compound) GetIntArray(name string) []int32 {
	if v, ok := c[name].(IntArray); ok {
		return []int32(v)
	}
	return nil
}

// GetLongArray returns the long array for the given name, or nil if not found or wrong type.
func (c Compound) GetLongArray(name string) []int64 {
	if v, ok := c[name].(LongArray); ok {
		return []int64(v)
	}
	return nil
}

// IntArray represents a TAG_Int_Array.
type IntArray []int32

func (IntArray) ID() byte { return TagIntArray }
func (a IntArray) write(w *Writer) error {
	if err := w.writeInt(int32(len(a))); err != nil {
		return err
	}
	for _, v := range a {
		if err := w.writeInt(v); err != nil {
			return err
		}
	}
	return nil
}

// LongArray represents a TAG_Long_Array.
type LongArray []int64

func (LongArray) ID() byte { return TagLongArray }
func (a LongArray) write(w *Writer) error {
	if err := w.writeInt(int32(len(a))); err != nil {
		return err
	}
	for _, v := range a {
		if err := w.writeLong(v); err != nil {
			return err
		}
	}
	return nil
}

// End represents a TAG_End (used to terminate compounds).
// This is typically not used directly.
type End struct{}

func (End) ID() byte              { return TagEnd }
func (End) write(w *Writer) error { return nil }
