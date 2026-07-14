package net_structures

import (
	"strings"
)

// color code mappings (§ and & codes)
var colorCodes = map[byte]string{
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
	'a': "green",
	'b': "aqua",
	'c': "red",
	'd': "light_purple",
	'e': "yellow",
	'f': "white",
}

var boolTrue = true

// ParseFormatted parses a string containing legacy color codes (§ or &)
// and MiniMessage-style tags (<red>, <bold>, <gradient:red:blue>, etc.)
// into a TextComponent tree.
func ParseFormatted(s string) TextComponent {
	// first pass: normalize § to & for uniform processing
	s = strings.ReplaceAll(s, "§", "&")

	// check if string has any formatting
	if !strings.ContainsAny(s, "&<") {
		return TextComponent{Text: s}
	}

	// parse MiniMessage tags first, then legacy codes within each segment
	root := TextComponent{Text: ""}
	parseMixed(s, &root)

	if len(root.Extra) == 1 && root.Text == "" {
		return root.Extra[0]
	}
	return root
}

// parseMixed handles a string that may contain both <tags> and &codes.
func parseMixed(s string, parent *TextComponent) {
	for len(s) > 0 {
		// find next tag
		tagStart := strings.Index(s, "<")
		if tagStart == -1 {
			// no more tags, parse remaining as legacy
			parseLegacyCodes(s, parent)
			return
		}

		// parse legacy codes before the tag
		if tagStart > 0 {
			parseLegacyCodes(s[:tagStart], parent)
		}

		// find closing >
		tagEnd := strings.Index(s[tagStart:], ">")
		if tagEnd == -1 {
			// malformed tag, treat rest as literal
			parseLegacyCodes(s[tagStart:], parent)
			return
		}
		tagEnd += tagStart

		tagContent := s[tagStart+1 : tagEnd]
		s = s[tagEnd+1:]

		// check if it's a closing tag
		if strings.HasPrefix(tagContent, "/") {
			// closing tag — return to parent (handled by recursion)
			continue
		}

		// find the matching closing tag
		closingTag := "</" + tagContent + ">"
		// handle parameterized tags like <gradient:red:blue>
		baseTag := tagContent
		if idx := strings.Index(baseTag, ":"); idx != -1 {
			closingTag = "</" + baseTag[:idx] + ">"
			baseTag = baseTag[:idx]
		}

		closeIdx := strings.Index(s, closingTag)
		var inner string
		if closeIdx != -1 {
			inner = s[:closeIdx]
			s = s[closeIdx+len(closingTag):]
		} else {
			// no closing tag — rest of string is content
			inner = s
			s = ""
		}

		child := applyTag(tagContent)
		parseMixed(inner, &child)
		parent.Extra = append(parent.Extra, child)
	}
}

// applyTag creates a TextComponent with the style for a tag name.
func applyTag(tag string) TextComponent {
	tc := TextComponent{Text: ""}

	// handle parameterized tags
	parts := strings.SplitN(tag, ":", 2)
	name := strings.ToLower(parts[0])

	switch name {
	// colors
	case "black", "dark_blue", "dark_green", "dark_aqua", "dark_red",
		"dark_purple", "gold", "gray", "dark_gray", "blue", "green",
		"aqua", "red", "light_purple", "yellow", "white":
		tc.Color = name

	// formatting
	case "bold", "b":
		tc.Bold = &boolTrue
	case "italic", "i", "em":
		tc.Italic = &boolTrue
	case "underlined", "underline", "u":
		tc.Underlined = &boolTrue
	case "strikethrough", "st":
		tc.Strikethrough = &boolTrue
	case "obfuscated", "obf", "magic":
		tc.Obfuscated = &boolTrue

	// hex color: <#ff5555> or <color:#ff5555>
	case "color":
		if len(parts) > 1 {
			tc.Color = parts[1]
		}

	// gradient (simplified: use first color)
	case "gradient":
		if len(parts) > 1 {
			colors := strings.Split(parts[1], ":")
			if len(colors) > 0 {
				tc.Color = colors[0]
			}
		}

	// reset
	case "reset":
		boolFalse := false
		tc.Color = "white"
		tc.Bold = &boolFalse
		tc.Italic = &boolFalse
		tc.Underlined = &boolFalse
		tc.Strikethrough = &boolFalse
		tc.Obfuscated = &boolFalse

	default:
		// try hex: <#ff5555>
		if strings.HasPrefix(name, "#") {
			tc.Color = name
		}
	}

	return tc
}

// parseLegacyCodes parses &-codes in a string and appends segments to parent.
func parseLegacyCodes(s string, parent *TextComponent) {
	if !strings.Contains(s, "&") {
		if s != "" {
			parent.Extra = append(parent.Extra, TextComponent{Text: s})
		}
		return
	}

	var current TextComponent
	i := 0
	for i < len(s) {
		if s[i] == '&' && i+1 < len(s) {
			code := toLower(s[i+1])

			// check color code
			if color, ok := colorCodes[code]; ok {
				// flush current segment
				if current.Text != "" {
					parent.Extra = append(parent.Extra, current)
				}
				current = TextComponent{Color: color}
				i += 2
				continue
			}

			// check formatting codes
			switch code {
			case 'l':
				if current.Text != "" {
					parent.Extra = append(parent.Extra, current)
					current = TextComponent{Color: current.Color}
				}
				current.Bold = &boolTrue
				i += 2
				continue
			case 'o':
				if current.Text != "" {
					parent.Extra = append(parent.Extra, current)
					current = TextComponent{Color: current.Color}
				}
				current.Italic = &boolTrue
				i += 2
				continue
			case 'n':
				if current.Text != "" {
					parent.Extra = append(parent.Extra, current)
					current = TextComponent{Color: current.Color}
				}
				current.Underlined = &boolTrue
				i += 2
				continue
			case 'm':
				if current.Text != "" {
					parent.Extra = append(parent.Extra, current)
					current = TextComponent{Color: current.Color}
				}
				current.Strikethrough = &boolTrue
				i += 2
				continue
			case 'k':
				if current.Text != "" {
					parent.Extra = append(parent.Extra, current)
					current = TextComponent{Color: current.Color}
				}
				current.Obfuscated = &boolTrue
				i += 2
				continue
			case 'r':
				// reset
				if current.Text != "" {
					parent.Extra = append(parent.Extra, current)
				}
				current = TextComponent{}
				i += 2
				continue
			}
		}

		current.Text += string(s[i])
		i++
	}

	if current.Text != "" {
		parent.Extra = append(parent.Extra, current)
	}
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

// PlainText extracts the plain text content from a TextComponent tree,
// stripping all formatting.
func (tc *TextComponent) PlainText() string {
	var sb strings.Builder
	tc.plainTextInto(&sb)
	return sb.String()
}

func (tc *TextComponent) plainTextInto(sb *strings.Builder) {
	sb.WriteString(tc.Text)
	for i := range tc.Extra {
		tc.Extra[i].plainTextInto(sb)
	}
}
