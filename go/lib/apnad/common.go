package apnad

import (
	"github.com/scionproto/scion/go/lib/common"
)

const (
	TypeOffset      = 0
	TypeLen         = 1
	HostIDOffset    = 1
	HostIDLen       = 3
	TimestampOffset = 4
	TimestampLen    = 4
	EphIDLen        = 8
)

const (
	GenerateCtrlEphID    = 0x00
	GenerateSessionEphID = 0x01
)

type EphID [EphIDLen]byte

const (
	SipHashKeySize = 16
	AESKeySize     = 16
	HMACKeySize    = 64
	PubkeyLen      = 32
)

func ProtocolStringToUint8(network string) (uint8, error) {
	switch network {
	case "udp4":
		return 0x01, nil
	default:
		return 0x00, common.NewBasicError("Unsupported protocol", nil)
	}
}
