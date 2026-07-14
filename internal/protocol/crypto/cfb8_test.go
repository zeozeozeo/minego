package crypto_test

import (
	"crypto/aes"
	"encoding/hex"
	"testing"

	"github.com/zeozeozeo/minego/internal/protocol/crypto"
)

// cfb8TestCases contains test cases for CFB8
// from https://github.com/Tnze/go-mc/blob/076f723e3d1467e8bb11fc09dd29e8e92caf339f/net/CFB8/cfb8_test.go#L15
var cfb8TestCases = []struct {
	key, iv, plaintext, ciphertext string
}{
	{
		"2b7e151628aed2a6abf7158809cf4f3c",
		"000102030405060708090a0b0c0d0e0f",
		"6bc1bee22e409f96e93d7e117393172a",
		"3b79424c9c0dd436bace9e0ed4586a4f",
	},
	{
		"2b7e151628aed2a6abf7158809cf4f3c",
		"3B3FD92EB72DAD20333449F8E83CFB4A",
		"ae2d8a571e03ac9c9eb76fac45af8e51",
		"c8b0723943d71f61a2e5b0e8cedf87c8",
	},
	{
		"2b7e151628aed2a6abf7158809cf4f3c",
		"C8A64537A0B3A93FCDE3CDAD9F1CE58B",
		"30c81c46a35ce411e5fbc1191a0a52ef",
		"260d20e9395d3501067286d3a2a7002f",
	},
	{
		"2b7e151628aed2a6abf7158809cf4f3c",
		"26751F67A3CBB140B1808CF187A4F4DF",
		"f69f2445df4f9b17ad2b417be66c3710",
		"c0af633cd9c599309f924802af599ee6",
	},
	{
		"2b7e151628aed2a6abf7158809cf4f3c",
		"000102030405060708090a0b0c0d0e0f",
		"0ecbd6d36cd12962ce671b4d96fb95aaa902096aeac366e13a6ae57c05d48673cf320c626689d05548f65fd6a108630c1d4e3aab543b006823c7a9422e97c0431587537c384f99a11488ffd9b2e9b46f49005a7e5cef64e27e2de3cf3fb87c1524766601",
		"5efb6f6b93cf5f0e135a0c932f59f9aaa2276e4b06cd4f5edca4baba735ac7708dd7c0f9e92c6b89d2245b0d9a6356b0e98529cd45e56df22e914ef9e0792facaab707af90c13162bfad06a240eb6adcbf3365fd84a003f8083f4662a7a27232c72c6c0c",
	},
}

func TestCFB8Encrypt(t *testing.T) {
	for i, tc := range cfb8TestCases {
		key, _ := hex.DecodeString(tc.key)
		iv, _ := hex.DecodeString(tc.iv)
		plaintext, _ := hex.DecodeString(tc.plaintext)

		block, err := aes.NewCipher(key)
		if err != nil {
			t.Fatalf("Test %d: Failed to create AES cipher: %v", i, err)
		}

		ciphertext := crypto.Encrypt(block, iv, plaintext)

		if hex.EncodeToString(ciphertext) != tc.ciphertext {
			t.Errorf("Test %d: Encryption failed\nExpected: %s\nGot:      %s",
				i, tc.ciphertext, hex.EncodeToString(ciphertext))
		}
	}
}

func TestCFB8Decrypt(t *testing.T) {
	for i, tc := range cfb8TestCases {
		key, _ := hex.DecodeString(tc.key)
		iv, _ := hex.DecodeString(tc.iv)
		ciphertext, _ := hex.DecodeString(tc.ciphertext)

		block, err := aes.NewCipher(key)
		if err != nil {
			t.Fatalf("Test %d: Failed to create AES cipher: %v", i, err)
		}

		plaintext := crypto.Decrypt(block, iv, ciphertext)

		if hex.EncodeToString(plaintext) != tc.plaintext {
			t.Errorf("Test %d: Decryption failed\nExpected: %s\nGot:      %s",
				i, tc.plaintext, hex.EncodeToString(plaintext))
		}
	}
}
