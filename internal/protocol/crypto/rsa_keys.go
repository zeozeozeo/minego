package crypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// ParseRSAPrivateKey parses an RSA private key from PEM format
// Supports both PKCS#8 and PKCS#1 formats
func ParseRSAPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
		return rsaPrivateKey, nil
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// ParseRSAPublicKey parses an RSA public key from PEM format
// Supports both PKIX and PKCS#1 formats
func ParseRSAPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	if publicKey, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rsaPublicKey, ok := publicKey.(*rsa.PublicKey); ok {
			return rsaPublicKey, nil
		}
	}

	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// ConvertPublicKeyToSPKI converts an RSA public key to SPKI DER format
func ConvertPublicKeyToSPKI(publicKey *rsa.PublicKey) ([]byte, error) {
	return x509.MarshalPKIXPublicKey(publicKey)
}

// ExtractPublicKeyFromPEM extracts the raw public key bytes from a PEM string
func ExtractPublicKeyFromPEM(publicKeyPEM string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	return block.Bytes, nil
}
