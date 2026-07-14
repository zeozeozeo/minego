package nbt

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Stringify converts an NBT tag to its SNBT (Stringified NBT) representation.
// https://minecraft.wiki/w/NBT_format#SNBT_format
func Stringify(tag Tag) string {
	var sb strings.Builder
	writeTag(&sb, tag)
	return sb.String()
}

func writeTag(sb *strings.Builder, tag Tag) {
	switch v := tag.(type) {
	case Byte:
		fmt.Fprintf(sb, "%db", int8(v))
	case Short:
		fmt.Fprintf(sb, "%ds", int16(v))
	case Int:
		fmt.Fprintf(sb, "%d", int32(v))
	case Long:
		fmt.Fprintf(sb, "%dL", int64(v))
	case Float:
		sb.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
		sb.WriteByte('f')
	case Double:
		sb.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 64))
		sb.WriteByte('d')
	case String:
		writeQuotedString(sb, string(v))
	case ByteArray:
		sb.WriteString("[B;")
		for i, b := range v {
			if i > 0 {
				sb.WriteByte(',')
			}
			fmt.Fprintf(sb, "%db", int8(b))
		}
		sb.WriteByte(']')
	case IntArray:
		sb.WriteString("[I;")
		for i, n := range v {
			if i > 0 {
				sb.WriteByte(',')
			}
			fmt.Fprintf(sb, "%d", n)
		}
		sb.WriteByte(']')
	case LongArray:
		sb.WriteString("[L;")
		for i, n := range v {
			if i > 0 {
				sb.WriteByte(',')
			}
			fmt.Fprintf(sb, "%dL", n)
		}
		sb.WriteByte(']')
	case *List:
		sb.WriteByte('[')
		for i, elem := range v.Elements {
			if i > 0 {
				sb.WriteByte(',')
			}
			writeTag(sb, elem)
		}
		sb.WriteByte(']')
	case Compound:
		sb.WriteByte('{')
		first := true
		for k, child := range v {
			if !first {
				sb.WriteByte(',')
			}
			first = false
			if needsQuoting(k) {
				writeQuotedString(sb, k)
			} else {
				sb.WriteString(k)
			}
			sb.WriteByte(':')
			writeTag(sb, child)
		}
		sb.WriteByte('}')
	case End:
		// nothing
	}
}

func writeQuotedString(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		default:
			sb.WriteRune(r)
		}
	}
	sb.WriteByte('"')
}

// needsQuoting returns true if a key needs quoting in SNBT.
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if !isUnquotedChar(r) {
			return true
		}
	}
	return false
}

func isUnquotedChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' || r == '+'
}

// FromJSON converts a JSON value to an NBT tag tree.
// This matches the vanilla server's DynamicOps<Tag> behavior:
//   - JSON object → Compound
//   - JSON array → List
//   - JSON string → String
//   - JSON integer → Int (if fits in int32)
//   - JSON float → Double
//   - JSON boolean → Byte (1/0)
//   - JSON null → Byte(0)
func FromJSON(data []byte) (Tag, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return jsonValueToTag(v)
}

func jsonValueToTag(v any) (Tag, error) {
	switch val := v.(type) {
	case map[string]any:
		compound := make(Compound, len(val))
		for k, child := range val {
			tag, err := jsonValueToTag(child)
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", k, err)
			}
			compound[k] = tag
		}
		return compound, nil
	case []any:
		if len(val) == 0 {
			return &List{ElementType: TagEnd}, nil
		}
		tags := make([]Tag, len(val))
		for i, child := range val {
			tag, err := jsonValueToTag(child)
			if err != nil {
				return nil, fmt.Errorf("index %d: %w", i, err)
			}
			tags[i] = tag
		}
		return &List{ElementType: tags[0].ID(), Elements: tags}, nil
	case string:
		return String(val), nil
	case float64:
		if val == math.Trunc(val) && val >= math.MinInt32 && val <= math.MaxInt32 {
			return Int(int32(val)), nil
		}
		return Double(val), nil
	case bool:
		if val {
			return Byte(1), nil
		}
		return Byte(0), nil
	case nil:
		return Byte(0), nil
	default:
		return nil, fmt.Errorf("unsupported JSON type %T", v)
	}
}
