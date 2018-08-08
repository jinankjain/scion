package main

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

func (a *ApnaSrv) getKeyHost(hostID common.RawBytes, port uint16) common.RawBytes {
	key := fmt.Sprintf("%s:%v", hostID, port)
	val, ok := a.HostToMacKey[key]
	if ok {
		return val
	}
	// TODO(jinank): Contact Management Service to get the key
	reply, err := a.ApnaMSConn.MacKeyRequest(hostID, port)
	if err != nil {
		log.Error(err.Error())
		return nil
	} else if reply.ErrorCode != apnams.ErrorMacKeyOk {
		log.Error(reply.ErrorCode.String())
		return nil
	} else {
		a.HostToMacKey[key] = reply.MacKey
		return reply.MacKey
	}
}

func (a *ApnaSrv) MacVerification() {
	for {
		p := <-a.MacQueue
		hid, err := apna.VerifyAndDecryptEphid(apna.EphID(p.ApnaPkt.LocalEphID),
			a.Config.MSConf.AESKey, a.Config.MSConf.HMACKey)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		macKey := a.getKeyHost(hid.Host(), p.ApnaPkt.LocalPort)
		match, err := p.ApnaPkt.Verify(macKey)
		if err != nil {
			log.Error(err.Error())
			continue
		}
		if !match {
			log.Error("Mac verification failed")
			continue
		}
		a.SvcForwardQueue <- p
	}
}
