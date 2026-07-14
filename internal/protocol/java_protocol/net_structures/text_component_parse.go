package net_structures

import (
	"strings"
	"unicode/utf8"
)

// reverse lookup: section code char → MC color name
var codeToMcColor = map[rune]string{
	'0': "black",
	'1': "dark_blue",
	'2': "dark_green",
	'3': "dark_aqua",
	'4': "dark_red",
	'5': "dark_purple",
	'6': "gold",
	'7': "gray",
	'8': "dark_gray",
	'9': "blue",
	'a': "green", 'A': "green",
	'b': "aqua", 'B': "aqua",
	'c': "red", 'C': "red",
	'd': "light_purple", 'D': "light_purple",
	'e': "yellow", 'E': "yellow",
	'f': "white", 'F': "white",
}

// reverse lookup: MC color name → tag name used in MiniMessage (same for all standard colors)
var miniMessageColors = map[string]bool{
	"black": true, "dark_blue": true, "dark_green": true, "dark_aqua": true,
	"dark_red": true, "dark_purple": true, "gold": true, "gray": true,
	"dark_gray": true, "blue": true, "green": true, "aqua": true,
	"red": true, "light_purple": true, "yellow": true, "white": true,
}

var miniMessageFormats = map[string]func(tc *TextComponent){
	"bold":          func(tc *TextComponent) { t := true; tc.Bold = &t },
	"italic":        func(tc *TextComponent) { t := true; tc.Italic = &t },
	"underlined":    func(tc *TextComponent) { t := true; tc.Underlined = &t },
	"strikethrough": func(tc *TextComponent) { t := true; tc.Strikethrough = &t },
	"obfuscated":    func(tc *TextComponent) { t := true; tc.Obfuscated = &t },
}

// FromColorCodes parses a string with Bukkit-style section sign (§) color/format codes
// into a TextComponent tree.
//
//	FromColorCodes("§6Hello §lworld") → gold "Hello " + gold+bold "world"
func FromColorCodes(s string) TextComponent {
	root := TextComponent{}
	var current *TextComponent
	var buf strings.Builder

	flush := func() {
		text := buf.String()
		buf.Reset()
		if text == "" {
			return
		}
		if current == nil {
			if root.Text == "" && len(root.Extra) == 0 {
				root.Text = text
				return
			}
			root.Extra = append(root.Extra, TextComponent{Text: text})
			return
		}
		current.Text = text
		root.Extra = append(root.Extra, *current)
		current = nil
	}

	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])

		// check for § followed by a code character
		if r == '§' && i+size < len(s) {
			code, codeSize := utf8.DecodeRuneInString(s[i+size:])

			if color, ok := codeToMcColor[code]; ok {
				// color resets formatting
				flush()
				current = &TextComponent{Color: color}
				i += size + codeSize
				continue
			}

			switch code {
			case 'l', 'L':
				flush()
				if current == nil {
					current = &TextComponent{}
				}
				t := true
				current.Bold = &t
			case 'o', 'O':
				flush()
				if current == nil {
					current = &TextComponent{}
				}
				t := true
				current.Italic = &t
			case 'n', 'N':
				flush()
				if current == nil {
					current = &TextComponent{}
				}
				t := true
				current.Underlined = &t
			case 'm', 'M':
				flush()
				if current == nil {
					current = &TextComponent{}
				}
				t := true
				current.Strikethrough = &t
			case 'k', 'K':
				flush()
				if current == nil {
					current = &TextComponent{}
				}
				t := true
				current.Obfuscated = &t
			case 'r', 'R':
				// reset
				flush()
				current = nil
			default:
				buf.WriteRune(r)
				i += size
				continue
			}
			i += size + codeSize
			continue
		}

		buf.WriteRune(r)
		i += size
	}

	flush()
	return root
}

// FromMiniMessage parses a subset of Adventure MiniMessage format into a TextComponent tree.
// Supports color tags (<gold>, <#ff0000>), format tags (<bold>, <italic>, etc.),
// <reset>, and <lang:key:arg1:arg2>.
//
//	FromMiniMessage("<gold>Hello <bold>world</bold></gold>")
func FromMiniMessage(s string) TextComponent {
	root := TextComponent{}
	var buf strings.Builder

	type styleFrame struct {
		tag string
		tc  TextComponent // style accumulated so far
	}
	var stack []styleFrame

	currentStyle := func() TextComponent {
		var tc TextComponent
		for _, f := range stack {
			mergeStyle(&tc, &f.tc)
		}
		return tc
	}

	flush := func() {
		text := buf.String()
		buf.Reset()
		if text == "" {
			return
		}
		styled := currentStyle()
		styled.Text = text
		if root.Text == "" && len(root.Extra) == 0 && styled.Color == "" && styled.Bold == nil &&
			styled.Italic == nil && styled.Underlined == nil && styled.Strikethrough == nil && styled.Obfuscated == nil {
			root.Text = text
		} else {
			root.Extra = append(root.Extra, styled)
		}
	}

	i := 0
	for i < len(s) {
		if s[i] == '<' {
			end := strings.IndexByte(s[i:], '>')
			if end == -1 {
				buf.WriteByte(s[i])
				i++
				continue
			}

			tag := s[i+1 : i+end]
			i += end + 1

			// closing tag
			if strings.HasPrefix(tag, "/") {
				closeTag := tag[1:]
				flush()
				// pop matching tag from stack
				for j := len(stack) - 1; j >= 0; j-- {
					if stack[j].tag == closeTag {
						stack = append(stack[:j], stack[j+1:]...)
						break
					}
				}
				continue
			}

			// reset
			if tag == "reset" {
				flush()
				stack = nil
				continue
			}

			// lang (translate)
			if strings.HasPrefix(tag, "lang:") {
				flush()
				parts := strings.Split(tag[5:], ":")
				tc := currentStyle()
				tc.Translate = parts[0]
				for _, arg := range parts[1:] {
					tc.With = append(tc.With, TextComponent{Text: arg})
				}
				root.Extra = append(root.Extra, tc)
				continue
			}

			// hex color
			if strings.HasPrefix(tag, "#") && len(tag) == 7 {
				flush()
				stack = append(stack, styleFrame{tag: tag, tc: TextComponent{Color: tag}})
				continue
			}

			// named color
			if miniMessageColors[tag] {
				flush()
				stack = append(stack, styleFrame{tag: tag, tc: TextComponent{Color: tag}})
				continue
			}

			// format tag
			if applyFmt, ok := miniMessageFormats[tag]; ok {
				flush()
				var tc TextComponent
				applyFmt(&tc)
				stack = append(stack, styleFrame{tag: tag, tc: tc})
				continue
			}

			// unknown tag, output literally
			buf.WriteByte('<')
			buf.WriteString(tag)
			buf.WriteByte('>')
			continue
		}

		buf.WriteByte(s[i])
		i++
	}

	flush()
	return root
}

// mergeStyle copies non-zero style fields from src onto dst.
func mergeStyle(dst, src *TextComponent) {
	if src.Color != "" {
		dst.Color = src.Color
	}
	if src.Bold != nil {
		dst.Bold = src.Bold
	}
	if src.Italic != nil {
		dst.Italic = src.Italic
	}
	if src.Underlined != nil {
		dst.Underlined = src.Underlined
	}
	if src.Strikethrough != nil {
		dst.Strikethrough = src.Strikethrough
	}
	if src.Obfuscated != nil {
		dst.Obfuscated = src.Obfuscated
	}
}
