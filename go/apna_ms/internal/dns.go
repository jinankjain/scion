package internal

import (
	"time"

	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/log"
)

var dnsRegister map[uint8]map[string]apnams.Certificate

func handleDNSRequest(req *apnams.DNSReq) *apnams.DNSReply {
	log.Debug("Got DNS request", "request", req)
	bench := &apnams.DNSRequestBenchmark{}
	start := time.Now()
	val, ok := dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()]
	if !ok {
		reply := &apnams.DNSReply{
			ErrorCode: apnams.ErrorNoEntries,
		}
		log.Debug("Reply sent", "reply", reply)
		return reply
	}
	reply := &apnams.DNSReply{
		ErrorCode:   apnams.ErrorDNSOk,
		Certificate: val,
	}
	log.Debug("DNS Request Reply sent", "reply", reply)
	bench.RequestTime = time.Since(start)
	dnsRequestBenchmarks = append(dnsRequestBenchmarks, bench)
	return reply
}

func handleDNSRegister(req *apnams.DNSRegister) *apnams.DNSRegisterReply {
	log.Debug("Got DNSRegister request", "request", req)
	bench := &apnams.DNSRegisterBenchmark{}
	start := time.Now()
	dnsRegister[req.Addr.Protocol] = make(map[string]apnams.Certificate)
	dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()] = req.Cert
	reply := &apnams.DNSRegisterReply{
		ErrorCode: apnams.ErrorDNSRegisterOk,
	}
	log.Debug("DNS Register Reply sent", "reply", reply)
	bench.RegisterTime = time.Since(start)
	dnsRegisterBenchmarks = append(dnsRegisterBenchmarks, bench)
	return reply
}
