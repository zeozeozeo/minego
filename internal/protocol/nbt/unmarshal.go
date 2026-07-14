package nbt

import (
	"fmt"
	"reflect"
	"strings"
)

// Unmarshal decodes NBT bytes in file format into a Go value.
//
// The target must be a pointer to a struct, map, or other supported type.
// See Marshal for the type mapping.
//
// For network protocol packets, use UnmarshalNetwork instead.
func Unmarshal(data []byte, v any) error {
	return UnmarshalOptions(data, v, false)
}

// UnmarshalNetwork decodes NBT bytes in network format (nameless root) into a Go value.
//
// This is the format used in Minecraft protocol packets.
func UnmarshalNetwork(data []byte, v any) error {
	return UnmarshalOptions(data, v, true)
}

// UnmarshalFile decodes NBT bytes in file format into a Go value.
// Returns the root tag name.
func UnmarshalFile(data []byte, v any) (string, error) {
	tag, rootName, err := DecodeFile(data)
	if err != nil {
		return "", err
	}
	return rootName, UnmarshalTag(tag, v)
}

// UnmarshalOptions decodes NBT bytes into a Go value with full control.
func UnmarshalOptions(data []byte, v any, network bool, opts ...ReaderOption) error {
	tag, _, err := Decode(data, network, opts...)
	if err != nil {
		return err
	}
	return UnmarshalTag(tag, v)
}

// UnmarshalTag converts an NBT Tag to a Go value.
func UnmarshalTag(tag Tag, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("unmarshal target must be a non-nil pointer")
	}
	return unmarshalValue(tag, rv.Elem())
}

// TagUnmarshaler allows types to customize how they are unmarshaled from NBT.
// This is useful for types like TextComponent that can be either a String or Compound.
type TagUnmarshaler interface {
	UnmarshalNBT(tag Tag) error
}

func unmarshalValue(tag Tag, v reflect.Value) error {
	// Handle nil tags
	if tag == nil {
		return nil
	}

	// If target implements Tag, set directly if same type
	if v.Type().Implements(reflect.TypeFor[Tag]()) {
		if reflect.TypeOf(tag).AssignableTo(v.Type()) {
			v.Set(reflect.ValueOf(tag))
			return nil
		}
	}

	// Handle pointer types
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return unmarshalValue(tag, v.Elem())
	}

	// Check if the value implements TagUnmarshaler
	if v.CanAddr() {
		if u, ok := v.Addr().Interface().(TagUnmarshaler); ok {
			return u.UnmarshalNBT(tag)
		}
	}

	// Handle interface types
	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		// any/interface{} - set the tag directly converted to native Go type
		v.Set(reflect.ValueOf(tagToNative(tag)))
		return nil
	}

	switch t := tag.(type) {
	case Byte:
		return unmarshalNumber(int64(t), v)

	case Short:
		return unmarshalNumber(int64(t), v)

	case Int:
		return unmarshalNumber(int64(t), v)

	case Long:
		return unmarshalNumber(int64(t), v)

	case Float:
		return unmarshalFloat(float64(t), v)

	case Double:
		return unmarshalFloat(float64(t), v)

	case String:
		if v.Kind() == reflect.String {
			v.SetString(string(t))
			return nil
		}
		return fmt.Errorf("cannot unmarshal String into %s", v.Type())

	case ByteArray:
		return unmarshalByteArray(t, v)

	case IntArray:
		return unmarshalIntArray(t, v)

	case LongArray:
		return unmarshalLongArray(t, v)

	case List:
		return unmarshalList(t, v)

	case Compound:
		return unmarshalCompound(t, v)

	case End:
		return nil

	default:
		return fmt.Errorf("unknown tag type: %T", tag)
	}
}

func unmarshalNumber(n int64, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		v.SetBool(n != 0)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(float64(n))
	default:
		return fmt.Errorf("cannot unmarshal number into %s", v.Type())
	}
	return nil
}

func unmarshalFloat(f float64, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		v.SetFloat(f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(int64(f))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(f))
	default:
		return fmt.Errorf("cannot unmarshal float into %s", v.Type())
	}
	return nil
}

