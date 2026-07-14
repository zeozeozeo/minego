package net_structures

import (
	"fmt"
	"io"
)

// String is a UTF-8 encoded string with a VarInt length prefix (byte count).
//
// The length prefix indicates the number of bytes, not characters.
// Maximum length is 32767 characters (which can be up to ~130KB in UTF-8).
type String string

// Encode writes the String to w with VarInt length prefix.
func (v String) Encode(w io.Writer) error {
	data := []byte(v)
	if err := VarInt(len(data)).Encode(w); err != nil {
		return fmt.Errorf("failed to write string length: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write string data: %w", err)
	}
	return nil
}

// DecodeString reads a String from r.
// maxLen is the maximum allowed string length in characters (0 = no limit).
func DecodeString(r io.Reader, maxLen int) (String, error) {
	length, err := DecodeVarInt(r)
	if err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}

	if length < 0 {
		return "", fmt.Errorf("negative string length: %d", length)
	}

	// Minecraft strings can have at most 3 bytes per character (UTF-8)
	// Plus some buffer for edge cases
	maxBytes := maxLen * 4
	if maxLen > 0 && int(length) > maxBytes {
		return "", fmt.Errorf("string byte length %d exceeds maximum %d", length, maxBytes)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return "", fmt.Errorf("failed to read string data: %w", err)
	}

	s := string(data)
	if maxLen > 0 && len([]rune(s)) > maxLen {
		return "", fmt.Errorf("string length %d exceeds maximum %d characters", len([]rune(s)), maxLen)
	}

	return String(s), nil
}

// Identifier is a namespaced location string.
//
// Format: "namespace:path" where:
//   - namespace: Only lowercase letters, digits, underscores, hyphens, and periods
//   - path: Same as namespace plus forward slashes
//   - If no colon, defaults to "minecraft" namespace
//
// Examples:
//
//	"minecraft:stone"
//	"minecraft:textures/block/stone.png"
//	"custom:my_item"
type Identifier string

// Encode writes the Identifier to w.
func (v Identifier) Encode(w io.Writer) error {
	return String(v).Encode(w)
}

// DecodeIdentifier reads an Identifier from r.
func DecodeIdentifier(r io.Reader) (Identifier, error) {
	s, err := DecodeString(r, 32767)
	if err != nil {
		return "", err
	}
	return Identifier(s), nil
}

// Namespace returns the namespace part of the identifier.
// Returns "minecraft" if no namespace is specified.
func (id Identifier) Namespace() string {
	s := string(id)
	for i, c := range s {
		if c == ':' {
			return s[:i]
		}
	}
	return "minecraft"
}

// Path returns the path part of the identifier.
func (id Identifier) Path() string {
	s := string(id)
	for i, c := range s {
		if c == ':' {
			return s[i+1:]
		}
	}
	return s
}
