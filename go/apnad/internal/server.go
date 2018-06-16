package internal

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net"

	"github.com/dchest/siphash"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/infra/transport"
	"github.com/scionproto/scion/go/lib/log"
)

func Init() {
	mac = hmac.New(sha256.New, apnad.ApnadConfig.HMACKey)
	siphasher = siphash.New(apnad.ApnadConfig.SipHashKey)
}

func ListenAndServe(ip net.IP, port int) error {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%v", ip, port))
	if err != nil {
		return err
	}
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		return err
	}
	log.Info("Started APNAD Service", "addr", serverConn.LocalAddr())
	NewAPI(transport.NewPacketTransport(serverConn)).Serve()
	return nil
}
