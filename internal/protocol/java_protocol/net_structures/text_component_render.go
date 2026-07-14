package net_structures

import (
	"fmt"
	"strings"
)

// MC color name -> ANSI escape code
var mcColorToANSI = map[string]string{
	"black":        "\033[30m",
	"dark_blue":    "\033[34m",
	"dark_green":   "\033[32m",
	"dark_aqua":    "\033[36m",
	"dark_red":     "\033[31m",
	"dark_purple":  "\033[35m",
	"gold":         "\033[33m",
	"gray":         "\033[37m",
	"dark_gray":    "\033[90m",
	"blue":         "\033[94m",
	"green":        "\033[92m",
	"aqua":         "\033[96m",
	"red":          "\033[91m",
	"light_purple": "\033[95m",
	"yellow":       "\033[93m",
	"white":        "\033[97m",
}

// MC color name - Bukkit section code
var mcColorToCode = map[string]string{
	"black":        "§0",
	"dark_blue":    "§1",
	"dark_green":   "§2",
	"dark_aqua":    "§3",
	"dark_red":     "§4",
	"dark_purple":  "§5",
	"gold":         "§6",
	"gray":         "§7",
	"dark_gray":    "§8",
	"blue":         "§9",
	"green":        "§a",
	"aqua":         "§b",
	"red":          "§c",
	"light_purple": "§d",
	"yellow":       "§e",
	"white":        "§f",
}

// componentWriter writes a single component's full content (content + extras) to b.
type componentWriter func(tc *TextComponent, b *strings.Builder)

// writeContent writes the resolved content of this component (without extras).
// If translate is non-nil and the component has a translate key, the key is
// resolved and %s / %N$s placeholders are substituted with With args rendered
// via the provided writer. Otherwise the raw text/key is written.
func (tc *TextComponent) writeContent(b *strings.Builder, write componentWriter, translate func(string) string) {
	if tc.Translate != "" {
		if translate != nil {
			if pattern := translate(tc.Translate); pattern != "" {
				writeFormatted(b, pattern, tc.With, write)
				return
			}
		}
		// no translation available, show key + with args as-is
		b.WriteString(tc.Translate)
		for i := range tc.With {
			write(&tc.With[i], b)
		}
		return
	}

	b.WriteString(tc.Text)
	b.WriteString(tc.Keybind)
	if tc.Score != nil {
		b.WriteString(tc.Score.Name)
	}
	b.WriteString(tc.Selector)
}

// writeFormatted handles MC's Java-style format strings (%s, %1$s, %%, %d).
func writeFormatted(b *strings.Builder, pattern string, args []TextComponent, write componentWriter) {
	seqIdx := 0
	i := 0
	for i < len(pattern) {
		if pattern[i] != '%' || i+1 >= len(pattern) {
			b.WriteByte(pattern[i])
			i++
			continue
		}

		j := i + 1

		// %%
		if pattern[j] == '%' {
			b.WriteByte('%')
			i = j + 1
			continue
		}

		// %N$s (positional)
		if j+2 < len(pattern) && pattern[j] >= '1' && pattern[j] <= '9' && pattern[j+1] == '$' && pattern[j+2] == 's' {
			argIdx := int(pattern[j]-'0') - 1
			if argIdx >= 0 && argIdx < len(args) {
				write(&args[argIdx], b)
			}
			i = j + 3
			continue
		}

		// %s
		if pattern[j] == 's' {
			if seqIdx < len(args) {
				write(&args[seqIdx], b)
				seqIdx++
			}
			i = j + 1
			continue
		}

		// %d
		if pattern[j] == 'd' {
			if seqIdx < len(args) {
				write(&args[seqIdx], b)
				seqIdx++
			}
			i = j + 1
			continue
		}

		// unknown format specifier, output literally
		fmt.Fprintf(b, "%%%c", pattern[j])
		i = j + 1
	}
}

// String returns the plain text content of the component and all children,
// with no formatting. Translate keys are shown as-is.
func (tc TextComponent) String() string {
	return tc.Render(nil)
}

// Render returns the plain text with translate keys resolved by fn (if non-nil).
func (tc TextComponent) Render(translate func(string) string) string {
	var b strings.Builder
	tc.writePlain(&b, translate)
	return b.String()
}

func (tc *TextComponent) writePlain(b *strings.Builder, translate func(string) string) {
	tc.writeContent(b, func(child *TextComponent, b *strings.Builder) {
		child.writePlain(b, translate)
	}, translate)
	for i := range tc.Extra {
		tc.Extra[i].writePlain(b, translate)
	}
}

// ANSI returns the text with ANSI terminal escape codes for colors and formatting.
// Translate keys are shown as-is.
func (tc TextComponent) ANSI() string {
	return tc.RenderANSI(nil)
}

// RenderANSI returns ANSI-formatted text with translate keys resolved by fn (if non-nil).
func (tc TextComponent) RenderANSI(translate func(string) string) string {
	var b strings.Builder
	if tc.writeANSI(&b, translate) {
		b.WriteString("\033[0m")
	}
	return b.String()
}

