package session_server_test

import (
	"testing"

	"github.com/zeozeozeo/minego/internal/protocol/java_protocol/net_structures"
	"github.com/zeozeozeo/minego/internal/protocol/java_protocol/session_server"
)

func TestComputeServerHash(t *testing.T) {
	// https://github.com/PrismarineJS/node-yggdrasil/blob/c2b1e534dc56d33d8ea0c1ba02ead058b9db07b1/test/index.js#L70
	serverID := "cat"
	sharedSecret := []byte("cat")
	publicKey := []byte("cat")

	result := session_server.ComputeServerHash(serverID, sharedSecret, publicKey)
	expected := "-af59e5b1d5d92e5c2c2776ed0e65e90be181f2a"

	if result != expected {
		t.Errorf("ComputeServerHash() = %q, expected %q", result, expected)
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		uuid     string
		expected bool
	}{
		// valid UUIDs without dashes
		{"550e8400e29b41d4a716446655440000", true},
		{"a1b2c3d4e5f6789012345678901234ab", true},

		// valid UUIDs with dashes
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"a1b2c3d4-e5f6-7890-1234-5678901234ab", true},

		// invalid UUIDs
		{"", false},
		{"invalid", false},
		{"550e8400e29b41d4a716446655440000x", false},     // 33 chars
		{"550e8400e29b41d4a71644665544000", false},       // 31 chars
		{"550e8400-e29b-41d4-a716-44665544000", false},   // missing char
		{"550e8400-e29b-41d4-a716-446655440000x", false}, // extra char
		{"550e8400xe29b-41d4-a716-446655440000", false},  // invalid char
	}

	for _, test := range tests {
		result := net_structures.ValidateUUID(test.uuid)
		if result != test.expected {
			t.Errorf("ValidateUUID(%q) = %v, expected %v", test.uuid, result, test.expected)
		}
	}
}

func TestValidateAccessToken(t *testing.T) {
	tests := []struct {
		token    string
		expected bool
	}{
		{"validtoken123456", true},
		{"a_very_long_access_token_that_is_longer_than_usual_access_token_but_should_still_be_valid", true},
		{"", false},                         // empty
		{"short", false},                    // too short
		{string(make([]byte, 3000)), false}, // too long
	}

	for _, test := range tests {
		result := session_server.ValidateAccessToken(test.token)
		if result != test.expected {
			t.Errorf("ValidateAccessToken(%q) = %v, expected %v", test.token, result, test.expected)
		}
	}
}
