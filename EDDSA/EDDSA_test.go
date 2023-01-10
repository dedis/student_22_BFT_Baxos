package EDDSA

import (
	"crypto/ed25519"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_Sign_Verify(t *testing.T) {
	//seed := "dd16afe4ff9ecb20567c4a638cef8ee276e938d8617479296936497b9f80fd70"
	message := []byte("123456")

	privateKey := GenerateEDDSAPrivateKey()
	publicKey := privateKey.Public().(ed25519.PublicKey)

	//sign the message
	sig := ed25519.Sign(privateKey, message)
	//verify the signature
	res := ed25519.Verify(publicKey, message, sig)

	require.Equal(t, true, res)
}

func Test_Write_Read_Sign_Verify(t *testing.T) {
	eddsaKey := GenerateEDDSAPrivateKey()

	err := WritePrivateKeyFile(eddsaKey, "test1.txt")
	if err != nil {
		return
	}

	err = WritePublicKeyFile(eddsaKey.Public().(ed25519.PublicKey), "test2.txt")
	if err != nil {
		return
	}

	privateKey, err := ReadPrivateKeyFile("test1.txt")
	if err != nil {
		return
	}
	publicKey, err := ReadPublicKeyFile("test2.txt")
	if err != nil {
		return
	}

	err = os.Remove("test1.txt")
	if err != nil {
		return
	}

	err = os.Remove("test2.txt")
	if err != nil {
		return
	}

	// write should equal to read
	require.Equal(t, eddsaKey, privateKey)
	require.Equal(t, eddsaKey.Public(), publicKey)

	// Sign and Verify test
	message := []byte("123456")
	//sign the message
	sig := ed25519.Sign(privateKey, message)
	//verify the signature
	res := ed25519.Verify(publicKey, message, sig)

	require.Equal(t, true, res)
}
