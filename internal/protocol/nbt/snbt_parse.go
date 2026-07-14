package nbt

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses an SNBT (stringified NBT) string into a Tag tree. It accepts the
// vanilla command grammar: compounds {k:v,...} with quoted or bare keys, lists
// [v,...], typed arrays [B;...]/[I;...]/[L;...], number suffixes b/s/l/f/d,
// single- or double-quoted strings, and bare words (true/false -> Byte 1/0).
func Parse(s string) (Tag, error) {
	p := &snbtParser{s: s}
	t, err := p.value(0)
	if err != nil {
		return nil, err
	}
	p.ws()
	if p.pos != len(p.s) {
		return nil, fmt.Errorf("snbt: trailing data at offset %d", p.pos)
	}
	return t, nil
}

type snbtParser struct {
	s   string
	pos int
}

func (p *snbtParser) ws() {
	for p.pos < len(p.s) {
		switch p.s[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *snbtParser) peek() byte {
	if p.pos < len(p.s) {
		return p.s[p.pos]
	}
	return 0
}

func (p *snbtParser) value(depth int) (Tag, error) {
	if depth > MaxDepth {
		return nil, fmt.Errorf("snbt: max depth %d exceeded", MaxDepth)
	}
	p.ws()
	if p.pos >= len(p.s) {
		return nil, fmt.Errorf("snbt: unexpected end of input")
	}
	switch p.s[p.pos] {
	case '{':
		return p.compound(depth)
	case '[':
		return p.listOrArray(depth)
	case '"', '\'':
		s, err := p.quoted()
		return String(s), err
	default:
		return p.literal()
	}
}

func (p *snbtParser) compound(depth int) (Tag, error) {
	p.pos++ // consume '{'
	c := Compound{}
	p.ws()
	if p.peek() == '}' {
		p.pos++
		return c, nil
	}
	for {
		p.ws()
		key, err := p.key()
		if err != nil {
			return nil, err
		}
		p.ws()
		if p.peek() != ':' {
			return nil, fmt.Errorf("snbt: expected ':' after key %q at offset %d", key, p.pos)
		}
		p.pos++
		v, err := p.value(depth + 1)
		if err != nil {
			return nil, err
		}
		c[key] = v
		p.ws()
		switch p.peek() {
		case ',':
			p.pos++
		case '}':
			p.pos++
			return c, nil
		default:
			return nil, fmt.Errorf("snbt: expected ',' or '}' at offset %d", p.pos)
		}
	}
}

func (p *snbtParser) key() (string, error) {
	if c := p.peek(); c == '"' || c == '\'' {
		return p.quoted()
	}
	tok := p.readToken()
	if tok == "" {
		return "", fmt.Errorf("snbt: empty key at offset %d", p.pos)
	}
	return tok, nil
}

func (p *snbtParser) listOrArray(depth int) (Tag, error) {
	if p.pos+2 < len(p.s) && p.s[p.pos+2] == ';' {
		switch p.s[p.pos+1] {
		case 'B':
			return p.byteArray()
		case 'I':
			return p.intArray()
		case 'L':
			return p.longArray()
		}
	}
	p.pos++ // consume '['
	list := List{}
	p.ws()
	if p.peek() == ']' {
		p.pos++
		return list, nil
	}
	for {
		v, err := p.value(depth + 1)
		if err != nil {
			return nil, err
		}
		if len(list.Elements) == 0 {
			list.ElementType = v.ID()
		} else if v.ID() != list.ElementType {
			return nil, fmt.Errorf("snbt: mixed list element types at offset %d", p.pos)
		}
		list.Elements = append(list.Elements, v)
		p.ws()
		switch p.peek() {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return list, nil
		default:
			return nil, fmt.Errorf("snbt: expected ',' or ']' at offset %d", p.pos)
		}
	}
}

func (p *snbtParser) byteArray() (Tag, error) {
	p.pos += 3 // consume "[B;"
	var arr ByteArray
	err := p.arrayElems(func(tok string) error {
		v, ok := parseIntToken(tok)
		if !ok {
			return fmt.Errorf("snbt: invalid byte-array element %q", tok)
		}
		arr = append(arr, byte(int8(v)))
		return nil
	})
	return arr, err
}

func (p *snbtParser) intArray() (Tag, error) {
	p.pos += 3 // consume "[I;"
	var arr IntArray
	err := p.arrayElems(func(tok string) error {
		v, ok := parseIntToken(tok)
		if !ok {
			return fmt.Errorf("snbt: invalid int-array element %q", tok)
		}
		arr = append(arr, int32(v))
		return nil
	})
	return arr, err
}

func (p *snbtParser) longArray() (Tag, error) {
	p.pos += 3 // consume "[L;"
	var arr LongArray
	err := p.arrayElems(func(tok string) error {
		v, ok := parseIntToken(tok)
		if !ok {
			return fmt.Errorf("snbt: invalid long-array element %q", tok)
		}
		arr = append(arr, v)
		return nil
	})
	return arr, err
}

func (p *snbtParser) arrayElems(add func(tok string) error) error {
	p.ws()
	if p.peek() == ']' {
		p.pos++
		return nil
	}
	for {
		p.ws()
		tok := p.readToken()
		if tok == "" {
			return fmt.Errorf("snbt: empty array element at offset %d", p.pos)
		}
		if err := add(tok); err != nil {
			return err
		}
		p.ws()
		switch p.peek() {
		case ',':
			p.pos++
		case ']':
			p.pos++
			return nil
		default:
			return fmt.Errorf("snbt: expected ',' or ']' in array at offset %d", p.pos)
		}
	}
}

func (p *snbtParser) quoted() (string, error) {
	q := p.s[p.pos]
	p.pos++
	var sb strings.Builder
	for p.pos < len(p.s) {
		c := p.s[p.pos]
		switch c {
		case '\\':
			p.pos++
			if p.pos >= len(p.s) {
				return "", fmt.Errorf("snbt: dangling escape")
			}
			sb.WriteByte(p.s[p.pos]) // lenient: keep the escaped char verbatim
			p.pos++
		case q:
			p.pos++
			return sb.String(), nil
		default:
			sb.WriteByte(c)
			p.pos++
		}
	}
	return "", fmt.Errorf("snbt: unterminated string")
}

func (p *snbtParser) readToken() string {
	start := p.pos
	for p.pos < len(p.s) && isBareword(p.s[p.pos]) {
		p.pos++
	}
	return p.s[start:p.pos]
}

func (p *snbtParser) literal() (Tag, error) {
	tok := p.readToken()
	if tok == "" {
		return nil, fmt.Errorf("snbt: unexpected character %q at offset %d", p.s[p.pos], p.pos)
	}
	return interpretLiteral(tok), nil
}

func isBareword(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' ||
		c == '_' || c == '-' || c == '.' || c == '+'
}

func interpretLiteral(tok string) Tag {
	switch tok {
	case "true":
		return Byte(1)
	case "false":
		return Byte(0)
	}
	if t, ok := parseNumber(tok); ok {
		return t
	}
	return String(tok)
}

func parseNumber(tok string) (Tag, bool) {
	if tok == "" {
		return nil, false
	}
	body := tok[:len(tok)-1]
	switch tok[len(tok)-1] {
	case 'b', 'B':
		if v, err := strconv.ParseInt(body, 10, 16); err == nil {
			return Byte(int8(v)), true
		}
	case 's', 'S':
		if v, err := strconv.ParseInt(body, 10, 32); err == nil {
			return Short(int16(v)), true
		}
	case 'l', 'L':
		if v, err := strconv.ParseInt(body, 10, 64); err == nil {
			return Long(v), true
		}
	case 'f', 'F':
		if v, err := strconv.ParseFloat(body, 32); err == nil {
			return Float(float32(v)), true
		}
	case 'd', 'D':
		if v, err := strconv.ParseFloat(body, 64); err == nil {
			return Double(v), true
		}
	}
	if v, err := strconv.ParseInt(tok, 10, 32); err == nil {
		return Int(int32(v)), true
	}
	if v, err := strconv.ParseInt(tok, 10, 64); err == nil {
		return Long(v), true
	}
	if v, err := strconv.ParseFloat(tok, 64); err == nil {
		return Double(v), true
	}
	return nil, false
}

// parseIntToken parses an integer array element, ignoring a trailing type suffix.
func parseIntToken(tok string) (int64, bool) {
	if n := len(tok); n > 0 {
		switch tok[n-1] {
		case 'b', 'B', 's', 'S', 'l', 'L':
			tok = tok[:n-1]
		}
	}
	v, err := strconv.ParseInt(tok, 10, 64)
	return v, err == nil
}
