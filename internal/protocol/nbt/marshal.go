package nbt

import (
	"fmt"
	"reflect"
	"strings"
)

// Marshal converts a Go value to NBT bytes in file format (with empty root name).
//
// The following Go types map to NBT types:
//   - bool         → Byte (0 or 1)
//   - int8         → Byte
//   - int16        → Short
//   - int32, int   → Int
//   - int64        → Long
//   - float32      → Float
//   - float64      → Double
//   - string       → String
//   - []byte       → ByteArray
//   - []int32      → IntArray
//   - []int64      → LongArray
//   - []T          → List (where T maps to a single NBT type)
//   - struct       → Compound
//   - map[string]T → Compound
//
// Struct fields can be tagged with `nbt:"name"` to specify the NBT key name.
// Use `nbt:"-"` to skip a field. Use `nbt:"name,omitempty"` to omit zero values.
//
// For network protocol packets, use MarshalNetwork instead.
func Marshal(v any) ([]byte, error) {
	return MarshalOptions(v, "", false)
}

// MarshalNetwork converts a Go value to NBT bytes in network format (nameless root).
//
// This is the format used in Minecraft protocol packets. The root compound tag
// has no name, saving 2 bytes compared to file format.
func MarshalNetwork(v any) ([]byte, error) {
	return MarshalOptions(v, "", true)
}

// MarshalFile converts a Go value to NBT bytes in file format with root name.
func MarshalFile(v any, rootName string) ([]byte, error) {
	return MarshalOptions(v, rootName, false)
}

// MarshalOptions converts a Go value to NBT bytes with full control.
func MarshalOptions(v any, rootName string, network bool) ([]byte, error) {
	tag, err := MarshalTag(v)
	if err != nil {
		return nil, err
	}
	return Encode(tag, rootName, network)
}

// MarshalTag converts a Go value to an NBT Tag without encoding to bytes.
func MarshalTag(v any) (Tag, error) {
	return marshalValue(reflect.ValueOf(v))
}

func marshalValue(v reflect.Value) (Tag, error) {
	// handle nil
	if !v.IsValid() {
		return Compound{}, nil
	}

	// dereference pointers
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return Compound{}, nil
		}
		v = v.Elem()
	}

	// check if it implements Tag interface
	if tag, ok := v.Interface().(Tag); ok {
		return tag, nil
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return Byte(1), nil
		}
		return Byte(0), nil

	case reflect.Int8:
		return Byte(v.Int()), nil

	case reflect.Int16:
		return Short(v.Int()), nil

	case reflect.Int32, reflect.Int:
		return Int(v.Int()), nil

	case reflect.Int64:
		return Long(v.Int()), nil

	case reflect.Uint8:
		return Byte(v.Uint()), nil

	case reflect.Uint16:
		return Short(v.Uint()), nil

	case reflect.Uint32, reflect.Uint:
		return Int(v.Uint()), nil

	case reflect.Uint64:
		return Long(v.Uint()), nil

	case reflect.Float32:
		return Float(v.Float()), nil

	case reflect.Float64:
		return Double(v.Float()), nil

	case reflect.String:
		return String(v.String()), nil

	case reflect.Slice:
		return marshalSlice(v)

	case reflect.Array:
		return marshalSlice(v)

	case reflect.Map:
		return marshalMap(v)

	case reflect.Struct:
		return marshalStruct(v)

	default:
		return nil, fmt.Errorf("cannot marshal type %s to NBT", v.Type())
	}
}

func marshalSlice(v reflect.Value) (Tag, error) {
	// Special cases for typed arrays
	switch v.Type().Elem().Kind() {
	case reflect.Uint8:
		// []byte → ByteArray
		if v.Kind() == reflect.Slice {
			return ByteArray(v.Bytes()), nil
		}
		// [N]byte → ByteArray
		data := make([]byte, v.Len())
		for i := 0; i < v.Len(); i++ {
			data[i] = byte(v.Index(i).Uint())
		}
		return ByteArray(data), nil

	case reflect.Int32:
		// []int32 → IntArray
		data := make(IntArray, v.Len())
		for i := 0; i < v.Len(); i++ {
			data[i] = int32(v.Index(i).Int())
		}
		return data, nil

	case reflect.Int64:
		// []int64 → LongArray
		data := make(LongArray, v.Len())
		for i := 0; i < v.Len(); i++ {
			data[i] = v.Index(i).Int()
		}
		return data, nil
	}

	// generic slice → List
	if v.Len() == 0 {
		return List{ElementType: TagEnd, Elements: nil}, nil
	}

	elements := make([]Tag, v.Len())
	var elemType byte

	for i := 0; i < v.Len(); i++ {
		elem, err := marshalValue(v.Index(i))
		if err != nil {
			return nil, fmt.Errorf("list element %d: %w", i, err)
		}
		elements[i] = elem

		if i == 0 {
			elemType = elem.ID()
		} else if elem.ID() != elemType {
			return nil, fmt.Errorf("list has mixed types: %s and %s",
				TagName(elemType), TagName(elem.ID()))
		}
	}

	return List{ElementType: elemType, Elements: elements}, nil
}

func marshalMap(v reflect.Value) (Tag, error) {
	if v.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("map keys must be strings, got %s", v.Type().Key())
	}

	compound := make(Compound)

	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		value, err := marshalValue(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("map key %q: %w", key, err)
		}
		compound[key] = value
	}

	return compound, nil
}

func marshalStruct(v reflect.Value) (Tag, error) {
	compound := make(Compound)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// skip unexported fields
		if !field.IsExported() {
			continue
		}

		// parse tag
		name, opts := parseTag(field.Tag.Get("nbt"))
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}

		// handle omitempty
		if opts.Contains("omitempty") && isEmptyValue(fieldValue) {
			continue
		}

		tag, err := marshalValue(fieldValue)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", field.Name, err)
		}

		compound[name] = tag
	}

	return compound, nil
}

// tagOptions is the string following a comma in a struct field tag.
type tagOptions string

// parseTag splits a struct field's nbt tag into name and options.
func parseTag(tag string) (string, tagOptions) {
	if before, after, ok := strings.Cut(tag, ","); ok {
		return before, tagOptions(after)
	}
	return tag, ""
}

// Contains reports whether a comma-separated list contains the option.
func (o tagOptions) Contains(opt string) bool {
	for o != "" {
		var next string
		if i := strings.Index(string(o), ","); i >= 0 {
			next = string(o[i+1:])
			o = o[:i]
		} else {
			next = ""
		}
		if string(o) == opt {
			return true
		}
		o = tagOptions(next)
	}
	return false
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}
