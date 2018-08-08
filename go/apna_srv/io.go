package main

import (
	"fmt"
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
)

func (a *ApnaSrv) SvcReceivePkt(done chan error) {
	for {
		rpkt := <-a.SvcRecieveQueue
		pkt, err := apna.NewSVCPktFromRaw(rpkt)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		a.MacQueue <- pkt
	}
}

func (a *ApnaSrv) SvcForward() {
	for {
		p := <-a.SvcForwardQueue
		raddr := &snet.Addr{IA: addr.IAInt(p.RemoteIA).IA(), Host: addr.SvcAP}
		buf, err := p.ApnaPkt.RawPkt()
		if err != nil {
			log.Error(err.Error())
			continue
		}
		sz, err := a.SVCConn.WriteToSCION(buf, raddr)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if len(buf) != sz {
			log.Error("Pkt sending failed")
			continue
		}
	}
}

func (a *ApnaSrv) siphashToHost(hid common.RawBytes) (net.IP, error) {
	if val, ok := a.mapSiphashToHost[hid.String()]; ok {
		return val, nil
	}
	reply, err := a.ApnaMSConn.SiphashToHostRequest(hid)
	if err != nil {
		return nil, err
	}
	if reply.ErrorCode != apnams.ErrorSiphashToHostOk {
		return nil, common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	a.mapSiphashToHost[hid.String()] = reply.Host
	return reply.Host, nil
}

func (a *ApnaSrv) EndHostForward() {
	for {
		p := <-a.EndHostForwardQueue
		hid, err := apna.VerifyAndDecryptEphid(apna.EphID(p.ApnaPkt.RemoteEphID),
			a.Config.MSConf.AESKey, a.Config.MSConf.HMACKey)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		dstIP, err := a.siphashToHost(hid.Host())
		if err != nil {
			log.Error(err.Error())
			continue
		}
		dstAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%v", dstIP, p.ApnaPkt.RemotePort))
		if err != nil {
			log.Error(err.Error())
			continue
		}
		rawPkt, err := p.RawPkt()
		if err != nil {
			log.Error(err.Error())
			continue
		}
		sz, err := a.UDPConn.WriteTo(rawPkt, dstAddr)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if len(rawPkt) != sz {
			log.Error("Size mismatch", "expected", len(rawPkt), "got", sz)
		}
	}
}
