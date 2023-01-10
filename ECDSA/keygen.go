package ECDSA

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	// PrivateKeyFileType is the PEM type for a private key.
	PrivateKeyFileType = "ECDSA PRIVATE KEY"

	// PublicKeyFileType is the PEM type for a public key.
	PublicKeyFileType = "ECDSA PUBLIC KEY"
)

var (
	NodeName       = []string{"1", "2"}
	PrivateKeyPath = []string{"../crypto/ECDSA/privateKey/node1.txt", "../crypto/ECDSA/privateKey/node2.txt"}
	PublicKeyPath  = []string{"../crypto/ECDSA/publicKey/node1.txt", "../crypto/ECDSA/publicKey/node2.txt"}
)

// GenerateECDSAPrivateKey returns a new ECDSA private key.
func GenerateECDSAPrivateKey() (pk *ecdsa.PrivateKey, err error) {
	pk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

// PrivateKeyToPEM encodes the private key in PEM format.
func PrivateKeyToPEM(key *ecdsa.PrivateKey) ([]byte, error) {
	var (
		marshalled []byte
		keyType    string
		err        error
	)

	marshalled, err = x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	keyType = PrivateKeyFileType

	b := &pem.Block{
		Type:  keyType,
		Bytes: marshalled,
	}
	return pem.EncodeToMemory(b), nil
}

// WritePrivateKeyFile writes a private key to the specified file.
func WritePrivateKeyFile(key *ecdsa.PrivateKey, filePath string) (err error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	b, err := PrivateKeyToPEM(key)
	if err != nil {
		return
	}

	_, err = f.Write(b)
	return
}

// PublicKeyToPEM encodes the public key in PEM format.
func PublicKeyToPEM(key *ecdsa.PublicKey) ([]byte, error) {
	var (
		marshalled []byte
		keyType    string
		err        error
	)

	marshalled, err = x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return nil, err
	}
	keyType = PublicKeyFileType

	b := &pem.Block{
		Type:  keyType,
		Bytes: marshalled,
	}

	return pem.EncodeToMemory(b), nil
}

// WritePublicKeyFile writes a public key to the specified file.
func WritePublicKeyFile(key *ecdsa.PublicKey, filePath string) (err error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}

	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	b, err := PublicKeyToPEM(key)
	if err != nil {
		return err
	}

	_, err = f.Write(b)
	return err
}

// ParsePrivateKey parses a PEM encoded private key.
func ParsePrivateKey(buf []byte) (key *ecdsa.PrivateKey, err error) {
	b, _ := pem.Decode(buf)
	key, err = x509.ParseECPrivateKey(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}
	return key, nil
}

// ReadPrivateKeyFile reads a private key from the specified file.
func ReadPrivateKeyFile(keyFile string) (key *ecdsa.PrivateKey, err error) {
	b, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return ParsePrivateKey(b)
}

// ParsePublicKey parses a PEM encoded public key
func ParsePublicKey(buf []byte) (*ecdsa.PublicKey, error) {
	b, _ := pem.Decode(buf)
	if b == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKIXPublicKey(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}
	return key.(*ecdsa.PublicKey), nil
}

// ReadPublicKeyFile reads a public key from the specified file.
func ReadPublicKeyFile(keyFile string) (key *ecdsa.PublicKey, err error) {
	b, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	return ParsePublicKey(b)
}
