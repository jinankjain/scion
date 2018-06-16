package internal

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"hash"
	"io"

	"github.com/scionproto/scion/go/lib/apnad"
)

const (
	ivPad  = 12
	ivLen  = 4
	macLen = 4
)

type ApnaCertificate struct {
	ephID              [apnad.EphIDLen]byte
	pubkey             [32]byte
	ctrlOrSessionEphId bool
	issuingAuthorityID []byte
	issuerEphID        []byte
	signature          []byte
}

var mac hash.Hash

func computeMac(iv, finalEphID []byte) ([]byte, error) {
	message := append(iv, finalEphID...)
	// TODO(jinankjain): Check bound on n
	_, err := mac.Write(message)
	if err != nil {
		return nil, err
	}
	expectedMAC := mac.Sum(nil)
	return expectedMAC[:macLen], nil
}

func verifyMac(message, msgMac []byte) bool {
	mac := hmac.New(sha256.New, apnad.ApnadConfig.HMACKey)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return bytes.Equal(expectedMAC[:macLen], msgMac)
}

func decryptEphID(iv, msg []byte) (*apnad.EphID, error) {
	block, err := aes.NewCipher(apnad.ApnadConfig.AESKey)
	if err != nil {
		return nil, err
	}
	padIV := make([]byte, ivPad)
	iv = append(iv, padIV...)
	plaintext := make([]byte, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	var ephID apnad.EphID
	for i, v := range ciphertext[aes.BlockSize:] {
		ephID[i] = msg[i] ^ v
	}
	return &ephID, nil
}

func encryptEphID(ephID *apnad.EphID) ([]byte, []byte, error) {
	plaintext := make([]byte, aes.BlockSize)
	block, err := aes.NewCipher(apnad.ApnadConfig.AESKey)
	if err != nil {
		return nil, nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv[:ivLen]); err != nil {
		return nil, nil, err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	finalEphID := make([]byte, apnad.EphIDLen)
	for i, v := range ephID {
		finalEphID[i] = ciphertext[aes.BlockSize+i] ^ v
	}
	return ciphertext[:ivLen], finalEphID, nil
}
