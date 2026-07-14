package crypto_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/protocol/crypto"
)

var sha1TestCases = map[string]string{
	"Notch": "4ed1f46bbe04bc756bcb17c0c7ce3e4632f06a48",
	"jeb_":  "-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1",
	"simon": "88e16a1019277b15d58faf0541e11910eb756f6",
}

func TestMinecraftSHA1(t *testing.T) {
	for username, expected := range sha1TestCases {
		actual := crypto.MinecraftSHA1(username)
		if actual != expected {
			t.Errorf("MinecraftSHA1(%q) = %q; want %q", username, actual, expected)
		}
	}
}
