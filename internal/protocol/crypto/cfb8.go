package crypto

// inspired by https://github.com/Tnze/go-mc/blob/076f723e3d1467e8bb11fc09dd29e8e92caf339f/net/CFB8/cfb8.go

import "crypto/cipher"

// Encrypt encrypts the plaintext using CFB8 mode with the given block cipher and IV.
// It returns the ciphertext which will be the same length as the plaintext.
func Encrypt(block cipher.Block, iv, plaintext []byte) []byte {
	stream := newCFB8(block, iv, false)
	ciphertext := make([]byte, len(plaintext))

	stream.xorKeyStream(ciphertext, plaintext)
	return ciphertext
}

// Decrypt decrypts the ciphertext using CFB8 mode with the given block cipher and IV.
// It returns the plaintext which will be the same length as the ciphertext.
func Decrypt(block cipher.Block, iv, ciphertext []byte) []byte {
	stream := newCFB8(block, iv, true)
	plaintext := make([]byte, len(ciphertext))

	stream.xorKeyStream(plaintext, ciphertext)
	return plaintext
}

// Stream exposes a cipher.Stream-compatible wrapper for CFB8.
type Stream struct{ c *cfb8 }

// XORKeyStream satisfies cipher.Stream.
func (s *Stream) XORKeyStream(dst, src []byte) { s.c.xorKeyStream(dst, src) }

// NewEncryptStream creates a cipher.Stream for encryption using CFB8.
func NewEncryptStream(block cipher.Block, iv []byte) cipher.Stream {
	return &Stream{c: newCFB8(block, iv, false)}
}

// NewDecryptStream creates a cipher.Stream for decryption using CFB8.
func NewDecryptStream(block cipher.Block, iv []byte) cipher.Stream {
	return &Stream{c: newCFB8(block, iv, true)}
}

type cfb8 struct {
	block     cipher.Block
	blockSize int
	iv        []byte
	temp      []byte
	decrypt   bool
}

func (c *cfb8) xorKeyStream(dst, src []byte) {
	for i := range src {
		copy(c.temp, c.iv)

		c.block.Encrypt(c.iv, c.iv)
		keystreamByte := c.iv[0]

		outputByte := src[i] ^ keystreamByte
		dst[i] = outputByte
		copy(c.iv, c.temp[1:])

		if c.decrypt {
			c.iv[c.blockSize-1] = src[i]
		} else {
			c.iv[c.blockSize-1] = outputByte
		}
	}
}

func newCFB8(block cipher.Block, iv []byte, decrypt bool) *cfb8 {
	ivCopy := make([]byte, len(iv))
	copy(ivCopy, iv)

	return &cfb8{
		block:     block,
		blockSize: block.BlockSize(),
		iv:        ivCopy,
		temp:      make([]byte, block.BlockSize()),
		decrypt:   decrypt,
	}
}
