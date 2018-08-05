package main

import (
	"net"

	"github.com/scionproto/scion/go/apna_srv/conf"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/snet"
)

type ApnaSrv struct {
	// Apna Configuration Directory
	Config *conf.Conf
	// MacVerification Queue
	MacQueue chan svcPkt
	// SvcForward Queue
	SvcForwardQueue chan svcPkt
	// SvcRecieve Queue
	SvcRecieveQueue chan common.RawBytes
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

func NewApnaSrv(id string, confDir string) (*ApnaSrv, error) {
	config, err := conf.Load(id, confDir)
	if err != nil {
		return nil, common.NewBasicError(ErrorConf, err)
	}
	a := &ApnaSrv{Config: config}
	if err := a.setup(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *ApnaSrv) Run() error {
	done := make(chan error)
	go a.StartServer(a.Config.PublicAddr, done)
	go a.StartSVC(a.Config.PublicAddr, a.Config.BindAddr, done)
	go a.SvcReceivePkt(done)
	err := <-done
	if err != nil {
		return err
	}
	return nil
}

type pkt = *apna.Pkt
type svcPkt = *apna.SVCPkt
