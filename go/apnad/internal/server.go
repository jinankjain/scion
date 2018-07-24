package internal

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/infra/transport"
	"github.com/scionproto/scion/go/lib/log"
)

func Init() {
	dnsRegister = make(map[uint8]map[string]apnad.Certificate)
	mapSiphashToHost = make(map[string]net.IP)
	siphashKey1 = binary.LittleEndian.Uint64(apnad.ApnadConfig.SipHashKey[:8])
	siphashKey2 = binary.LittleEndian.Uint64(apnad.ApnadConfig.SipHashKey[8:])
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
