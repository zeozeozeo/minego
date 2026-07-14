package lang_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/data/lang"
	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"item.minecraft.iron_sword", "Iron Sword"},
		{"item.minecraft.diamond_sword", "Diamond Sword"},
		{"item.minecraft.apple", "Apple"},
		{"item.minecraft.golden_apple", "Golden Apple"},
		{"item.minecraft.stick", "Stick"},
		{"item.minecraft.diamond", "Diamond"},
		{"block.minecraft.stone", "Stone"},
		{"block.minecraft.dirt", "Dirt"},
		{"block.minecraft.diamond_block", "Block of Diamond"},
		{"block.minecraft.iron_block", "Block of Iron"},
		{"block.minecraft.oak_planks", "Oak Planks"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := lang.Translate(tt.key); got != tt.want {
				t.Errorf("Translate(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestTranslateNotFound(t *testing.T) {
	if got := lang.Translate("nonexistent.translation.key"); got != "" {
		t.Errorf("Translate for nonexistent key = %q, want empty string", got)
	}
}

func TestTextComponentString(t *testing.T) {
	tests := []struct {
		name string
		tc   ns.TextComponent
		want string
	}{
		{
			"plain text",
			ns.TextComponent{Text: "Hello"},
			"Hello",
		},
		{
			"chat message",
			ns.TextComponent{
				Translate: "chat.type.text",
				With:      []ns.TextComponent{{Text: "Steve"}, {Text: "Hello world"}},
			},
			"<Steve> Hello world",
		},
		{
			"nested translate",
			ns.TextComponent{
				Translate: "chat.type.announcement",
				With:      []ns.TextComponent{{Text: "Server"}, {Text: "Welcome!"}},
			},
			"[Server] Welcome!",
		},
		{
			"with extra",
			ns.TextComponent{
				Text:  "Hello ",
				Extra: []ns.TextComponent{{Text: "World"}},
			},
			"Hello World",
		},
		{
			"unknown key",
			ns.TextComponent{Translate: "nonexistent.key"},
			"nonexistent.key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tc.Render(lang.Translate)
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateUI(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"menu.singleplayer", "Singleplayer"},
		{"menu.multiplayer", "Multiplayer"},
		{"menu.options", "Options..."},
		{"menu.quit", "Quit Game"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := lang.Translate(tt.key); got != tt.want {
				t.Errorf("Translate(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
