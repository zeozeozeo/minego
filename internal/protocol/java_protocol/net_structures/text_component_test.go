package net_structures_test

import (
	"bytes"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

// TextComponent wire format: NBT tag (network format, no root name)
//   - Simple text: NBT String tag (0x08) + length + UTF-8 bytes
//   - Complex: NBT Compound tag (0x0a) + fields + end tag (0x00)
//
// Reference: https://minecraft.wiki/w/Java_Edition_protocol/Data_types#Text_Component

var textComponentTestCases = []struct {
	name string
	raw  []byte
	text string
}{
	{
		name: "simple text",
		raw:  []byte{0x08, 0x00, 0x05, 'H', 'e', 'l', 'l', 'o'},
		text: "Hello",
	},
	{
		name: "world",
		raw:  []byte{0x08, 0x00, 0x05, 'W', 'o', 'r', 'l', 'd'},
		text: "World",
	},
}

func TestTextComponent(t *testing.T) {
	for _, tc := range textComponentTestCases {
		t.Run(tc.name+" decode", func(t *testing.T) {
			var got ns.TextComponent
			buf := ns.NewReader(tc.raw)
			if err := got.Decode(buf); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if got.Text != tc.text {
				t.Errorf("Text mismatch: got %q, want %q", got.Text, tc.text)
			}
		})

		t.Run(tc.name+" encode", func(t *testing.T) {
			comp := ns.TextComponent{Text: tc.text}
			buf := ns.NewWriter()
			if err := comp.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), tc.raw) {
				t.Errorf("encode mismatch:\n  got:  %x\n  want: %x", buf.Bytes(), tc.raw)
			}
		})
	}
}

func TestTextComponent_Complex(t *testing.T) {
	cases := []ns.TextComponent{
		{Text: "Styled", Color: "red"},
		{Text: "Hello, ", Extra: []ns.TextComponent{{Text: "World", Color: "gold"}}},
		{Translate: "chat.type.text", With: []ns.TextComponent{{Text: "Player"}, {Text: "Hello"}}},
		{Text: "Click me", ClickEvent: &ns.ClickEvent{Action: "open_url", URL: "https://minecraft.net"}},
	}

	for _, tc := range cases {
		name := tc.Text
		if name == "" {
			name = tc.Translate
		}
		t.Run(name, func(t *testing.T) {
			buf := ns.NewWriter()
			if err := tc.Encode(buf); err != nil {
				t.Fatalf("encode error: %v", err)
			}

			var decoded ns.TextComponent
			if err := decoded.Decode(ns.NewReader(buf.Bytes())); err != nil {
				t.Fatalf("decode error: %v", err)
			}

			if decoded.Text != tc.Text {
				t.Errorf("Text mismatch: got %q, want %q", decoded.Text, tc.Text)
			}
			if decoded.Color != tc.Color {
				t.Errorf("Color mismatch: got %q, want %q", decoded.Color, tc.Color)
			}
			if decoded.Translate != tc.Translate {
				t.Errorf("Translate mismatch: got %q, want %q", decoded.Translate, tc.Translate)
			}
			if len(decoded.Extra) != len(tc.Extra) {
				t.Errorf("Extra length mismatch: got %d, want %d", len(decoded.Extra), len(tc.Extra))
			}
			if len(decoded.With) != len(tc.With) {
				t.Errorf("With length mismatch: got %d, want %d", len(decoded.With), len(tc.With))
			}

			// re-encode should match
			buf2 := ns.NewWriter()
			if err := decoded.Encode(buf2); err != nil {
				t.Fatalf("re-encode error: %v", err)
			}
			if !bytes.Equal(buf.Bytes(), buf2.Bytes()) {
				t.Errorf("round-trip mismatch:\n  original:   %x\n  re-encoded: %x", buf.Bytes(), buf2.Bytes())
			}
		})
	}
}

func TestTextComponent_StringOptimization(t *testing.T) {
	simple := ns.NewTextComponent("Hello")
	styled := ns.TextComponent{Text: "Hello", Color: "red"}

	simpleBuf := ns.NewWriter()
	simple.Encode(simpleBuf)

	styledBuf := ns.NewWriter()
	styled.Encode(styledBuf)

	if simpleBuf.Bytes()[0] != 0x08 {
		t.Errorf("simple text should use String tag (0x08), got 0x%02x", simpleBuf.Bytes()[0])
	}
	if styledBuf.Bytes()[0] != 0x0a {
		t.Errorf("styled text should use Compound tag (0x0a), got 0x%02x", styledBuf.Bytes()[0])
	}
	if simpleBuf.Len() >= styledBuf.Len() {
		t.Errorf("simple (%d bytes) should be smaller than styled (%d bytes)", simpleBuf.Len(), styledBuf.Len())
	}
}