func unmarshalByteArray(data ByteArray, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes([]byte(data))
			return nil
		}
	case reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			n := min(len(data), v.Len())
			for i := 0; i < n; i++ {
				v.Index(i).SetUint(uint64(data[i]))
			}
			return nil
		}
	}
	return fmt.Errorf("cannot unmarshal ByteArray into %s", v.Type())
}

func unmarshalIntArray(data IntArray, v reflect.Value) error {
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal IntArray into %s", v.Type())
	}

	elemType := v.Type().Elem()
	slice := reflect.MakeSlice(v.Type(), len(data), len(data))

	for i, val := range data {
		elem := slice.Index(i)
		switch elemType.Kind() {
		case reflect.Int, reflect.Int32, reflect.Int64:
			elem.SetInt(int64(val))
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			elem.SetUint(uint64(val))
		default:
			return fmt.Errorf("cannot unmarshal IntArray element into %s", elemType)
		}
	}

	v.Set(slice)
	return nil
}

func unmarshalLongArray(data LongArray, v reflect.Value) error {
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal LongArray into %s", v.Type())
	}

	elemType := v.Type().Elem()
	slice := reflect.MakeSlice(v.Type(), len(data), len(data))

	for i, val := range data {
		elem := slice.Index(i)
		switch elemType.Kind() {
		case reflect.Int, reflect.Int64:
			elem.SetInt(val)
		case reflect.Uint, reflect.Uint64:
			elem.SetUint(uint64(val))
		default:
			return fmt.Errorf("cannot unmarshal LongArray element into %s", elemType)
		}
	}

	v.Set(slice)
	return nil
}

func unmarshalList(list List, v reflect.Value) error {
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal List into %s", v.Type())
	}

	slice := reflect.MakeSlice(v.Type(), len(list.Elements), len(list.Elements))

	for i, elem := range list.Elements {
		if err := unmarshalValue(elem, slice.Index(i)); err != nil {
			return fmt.Errorf("list element %d: %w", i, err)
		}
	}

	v.Set(slice)
	return nil
}

func unmarshalCompound(compound Compound, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Map:
		return unmarshalCompoundToMap(compound, v)
	case reflect.Struct:
		return unmarshalCompoundToStruct(compound, v)
	default:
		return fmt.Errorf("cannot unmarshal Compound into %s", v.Type())
	}
}

func unmarshalCompoundToMap(compound Compound, v reflect.Value) error {
	if v.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("map keys must be strings")
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	elemType := v.Type().Elem()

	for name, tag := range compound {
		elem := reflect.New(elemType).Elem()
		if err := unmarshalValue(tag, elem); err != nil {
			return fmt.Errorf("map key %q: %w", name, err)
		}
		v.SetMapIndex(reflect.ValueOf(name), elem)
	}

	return nil
}

func unmarshalCompoundToStruct(compound Compound, v reflect.Value) error {
	t := v.Type()

	// Build field index by NBT name
	fields := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name, _ := parseTag(field.Tag.Get("nbt"))
		if name == "-" {
			continue
		}
		if name == "" {
			name = field.Name
		}

		// Also try lowercase for flexibility
		fields[name] = i
		fields[strings.ToLower(name)] = i
	}

	for name, tag := range compound {
		idx, ok := fields[name]
		if !ok {
			idx, ok = fields[strings.ToLower(name)]
		}
		if !ok {
			// Unknown field, skip
			continue
		}

		fieldValue := v.Field(idx)
		if err := unmarshalValue(tag, fieldValue); err != nil {
			return fmt.Errorf("field %s: %w", name, err)
		}
	}

	return nil
}

// tagToNative converts an NBT tag to a native Go type for interface{} targets.
func tagToNative(tag Tag) any {
	switch t := tag.(type) {
	case Byte:
		return int8(t)
	case Short:
		return int16(t)
	case Int:
		return int32(t)
	case Long:
		return int64(t)
	case Float:
		return float32(t)
	case Double:
		return float64(t)
	case String:
		return string(t)
	case ByteArray:
		return []byte(t)
	case IntArray:
		return []int32(t)
	case LongArray:
		return []int64(t)
	case List:
		result := make([]any, len(t.Elements))
		for i, elem := range t.Elements {
			result[i] = tagToNative(elem)
		}
		return result
	case Compound:
		result := make(map[string]any)
		for k, v := range t {
			result[k] = tagToNative(v)
		}
		return result
	default:
		return nil
	}
}
