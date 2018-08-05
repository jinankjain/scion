package main

import (
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/apna"
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
		currLen := 0
		size := len(buf)
		for currLen != size {
			len, err := a.SVCConn.WriteToSCION(buf[0:], raddr)
			if err != nil {
				log.Error(err.Error())
				continue
			}
			currLen += len
		}
	}
}
