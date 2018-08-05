package internal

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/infra/transport"
	"github.com/scionproto/scion/go/lib/log"
)

func Init() {
	dnsRegister = make(map[uint8]map[string]apnams.Certificate)
	mapSiphashToHost = make(map[string]net.IP)
	siphashKey1 = binary.LittleEndian.Uint64(apnams.ApnaMSConfig.SipHashKey[:8])
	siphashKey2 = binary.LittleEndian.Uint64(apnams.ApnaMSConfig.SipHashKey[8:])
}

func ListenAndServe(ip net.IP, port int) error {
	serverAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%v", ip, port))
	if err != nil {
		return err
	}
	serverConn, err := net.ListenUDP("udp4", serverAddr)
	if err != nil {
		return err
	}
	log.Info("Started APNA Management Service", "addr", serverConn.LocalAddr())
	NewAPI(transport.NewPacketTransport(serverConn)).Serve()
	return nil
}
