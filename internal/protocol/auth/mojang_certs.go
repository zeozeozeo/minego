package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"
)

type MojangKeyPair struct {
	PrivateKey string `json:"privateKey"`
	PublicKey  string `json:"publicKey"`
}

type MojangCertificate struct {
	ExpiresAt            string        `json:"expiresAt"`
	KeyPair              MojangKeyPair `json:"keyPair"`
	PublicKeySignature   string        `json:"publicKeySignature"`
	PublicKeySignatureV2 string        `json:"publicKeySignatureV2"`
	RefreshedAfter       string        `json:"refreshedAfter"`
}

type MojangCertificateData struct {
	Certificate    *MojangCertificate
	PrivateKey     *rsa.PrivateKey
	PublicKey      *rsa.PublicKey
	PublicKeyBytes []byte // SPKI DER format
	SignatureBytes []byte // Mojang signature V2 (512 bytes)
	ExpiryTime     time.Time
}

// FetchMojangCertificate retrieves the Mojang signing certificate for chat messages
func FetchMojangCertificate(accessToken string) (*MojangCertificateData, error) {
	req, err := http.NewRequest("POST", "https://api.minecraftservices.com/player/certificates", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch certificate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("certificate request failed with status %d", resp.StatusCode)
	}

	var cert MojangCertificate
	if err := json.NewDecoder(resp.Body).Decode(&cert); err != nil {
		return nil, fmt.Errorf("failed to parse certificate response: %w", err)
	}

	privateKey, err := parseRSAPrivateKey(cert.KeyPair.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey, err := parseRSAPublicKey(cert.KeyPair.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	expiryTime, err := time.Parse(time.RFC3339Nano, cert.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate expiry time: %w", err)
	}

	block, _ := pem.Decode([]byte(cert.KeyPair.PublicKey))
	var publicKeyBytes []byte
	if block != nil {
		publicKeyBytes = block.Bytes // SPKI DER
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(cert.PublicKeySignatureV2)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Mojang signature: %w", err)
	}

	return &MojangCertificateData{
		Certificate:    &cert,
		PrivateKey:     privateKey,
		PublicKey:      publicKey,
		PublicKeyBytes: publicKeyBytes,
		SignatureBytes: signatureBytes,
		ExpiryTime:     expiryTime,
	}, nil
}

func parseRSAPrivateKey(privateKeyPEM string) (*rsa.PrivateKey, error) {
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

func parseRSAPublicKey(publicKeyPEM string) (*rsa.PublicKey, error) {
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
