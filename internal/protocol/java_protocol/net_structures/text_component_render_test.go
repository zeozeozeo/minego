package net_structures_test

import (
	"encoding/json"
	"testing"

	ns "github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
)

func boolPtr(v bool) *bool { return &v }

func TestTextComponent_String(t *testing.T) {
	cases := []struct {
		name string
		tc   ns.TextComponent
		want string
	}{
		{"plain", ns.TextComponent{Text: "Hello"}, "Hello"},
		{"with extra", ns.TextComponent{Text: "Hello, ", Extra: []ns.TextComponent{{Text: "World"}}}, "Hello, World"},
		{"translate", ns.TextComponent{Translate: "chat.type.text"}, "chat.type.text"},
		{"translate with args", ns.TextComponent{Translate: "chat.type.text", With: []ns.TextComponent{{Text: "Player"}, {Text: "Hello"}}}, "chat.type.textPlayerHello"},
		{"nested", ns.TextComponent{Text: "a", Extra: []ns.TextComponent{{Text: "b", Extra: []ns.TextComponent{{Text: "c"}}}}}, "abc"},
		{"empty", ns.TextComponent{}, ""},
		{"keybind", ns.TextComponent{Keybind: "key.jump"}, "key.jump"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.tc.String()
			if got != c.want {
				t.Errorf("String() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestTextComponent_ANSI(t *testing.T) {
	tc := ns.TextComponent{Text: "Hello", Color: "red"}
	got := tc.ANSI()
	if got != "\033[91mHello\033[0m" {
		t.Errorf("ANSI() = %q, want %q", got, "\033[91mHello\033[0m")
	}

	// bold
	tc = ns.TextComponent{Text: "Bold", Bold: boolPtr(true)}
	got = tc.ANSI()
	if got != "\033[1mBold\033[0m" {
		t.Errorf("ANSI() = %q, want %q", got, "\033[1mBold\033[0m")
	}

	// hex color
	tc = ns.TextComponent{Text: "Hex", Color: "#ff5555"}
	got = tc.ANSI()
	if got != "\033[38;2;255;85;85mHex\033[0m" {
		t.Errorf("ANSI() = %q, want %q", got, "\033[38;2;255;85;85mHex\033[0m")
	}

	// no style = no reset
	tc = ns.TextComponent{Text: "Plain"}
	got = tc.ANSI()
	if got != "Plain" {
		t.Errorf("ANSI() = %q, want %q", got, "Plain")
	}
}

func TestTextComponent_ColorCodes(t *testing.T) {
	tc := ns.TextComponent{Text: "Hello", Color: "green", Bold: boolPtr(true)}
	got := tc.ColorCodes()
	if got != "§a§lHello" {
		t.Errorf("ColorCodes() = %q, want %q", got, "§a§lHello")
	}

	// with extra
	tc = ns.TextComponent{
		Text:  "Hello ",
		Color: "gold",
		Extra: []ns.TextComponent{{Text: "World", Color: "red"}},
	}
	got = tc.ColorCodes()
	if got != "§6Hello §cWorld" {
		t.Errorf("ColorCodes() = %q, want %q", got, "§6Hello §cWorld")
	}
}

func TestTextComponent_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		json string
		want string
	}{
		{"plain string", `"Hello"`, "Hello"},
		{"object with text", `{"text":"Hello"}`, "Hello"},
		{"with color", `{"text":"Hello","color":"red"}`, "Hello"},
		{"with extra", `{"text":"Hello ","extra":[{"text":"World"}]}`, "Hello World"},
		{"translate", `{"translate":"chat.type.text","with":[{"text":"Player"},{"text":"msg"}]}`, "chat.type.textPlayermsg"},
		{"nested extra", `{"text":"a","extra":[{"text":"b","extra":[{"text":"c"}]}]}`, "abc"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var tc ns.TextComponent
			if err := json.Unmarshal([]byte(c.json), &tc); err != nil {
				t.Fatalf("Unmarshal(%s) error: %v", c.json, err)
			}
			got := tc.String()
			if got != c.want {
				t.Errorf("String() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestTextComponent_MiniMessage(t *testing.T) {
	tc := ns.TextComponent{Text: "Hello", Color: "red"}
	got := tc.MiniMessage()
	if got != "<red>Hello</red>" {
		t.Errorf("MiniMessage() = %q, want %q", got, "<red>Hello</red>")
	}

	// translate
	tc = ns.TextComponent{
		Translate: "chat.type.text",
		With:      []ns.TextComponent{{Text: "Player"}, {Text: "Hello"}},
	}
	got = tc.MiniMessage()
	if got != "<lang:chat.type.text:Player:Hello>" {
		t.Errorf("MiniMessage() = %q, want %q", got, "<lang:chat.type.text:Player:Hello>")
	}

	// bold + color
	tc = ns.TextComponent{Text: "wow", Color: "gold", Bold: boolPtr(true)}
	got = tc.MiniMessage()
	if got != "<gold><bold>wow</bold></gold>" {
		t.Errorf("MiniMessage() = %q, want %q", got, "<gold><bold>wow</bold></gold>")
	}
}
