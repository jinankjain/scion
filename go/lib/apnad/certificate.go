package apnad

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
)

const (
	EncryptEphidLen      = 16
	CertificateLen       = 64
	CertificatePubKeyLen = 32
)

type Certificate struct {
	Ephid     common.RawBytes
	Pubkey    common.RawBytes
	RecvOnly  uint8
	ExpTime   common.RawBytes
	Signature common.RawBytes
}

func (c *Certificate) Bytes() common.RawBytes {
	var buf common.RawBytes
	buf = append(buf, c.Ephid...)
	buf = append(buf, c.Pubkey...)
	buf = append(buf, byte(c.RecvOnly))
	buf = append(buf, c.ExpTime...)
	return buf
}

func (c *Certificate) Sign() error {
	sign, err := crypto.Sign(c.Bytes(), ApnadConfig.Privkey, crypto.Ed25519)
	if err != nil {
		return err
	}
	c.Signature = sign
	return nil
}

func (c *Certificate) Verify() error {
	return crypto.Verify(c.Bytes(), c.Signature, ApnadConfig.Pubkey, crypto.Ed25519)
}
