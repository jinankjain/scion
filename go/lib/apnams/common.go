package apnams

import (
	"github.com/scionproto/scion/go/lib/common"
)

const (
	SipHashKeySize = 16
	AESKeySize     = 16
	HMACKeySize    = 64
	PubkeyLen      = 32
	ApnaIDLen      = 16
)

func ProtocolStringToUint8(network string) (uint8, error) {
	switch network {
	case "udp4":
		return 0x01, nil
	default:
		return 0x00, common.NewBasicError("Unsupported protocol", nil)
	}
}
