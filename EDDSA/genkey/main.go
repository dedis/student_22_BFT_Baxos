package main

import (
	"crypto/ed25519"
	"student_22_BFT_Baxos/EDDSA"
)

const Num = 10

func main() {

	for i, path := range EDDSA.PrivateKeyPath {
		eddsaKey := EDDSA.GenerateEDDSAPrivateKey()

		err := EDDSA.WritePrivateKeyFile(eddsaKey, path)
		if err != nil {
			return
		}

		err = EDDSA.WritePublicKeyFile(eddsaKey.Public().(ed25519.PublicKey), EDDSA.PublicKeyPath[i])
		if err != nil {
			return
		}

		if i == Num-1 {
			break
		}
	}
}
