package internal

import (
	"fmt"
	"time"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/log"
)

var dnsRegister map[uint8]map[string]apnad.Certificate

func handleDNSRequest(req *apnad.DNSReq) *apnad.DNSReply {
	b := &apnad.DNSReplyBenchmark{}

	t := time.Now()
	name := fmt.Sprintf("%s-%s", "ephid_benchmark", t.Format("2006-01-02 15:04:05"))
	log.SetupLogFile(name, logDir, "info", 20, 100, 0)

	start := time.Now()
	val, ok := dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()]
	if !ok {
		reply := &apnad.DNSReply{
			ErrorCode: apnad.ErrorNoEntries,
		}
		return reply
	}
	reply := &apnad.DNSReply{
		ErrorCode:   apnad.ErrorDNSOk,
		Certificate: val,
	}
	b.ReplyTime = time.Since(start)
	log.Info(b.String())
	return reply
}

func handleDNSRegister(req *apnad.DNSRegister) *apnad.DNSRegisterReply {
	b := &apnad.DNSRegisterBenchmark{}

	t := time.Now()
	name := fmt.Sprintf("%s-%s", "ephid_benchmark", t.Format("2006-01-02 15:04:05"))
	log.SetupLogFile(name, logDir, "info", 20, 100, 0)

	start := time.Now()
	dnsRegister[req.Addr.Protocol] = make(map[string]apnad.Certificate)
	dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()] = req.Cert
	reply := &apnad.DNSRegisterReply{
		ErrorCode: apnad.ErrorDNSRegisterOk,
	}
	b.RegisterTime = time.Since(start)
	log.Info(b.String())
	return reply
}
