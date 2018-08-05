package apnams

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
)

// EncryptCert encrypts Session Certificate for APNA
func EncryptCert(sharedKey common.RawBytes, cert *Certificate) (common.RawBytes, error) {
	rawData := cert.RawCert()
	return crypto.Encrypt(sharedKey, rawData, crypto.Curve25519xSalsa20Poly1305)
}

// DecryptCert decrypts encrypted Session Certificate for APNA
func DecryptCert(sharedKey common.RawBytes, ecert common.RawBytes) (*Certificate, error) {
	rawCert, err := crypto.Decrypt(sharedKey, ecert, crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return nil, err
	}
	return NewCertificateFromRawBytes(rawCert)
}

// EncryptData encrypts data associated with ApnaPayload
func EncryptData(sharedKey common.RawBytes, data common.RawBytes) (common.RawBytes, error) {
	return crypto.Encrypt(sharedKey, data, crypto.Curve25519xSalsa20Poly1305)
}

// DecryptData decrypts data associated with ApnaPayload
func DecryptData(sharedKey common.RawBytes, edata common.RawBytes) (common.RawBytes, error) {
	return crypto.Decrypt(sharedKey, edata, crypto.Curve25519xSalsa20Poly1305)
}
