package main

import (
	"student_22_BFT_Baxos/ECDSA"
)

func main() {
	for i, path := range ECDSA.PrivateKeyPath {
		ecdsaKey, err := ECDSA.GenerateECDSAPrivateKey()
		if err != nil {
			return
		}

		err = ECDSA.WritePrivateKeyFile(ecdsaKey, path)
		if err != nil {
			return
		}

		err = ECDSA.WritePublicKeyFile(&ecdsaKey.PublicKey, ECDSA.PublicKeyPath[i])
		if err != nil {
			return
		}
	}
}
