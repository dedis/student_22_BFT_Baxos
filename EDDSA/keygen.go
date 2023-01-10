package EDDSA

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os"
)

const (
	VRFKeySeed string = "dd16afe4ff9ecb20567c4a638cef8ee276e938d8617479296936497b9f80fd70"
)

var (
	NodeName       = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	PrivateKeyPath = []string{
		"doc/privateKey/node1.txt",
		"doc/privateKey/node2.txt",
		"doc/privateKey/node3.txt",
		"doc/privateKey/node4.txt",
		"doc/privateKey/node5.txt",
		"doc/privateKey/node6.txt",
		"doc/privateKey/node7.txt",
		"doc/privateKey/node8.txt",
		"doc/privateKey/node9.txt",
		"doc/privateKey/node10.txt",
	}
	PublicKeyPath = []string{
		"doc/publicKey/node1.txt",
		"doc/publicKey/node2.txt",
		"doc/publicKey/node3.txt",
		"doc/publicKey/node4.txt",
		"doc/publicKey/node5.txt",
		"doc/publicKey/node6.txt",
		"doc/publicKey/node7.txt",
		"doc/publicKey/node8.txt",
		"doc/publicKey/node9.txt",
		"doc/publicKey/node10.txt",
	}
)

func GenerateEDDSAPrivateKey() (pk ed25519.PrivateKey) {
	// generate the vrf key pairs
	//givenSeedInBytes, err := hex.DecodeString(VRFKeySeed)
	//if err != nil {
	//	fmt.Printf("failed to decode seed from given seed")
	//}
	//vrfPrivateKey := ed25519.NewKeyFromSeed(givenSeedInBytes)
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic("Failed to generate a key")
	}
	return privateKey
}

func WritePrivateKeyFile(key ed25519.PrivateKey, filePath string) (err error) {
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
	_, err = f.Write([]byte(hex.EncodeToString(key)))
	return
}

func WritePublicKeyFile(key ed25519.PublicKey, filePath string) (err error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}

	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()

	_, err = f.Write([]byte(hex.EncodeToString(key)))
	return err
}

func ReadPrivateKeyFile(keyFile string) (ed25519.PrivateKey, error) {
	str, err := os.ReadFile(keyFile)
	if err != nil {
		fmt.Println("read private key error ", err)
		return nil, err
	}
	key, _ := hex.DecodeString(string(str))
	return key, nil
}

func ReadPublicKeyFile(keyFile string) (ed25519.PublicKey, error) {
	str, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	key, _ := hex.DecodeString(string(str))
	return key, nil
}
