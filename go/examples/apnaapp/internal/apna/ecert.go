package apna

import (
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
)

type Ecert struct {
	Cert      *apnad.Certificate
	SharedKey common.RawBytes
}

func (e *Ecert) Encrypt() (common.RawBytes, error) {
	rawData := e.Cert.RawCert()
	return crypto.Encrypt(e.SharedKey, rawData, crypto.Curve25519xSalsa20Poly1305)
}

func (e *Ecert) Decrypt(edata common.RawBytes) (*apnad.Certificate, error) {
	uedata, err := crypto.Decrypt(e.SharedKey, edata, crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return nil, err
	}
	return apnad.NewCertificateFromRawBytes(uedata)
}
