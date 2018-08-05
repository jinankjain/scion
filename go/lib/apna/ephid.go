package apna

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"

	"github.com/scionproto/scion/go/lib/common"
)

type HID common.RawBytes

const (
	CtrlEphID    = 0x0
	SessionEphID = 0x1
)

const (
	IvLen      = 4
	HIDLen     = 8
	MacLen     = 4
	HostOffset = 1
	HostLen    = 3
	TimeOffset = 4
	TimeLen    = 4
	MacOffset  = 12
	TypeOffset = 0
	TypeLen    = 1
	EHIDOffset = 4
	EHIDLen    = 8
)

func (h HID) Type() byte {
	return h[TypeOffset]
}

func (h HID) Host() common.RawBytes {
	return common.RawBytes(h[HostOffset : HostOffset+HostLen])
}

func (h HID) ExpTime() uint32 {
	return binary.LittleEndian.Uint32(h[TimeOffset : TimeOffset+TimeLen])
}

func GetHID(kind byte, host common.RawBytes, expTime common.RawBytes) HID {
	hid := make([]byte, 8)
	hid[TypeOffset] = kind
	copy(hid[HostOffset:HostOffset+HostLen], host)
	copy(hid[TimeOffset:TimeOffset+TimeLen], expTime)
	return hid
}

type EphID common.RawBytes

func (e EphID) IV() common.RawBytes {
	return common.RawBytes(e[:IvLen])
}

func (e EphID) MAC() common.RawBytes {
	return common.RawBytes(e[MacOffset : MacOffset+MacLen])
}

func (e EphID) EHID() common.RawBytes {
	return common.RawBytes(e[EHIDOffset : EHIDOffset+EHIDLen])
}

const (
	ErrMacVerificationFailed = "MAC verification failed"
)

func VerifyAndDecryptEphid(e EphID, AESKey common.RawBytes, HMACKey common.RawBytes) (HID, error) {
	verify, err := verifyMac(e, HMACKey)
	if err != nil {
		return nil, err
	}
	if !verify {
		return nil, common.NewBasicError(ErrMacVerificationFailed, nil)
	}
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	copy(iv[:IvLen], e[:IvLen])
	plaintext := make([]byte, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	hid := make([]byte, HIDLen)
	for i := 0; i < HIDLen; i++ {
		hid[i] = ciphertext[aes.BlockSize+i] ^ e[i+IvLen]
	}
	return hid, nil
}

func verifyMac(e EphID, HMACKey common.RawBytes) (bool, error) {
	mac := hmac.New(sha256.New, HMACKey)
	_, err := mac.Write(e[:MacOffset])
	if err != nil {
		return false, err
	}
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(expectedMAC[:MacLen], e[MacOffset:]), nil
}

func computeMac(e EphID, HMACKey common.RawBytes) error {
	mac := hmac.New(sha256.New, HMACKey)
	_, err := mac.Write(e[:MacOffset])
	if err != nil {
		return err
	}
	expectedMac := mac.Sum(nil)
	copy(e[MacOffset:], expectedMac[:MacLen])
	return nil
}

func EncryptAndSignHostID(h HID, AESKey common.RawBytes, HMACKey common.RawBytes) (EphID, error) {
	plaintext := make([]byte, aes.BlockSize)
	block, err := aes.NewCipher(AESKey)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv[:IvLen]); err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	ephID := make([]byte, IvLen+HIDLen+MacLen)
	copy(ephID[:IvLen], iv[:IvLen])
	for i, v := range h {
		ephID[i+IvLen] = ciphertext[aes.BlockSize+i] ^ v
	}
	err = computeMac(ephID, HMACKey)
	if err != nil {
		return nil, err
	}
	return ephID, nil
}
