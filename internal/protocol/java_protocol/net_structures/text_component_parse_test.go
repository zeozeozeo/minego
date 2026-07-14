package net_structures_test

import (
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func TestFromColorCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // round-trip via ColorCodes()
	}{
		{"plain", "Hello", "Hello"},
		{"single color", "§6Hello", "§6Hello"},
		{"color + format", "§6§lHello", "§6§lHello"},
		{"multiple segments", "§6Hello §cWorld", "§6Hello §cWorld"},
		{"reset", "§6Hello§r World", "§6Hello World"},
		{"format only", "§lBold", "§lBold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := ns.FromColorCodes(tt.input)
			got := tc.ColorCodes()
			if got != tt.want {
				t.Errorf("FromColorCodes(%q).ColorCodes() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFromColorCodesPlainText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "Hello", "Hello"},
		{"colored", "§6Hello", "Hello"},
		{"multi", "§6Hello §cWorld", "Hello World"},
		{"formatted", "§6§lHello", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := ns.FromColorCodes(tt.input)
			got := tc.String()
			if got != tt.want {
				t.Errorf("FromColorCodes(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFromMiniMessage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // round-trip via MiniMessage()
	}{
		{"plain", "Hello", "Hello"},
		{"single color", "<gold>Hello</gold>", "<gold>Hello</gold>"},
		{"nested", "<gold>Hello <bold>world</bold></gold>", "<gold>Hello </gold><gold><bold>world</bold></gold>"},
		{"hex color", "<#ff0000>Red</#ff0000>", "<#ff0000>Red</#ff0000>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := ns.FromMiniMessage(tt.input)
			got := tc.MiniMessage()
			if got != tt.want {
				t.Errorf("FromMiniMessage(%q).MiniMessage() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFromMiniMessagePlainText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "Hello", "Hello"},
		{"colored", "<gold>Hello</gold>", "Hello"},
		{"multi", "<gold>Hello </gold><red>World</red>", "Hello World"},
		{"reset", "<gold>Hello</gold><reset>World", "HelloWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := ns.FromMiniMessage(tt.input)
			got := tc.String()
			if got != tt.want {
				t.Errorf("FromMiniMessage(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
