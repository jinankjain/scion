package server

import (
	"github.com/scionproto/scion/go/lib/common"
)

// Session holds the state for a connection
type Session struct {
	LocalPrivKey        common.RawBytes
	SessionSharedSecret common.RawBytes
	CtrlSharedSecret    common.RawBytes
	LocalEphID          common.RawBytes
	RemoteEphID         common.RawBytes
}