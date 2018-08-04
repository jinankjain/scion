package main

import (
	"net"

	"github.com/scionproto/scion/go/apna_srv/conf"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
)

type ApnaSrv struct {
	// Apna Configuration
	Config *conf.Conf
	// MacVerification Queue
	MacQueue chan pkt
	// SvcForward Queue
	SvcForwardQueue chan pkt
	// SvcRecieve Queue
	SvcRecieveQueue chan pkt
	// EndHostForward Queue
	EndHostForwardQueue chan pkt
	// Failure Queue
	FailureQueue chan pkt
	// Endhost ReceiveQueue
	EndHostRecieveQueue chan pkt
	// HostToMacKey
	HostToMacKey map[string]common.RawBytes
	// UdpConn for service
	UDPConn *net.UDPConn
	// SVCConn for service
	SVCConn *snet.Conn
}

type pkt = *apna.Pkt

func (a *ApnaSrv) getKeyHost(hostID string) common.RawBytes {
	if key, ok := a.HostToMacKey[hostID]; ok {
		return key
	} else {
		// TODO(jinank): Contact Management Service to get the key
	}
	return nil
}

func (a *ApnaSrv) macVerification() {
	p := <-a.MacQueue
	log.Info("Mac verification Queue got", "pkt", p.String())
	a.SvcForwardQueue <- p
}
