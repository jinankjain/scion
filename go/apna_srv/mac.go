package main

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

func (a *ApnaSrv) getKeyHost(hostID string) common.RawBytes {
	if key, ok := a.HostToMacKey[hostID]; ok {
		return key
	} else {
		// TODO(jinank): Contact Management Service to get the key
	}
	return nil
}

func (a *ApnaSrv) MacVerification() {
	for {
		p := <-a.MacQueue
		log.Info("Mac verification Queue got", "pkt", p.String())
		a.SvcForwardQueue <- p
	}
}
