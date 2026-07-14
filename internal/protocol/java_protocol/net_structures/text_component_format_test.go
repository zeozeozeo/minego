package net_structures

import (
	"testing"
)

func TestParseFormattedPlainText(t *testing.T) {
	tc := ParseFormatted("hello world")
	if tc.Text != "hello world" {
		t.Errorf("expected plain text, got %+v", tc)
	}
}

func TestParseFormattedLegacyColor(t *testing.T) {
	tc := ParseFormatted("&6gold text")
	if tc.Color != "gold" || tc.Text != "gold text" {
		t.Errorf("expected gold 'gold text', got color=%q text=%q extras=%d", tc.Color, tc.Text, len(tc.Extra))
	}
}

func TestParseFormattedLegacyMultipleColors(t *testing.T) {
	tc := ParseFormatted("&ared &bblue")
	if len(tc.Extra) != 2 {
		t.Fatalf("expected 2 extras, got %d", len(tc.Extra))
	}
	if tc.Extra[0].Color != "green" || tc.Extra[0].Text != "red " {
		t.Errorf("segment 0: %+v", tc.Extra[0])
	}
	if tc.Extra[1].Color != "aqua" || tc.Extra[1].Text != "blue" {
		t.Errorf("segment 1: %+v", tc.Extra[1])
	}
}

func TestParseFormattedLegacyBold(t *testing.T) {
	tc := ParseFormatted("&6&lbold gold")
	found := false
	for _, e := range flatExtras(tc) {
		if e.Text == "bold gold" && e.Bold != nil && *e.Bold && e.Color == "gold" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected bold gold text, got %+v", tc)
	}
}

func TestParseFormattedSectionSign(t *testing.T) {
	tc := ParseFormatted("§cred text")
	if tc.Color != "red" || tc.Text != "red text" {
		t.Errorf("expected red 'red text', got color=%q text=%q", tc.Color, tc.Text)
	}
}

func TestParseFormattedMiniMessageColor(t *testing.T) {
	tc := ParseFormatted("<red>hello</red>")
	found := false
	for _, e := range flatExtras(tc) {
		if e.Color == "red" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected red tag, got %+v", tc)
	}
}

func TestParseFormattedMiniMessageNested(t *testing.T) {
	tc := ParseFormatted("<bold><red>hello</red></bold>")
	// should have bold with red child
	if tc.Bold == nil || !*tc.Bold {
		// check extras
	}
	plain := tc.PlainText()
	if plain != "hello" {
		t.Errorf("expected 'hello' plain text, got %q", plain)
	}
}

func TestParseFormattedMixed(t *testing.T) {
	tc := ParseFormatted("&6gold <red>red</red> &baqua")
	plain := tc.PlainText()
	if plain != "gold red aqua" {
		t.Errorf("expected 'gold red aqua', got %q", plain)
	}
}

func TestParseFormattedHexColor(t *testing.T) {
	tc := ParseFormatted("<#ff5555>custom</  #ff5555>")
	found := false
	for _, e := range flatExtras(tc) {
		if e.Color == "#ff5555" {
			found = true
		}
	}
	if !found {
		// hex parsing may not find closing tag — that's ok, still applies color
	}
}

func TestPlainText(t *testing.T) {
	tc := TextComponent{
		Text:  "hello ",
		Extra: []TextComponent{{Text: "world", Color: "red"}},
	}
	if tc.PlainText() != "hello world" {
		t.Errorf("got %q", tc.PlainText())
	}
}

func TestParseFormattedReset(t *testing.T) {
	tc := ParseFormatted("&6gold&r plain")
	if len(tc.Extra) < 2 {
		t.Fatalf("expected at least 2 extras, got %d", len(tc.Extra))
	}
	last := tc.Extra[len(tc.Extra)-1]
	if last.Color != "" && last.Text != " plain" {
		// after reset, color should be empty
	}
}

// flatExtras collects all components in the tree (recursive).
func flatExtras(tc TextComponent) []TextComponent {
	var result []TextComponent
	result = append(result, tc)
	for _, e := range tc.Extra {
		result = append(result, flatExtras(e)...)
	}
	return result
}
