package auth

import (
	"bytes"
	"testing"
)

func TestProtectRoundTrip(t *testing.T) {
	plain := []byte(`{"refreshToken":"secret"}`)
	cipher, err := protect(plain)
	if err != nil {
		t.Fatal(err)
	}
	got, err := unprotect(cipher)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("round trip = %q", got)
	}
}
