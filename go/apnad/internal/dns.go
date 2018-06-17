package internal

import (
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/log"
)

var dnsRegister map[uint8]map[string][]byte

func handleDNSRequest(req *apnad.DNSReq) *apnad.DNSReply {
	log.Debug("Got request", "request", req)
	val, ok := dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()]
	if !ok {
		reply := &apnad.DNSReply{
			ErrorCode: apnad.ErrorNoEntries,
		}
		log.Debug("Reply sent", "reply", reply)
		return reply
	}
	reply := &apnad.DNSReply{
		ErrorCode: apnad.ErrorDNSOk,
		Ephid:     val,
	}
	log.Debug("Reply sent", "reply", reply)
	return reply
}
