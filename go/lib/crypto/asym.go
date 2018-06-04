// Copyright 2017 ETH Zurich
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypto

import (
	"crypto/rand"
	"strings"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"

	"github.com/scionproto/scion/go/lib/common"
)

const (
	Ed25519                    = "ed25519"
	Curve25519xSalsa20Poly1305 = "curve25519xsalsa20poly1305"
	InvalidKeySize             = "Invalid key size"
	UnsupportedSignAlgo        = "Unsupported signing algorithm"
	InvalidSignature           = "Invalid signature"
	FailedToGenerateKeyPairs   = "Failed to generate key pairs"
)

// Sign takes a signature input and a signing key to create a signature. Currently only
// ed25519 is supported
func Sign(sigInput, signKey common.RawBytes, signAlgo string) (common.RawBytes, error) {
	switch strings.ToLower(signAlgo) {
	case Ed25519:
		if len(signKey) != ed25519.PrivateKeySize {
			return nil, common.NewBasicError(InvalidKeySize, nil, "expected",
				ed25519.PrivateKeySize, "actual", len(signKey))
		}
		return ed25519.Sign(ed25519.PrivateKey(signKey), sigInput), nil
	default:
		return nil, common.NewBasicError(UnsupportedSignAlgo, nil, "algo", signAlgo)
	}
}

// Verify takes a signature input and a verifying key and returns an error, if the
// signature does not match. Currently only ed25519 is supported.
func Verify(sigInput, sig, verifyKey common.RawBytes, signAlgo string) error {
	switch strings.ToLower(signAlgo) {
	case Ed25519:
		if len(verifyKey) != ed25519.PublicKeySize {
			return common.NewBasicError(InvalidKeySize, nil,
				"expected", ed25519.PublicKeySize, "actual", len(verifyKey))
		}
		if !ed25519.Verify(ed25519.PublicKey(verifyKey), sigInput, sig) {
			return common.NewBasicError(InvalidSignature, nil)
		}
		return nil
	default:
		return common.NewBasicError(UnsupportedSignAlgo, nil, "algo", signAlgo)
	}
}

// GenKeyPairs generates public/private keys pairs
func GenKeyPairs(keygenAlgo string) (common.RawBytes, common.RawBytes, error) {
	switch strings.ToLower(keygenAlgo) {
	case Curve25519xSalsa20Poly1305:
		pubkey, privkey, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, common.NewBasicError(FailedToGenerateKeyPairs, err,
				"keygenAlgo", keygenAlgo)
		}
		return pubkey[:], privkey[:], nil
	case Ed25519:
		pubkey, privkey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, common.NewBasicError(FailedToGenerateKeyPairs, err,
				"keygenAlgo", keygenAlgo)
		}
		return common.RawBytes(pubkey), common.RawBytes(privkey), nil
	default:
		return nil, nil, common.NewBasicError(UnsupportedSignAlgo, nil, "algo", keygenAlgo)
	}
}

// GenSharedSecret generates common secret for a given pairs of keys
func GenSharedSecret(pubkey common.RawBytes, privkey common.RawBytes, algo string) (common.RawBytes,
	error) {
	switch strings.ToLower(algo) {
	case Curve25519xSalsa20Poly1305:
		if len(pubkey) != 32 || len(privkey) != 32 {
			return nil, common.NewBasicError(InvalidKeySize, nil, "algo", algo)
		}
		var pub, priv, secret *[32]byte
		for _, i := range pubkey {
			pub[i] = pubkey[i]
			priv[i] = privkey[i]
		}
		secret = new([32]byte)
		box.Precompute(secret, pub, priv)
		return secret[:], nil
	default:
		return nil, common.NewBasicError(UnsupportedSignAlgo, nil, "algo", algo)
	}
}
