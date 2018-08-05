package apnams

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
)

// GenKeyPairs returns public-private key pairs for ApnaID
func GenKeyPairs() (common.RawBytes, common.RawBytes, error) {
	return crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
}

// GenSharedSecret returns shared secret for diffie-hellman kex
func GenSharedSecret(pubkey, privkey common.RawBytes) (common.RawBytes, error) {
	return crypto.GenSharedSecret(pubkey, privkey, crypto.Curve25519xSalsa20Poly1305)
}
