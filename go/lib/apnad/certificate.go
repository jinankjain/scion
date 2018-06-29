package apnad

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
)

const (
	EncryptEphidOffset      = 0
	EncryptEphidLen         = 16
	PubkeyOffset            = 16
	RecvOnlyOffset          = 32
	RecvOnlySize            = 1
	ExpTimeOffset           = 33
	SignatureOffset         = 37
	CertificateSignatureLen = 64
	CertificatePubKeyLen    = 32
	CertficateSize          = 117
)

const (
	ErrInsufficientBytes = "Insufficient bytes to construct a new certificate"
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

func (c *Certificate) RawCert() common.RawBytes {
	buf := c.Bytes()
	buf = append(buf, c.Signature...)
	return buf
}

func NewCertificateFromRawBytes(raw common.RawBytes) (*Certificate, error) {
	if len(raw) != CertficateSize {
		return nil, common.NewBasicError(ErrInsufficientBytes, nil)
	}
	cert := &Certificate{
		Ephid:     raw[EncryptEphidOffset:(EncryptEphidOffset + EncryptEphidLen)],
		Pubkey:    raw[PubkeyOffset:(PubkeyOffset + PubkeyLen)],
		RecvOnly:  raw[RecvOnlyOffset:(RecvOnlyOffset + RecvOnlySize)][0],
		ExpTime:   raw[ExpTimeOffset:(ExpTimeOffset + TimestampLen)],
		Signature: raw[SignatureOffset:],
	}
	return cert, nil
}

func (c *Certificate) String() string {
	return fmt.Sprintf("Ephid: %s, Pubkey: %s, RecvOnly: %x, ExpTime: %s, Signature: %s", c.Ephid,
		c.Pubkey, c.RecvOnly, c.ExpTime, c.Signature)
}
