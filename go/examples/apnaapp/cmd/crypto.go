package cmd

import (
	"crypto/rand"

	"golang.org/x/crypto/ed25519"
)

func genKey() (ed25519.PublicKey, ed25519.PrivateKey) {
	pubkey, privkey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return pubkey, privkey
}
