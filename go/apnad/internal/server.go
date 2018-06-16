package internal

import (
	"fmt"
	"net"

	"github.com/scionproto/scion/go/lib/log"
)

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
	return nil
}