func (tc *TextComponent) writeANSI(b *strings.Builder, translate func(string) string) bool {
	prefix := tc.ansiPrefix()
	styled := prefix != ""
	if styled {
		b.WriteString(prefix)
	}

	tc.writeContent(b, func(child *TextComponent, b *strings.Builder) {
		if child.writeANSI(b, translate) {
			styled = true
		}
	}, translate)

	for i := range tc.Extra {
		// reset before each styled child so parent style doesn't bleed
		if styled {
			b.WriteString("\033[0m")
		}
		if tc.Extra[i].writeANSI(b, translate) {
			styled = true
		}
	}
	return styled
}

func (tc *TextComponent) ansiPrefix() string {
	var codes []string

	if tc.Color != "" {
		if ansi, ok := mcColorToANSI[tc.Color]; ok {
			codes = append(codes, ansi)
		} else if strings.HasPrefix(tc.Color, "#") && len(tc.Color) == 7 {
			// hex color → 24-bit ANSI
			var r, g, b int
			fmt.Sscanf(tc.Color[1:], "%02x%02x%02x", &r, &g, &b)
			codes = append(codes, fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b))
		}
	}
	if tc.Bold != nil && *tc.Bold {
		codes = append(codes, "\033[1m")
	}
	if tc.Italic != nil && *tc.Italic {
		codes = append(codes, "\033[3m")
	}
	if tc.Underlined != nil && *tc.Underlined {
		codes = append(codes, "\033[4m")
	}
	if tc.Strikethrough != nil && *tc.Strikethrough {
		codes = append(codes, "\033[9m")
	}
	if tc.Obfuscated != nil && *tc.Obfuscated {
		codes = append(codes, "\033[8m")
	}

	return strings.Join(codes, "")
}

// ColorCodes returns the text with Bukkit-style section sign (§) color codes.
// Translate keys are shown as-is.
func (tc TextComponent) ColorCodes() string {
	return tc.RenderColorCodes(nil)
}

// RenderColorCodes returns section-sign colored text with translate keys resolved by fn (if non-nil).
func (tc TextComponent) RenderColorCodes(translate func(string) string) string {
	var b strings.Builder
	tc.writeColorCodes(&b, translate)
	return b.String()
}

func (tc *TextComponent) writeColorCodes(b *strings.Builder, translate func(string) string) {
	if tc.Color != "" {
		if code, ok := mcColorToCode[tc.Color]; ok {
			b.WriteString(code)
		}
	}
	if tc.Bold != nil && *tc.Bold {
		b.WriteString("§l")
	}
	if tc.Italic != nil && *tc.Italic {
		b.WriteString("§o")
	}
	if tc.Underlined != nil && *tc.Underlined {
		b.WriteString("§n")
	}
	if tc.Strikethrough != nil && *tc.Strikethrough {
		b.WriteString("§m")
	}
	if tc.Obfuscated != nil && *tc.Obfuscated {
		b.WriteString("§k")
	}

	tc.writeContent(b, func(child *TextComponent, b *strings.Builder) {
		child.writeColorCodes(b, translate)
	}, translate)
	for i := range tc.Extra {
		tc.Extra[i].writeColorCodes(b, translate)
	}
}

// MiniMessage returns the text in Adventure MiniMessage format.
// Translate keys are shown as-is.
func (tc TextComponent) MiniMessage() string {
	return tc.RenderMiniMessage(nil)
}

// RenderMiniMessage returns MiniMessage-formatted text with translate keys resolved by fn (if non-nil).
func (tc TextComponent) RenderMiniMessage(translate func(string) string) string {
	var b strings.Builder
	tc.writeMiniMessage(&b, translate)
	return b.String()
}

func (tc *TextComponent) writeMiniMessage(b *strings.Builder, translate func(string) string) {
	var tags []string

	if tc.Color != "" {
		tags = append(tags, tc.Color)
	}
	if tc.Bold != nil && *tc.Bold {
		tags = append(tags, "bold")
	}
	if tc.Italic != nil && *tc.Italic {
		tags = append(tags, "italic")
	}
	if tc.Underlined != nil && *tc.Underlined {
		tags = append(tags, "underlined")
	}
	if tc.Strikethrough != nil && *tc.Strikethrough {
		tags = append(tags, "strikethrough")
	}
	if tc.Obfuscated != nil && *tc.Obfuscated {
		tags = append(tags, "obfuscated")
	}

	for _, tag := range tags {
		b.WriteByte('<')
		b.WriteString(tag)
		b.WriteByte('>')
	}

	if tc.Translate != "" {
		b.WriteString("<lang:")
		b.WriteString(tc.Translate)
		for _, arg := range tc.With {
			b.WriteByte(':')
			arg.writeMiniMessage(b, translate)
		}
		b.WriteByte('>')
	} else if tc.Keybind != "" {
		b.WriteString("<key:")
		b.WriteString(tc.Keybind)
		b.WriteByte('>')
	} else {
		b.WriteString(tc.Text)
		if tc.Score != nil {
			b.WriteString(tc.Score.Name)
		}
		b.WriteString(tc.Selector)
	}

	for _, child := range tc.Extra {
		child.writeMiniMessage(b, translate)
	}

	// close tags in reverse
	for i := len(tags) - 1; i >= 0; i-- {
		b.WriteString("</")
		b.WriteString(tags[i])
		b.WriteByte('>')
	}
}
