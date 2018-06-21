package apnad

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
