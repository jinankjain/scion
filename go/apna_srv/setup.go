package main

import (
	"fmt"
	"net"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
)

const (
	initAttempts  = 100
	initInterval  = time.Second
	ErrorConf     = "Unable to load configuration"
	ErrorDispInit = "Unable to initialize dispatcher"
	ErrorSNET     = "Unable to create local SCION Network context"
	ErrorApnaMS   = "Unable to connect to apna management service"
	MaxBufSize    = 2 << 12
	ApnaMSIP      = "127.0.0.1"
	ApnaMSPort    = 6000
)

func (a *ApnaSrv) setup() error {
	a.MacQueue = make(chan svcPkt, 16)
	a.SvcForwardQueue = make(chan svcPkt, 16)
	a.SvcRecieveQueue = make(chan common.RawBytes, 16)
	a.EndHostForwardQueue = make(chan svcPkt, 16)
	a.FailureQueue = make(chan pkt, 16)
	a.EndHostRecieveQueue = make(chan pkt, 16)
	a.HostToMacKey = make(map[string]common.RawBytes)
	a.mapSiphashToHost = make(map[string]net.IP)
	if err := initSNET(a.Config.PublicAddr.IA, initAttempts, initInterval); err != nil {
		return common.NewBasicError(ErrorSNET, err)
	}
	con, err := initApnaMS()
	if err != nil {
		return common.NewBasicError(ErrorApnaMS, err)
	}
	a.ApnaMSConn = con
	return nil
}

func initApnaMS() (apnams.Connector, error) {
	service := apnams.NewService(ApnaMSIP, ApnaMSPort)
	return service.Connect()
}

func initSNET(ia addr.IA, attempts int, sleep time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		if err = snet.Init(ia, *sciondPath, *dispPath); err == nil {
			break
		}
		log.Error("Unable to initialize snet", "Retry interval", sleep, "err", err)
		time.Sleep(sleep)
	}
	return err
}

func (a *ApnaSrv) StartSVC(pubAddr, bindAddr *snet.Addr, done chan error) {
	copyPubAddr := pubAddr.Copy()
	copyPubAddr.L4Port += 1
	conn, err := snet.ListenSCIONWithBindSVC("udp4", copyPubAddr, bindAddr,
		addr.SvcAP)
	for err != nil {
		log.Error(err.Error())
		pubAddr.L4Port += 1
		conn, err = snet.ListenSCIONWithBindSVC("udp4", copyPubAddr, bindAddr,
			addr.SvcAP)
	}
	a.SVCConn = conn
	log.Info("Started APNA service on", "addr", copyPubAddr.String())
	buf := make([]byte, MaxBufSize)
	for {
		_, addr, err := conn.ReadFromSCION(buf)
		pkt, err := apna.NewPktFromRaw(buf)
		if err != nil {
			log.Info("Pkt parsing failed!!!", "err", err)
			continue
		}
		if err != nil {
			log.Error("Unable to read packet from the network", "err", err)
			continue
		}
		// Create the SVCPkt here
		sPkt := &apna.SVCPkt{
			RemoteIA: addr.IA.IAInt(),
			ApnaPkt:  *pkt,
		}
		// Read will read the pkt and send the raddr to guy
		a.EndHostForwardQueue <- sPkt
	}
}

func (a *ApnaSrv) StartServer(addr *snet.Addr, done chan error) {
	laddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%v", addr.Host.IP(), addr.L4Port))
	if err != nil {
		log.Error(err.Error())
		done <- err
	}
	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		log.Error(err.Error())
		done <- err
	}
	a.UDPConn = conn
	log.Info("Started UDP server on", "addr", laddr.String())
	buf := make([]byte, MaxBufSize)
	for {
		len, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Error("Unable to read network packet", "err", err)
			continue
		}
		tmp := make([]byte, len)
		copy(tmp, buf[:len])
		a.SvcRecieveQueue <- tmp
	}
}
