package crypto

// https://minecraft.wiki/w/Protocol_encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
)

type Encryption struct {
	encryptStream cipher.Stream
	decryptStream cipher.Stream
	sharedSecret  []byte
}

func NewEncryption() *Encryption {
	return &Encryption{}
}

func (e *Encryption) GenerateSharedSecret() ([]byte, error) {
	e.sharedSecret = make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, e.sharedSecret); err != nil {
		return nil, fmt.Errorf("failed to generate shared secret: %w", err)
	}
	return e.sharedSecret, nil
}

func (e *Encryption) SetSharedSecret(secret []byte) {
	e.sharedSecret = secret
}

func (e *Encryption) GetSharedSecret() []byte {
	return e.sharedSecret
}

func (e *Encryption) EncryptWithPublicKey(publicKeyBytes []byte, data []byte) ([]byte, error) {
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encrypted, nil
}

func (e *Encryption) DecryptWithPrivateKey(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}
	return decrypted, nil
}

func (e *Encryption) EnableEncryption() error {
	if e.sharedSecret == nil {
		return fmt.Errorf("shared secret not set")
	}

	block, err := aes.NewCipher(e.sharedSecret)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	e.encryptStream = NewEncryptStream(block, e.sharedSecret)
	e.decryptStream = NewDecryptStream(block, e.sharedSecret)

	return nil
}

func (e *Encryption) Encrypt(data []byte) []byte {
	if e.encryptStream == nil {
		return data
	}
	encrypted := make([]byte, len(data))
	e.encryptStream.XORKeyStream(encrypted, data)
	return encrypted
}

func (e *Encryption) Decrypt(data []byte) []byte {
	if e.decryptStream == nil {
		return data
	}
	decrypted := make([]byte, len(data))
	e.decryptStream.XORKeyStream(decrypted, data)
	return decrypted
}

func (e *Encryption) IsEnabled() bool {
	return e.encryptStream != nil && e.decryptStream != nil
}
