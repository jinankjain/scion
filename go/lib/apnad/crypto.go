package apnad

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"io"

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

const (
	IvPad  = 12
	IvLen  = 4
	MacLen = 4
)

func ComputeMac(iv, finalEphID []byte) ([]byte, error) {
	message := append(iv, finalEphID...)
	mac := hmac.New(sha256.New, ApnadConfig.HMACKey)
	// TODO(jinankjain): Check bound on n
	_, err := mac.Write(message)
	if err != nil {
		return nil, err
	}
	expectedMAC := mac.Sum(nil)
	return expectedMAC[:MacLen], nil
}

func VerifyMac(message, msgMac []byte) (bool, error) {
	mac := hmac.New(sha256.New, ApnadConfig.HMACKey)
	_, err := mac.Write(message)
	if err != nil {
		return false, err
	}
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(expectedMAC[:MacLen], msgMac), nil
}

func DecryptEphID(iv, msg []byte) (*EphID, error) {
	block, err := aes.NewCipher(ApnadConfig.AESKey)
	if err != nil {
		return nil, err
	}
	padIV := make([]byte, IvPad)
	iv = append(iv, padIV...)
	plaintext := make([]byte, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	var ephID EphID
	for i := 0; i < EphIDLen; i++ {
		ephID[i] = ciphertext[aes.BlockSize+i] ^ msg[i]
	}
	return &ephID, nil
}

func EncryptEphID(ephID *EphID) ([]byte, []byte, error) {
	plaintext := make([]byte, aes.BlockSize)
	block, err := aes.NewCipher(ApnadConfig.AESKey)
	if err != nil {
		return nil, nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv[:IvLen]); err != nil {
		return nil, nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	finalEphID := make([]byte, EphIDLen)
	for i, v := range ephID {
		finalEphID[i] = ciphertext[aes.BlockSize+i] ^ v
	}
	return ciphertext[:IvLen], finalEphID, nil
}
