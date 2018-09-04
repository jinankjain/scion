package internal

import (
	"time"

	"github.com/scionproto/scion/go/lib/apnams"
)

var dnsRegister map[uint8]map[string]apnams.Certificate

func handleDNSRequest(req *apnams.DNSReq) *apnams.DNSReply {
	bench := &apnams.DNSRequestBenchmark{}
	start := time.Now()
	val, ok := dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()]
	if !ok {
		reply := &apnams.DNSReply{
			ErrorCode: apnams.ErrorNoEntries,
		}
		return reply
	}
	reply := &apnams.DNSReply{
		ErrorCode:   apnams.ErrorDNSOk,
		Certificate: val,
	}
	bench.RequestTime = time.Since(start)
	dnsRequestBenchmarks = append(dnsRequestBenchmarks, bench)
	return reply
}

func handleDNSRegister(req *apnams.DNSRegister) *apnams.DNSRegisterReply {
	bench := &apnams.DNSRegisterBenchmark{}
	start := time.Now()
	if _, ok := dnsRegister[req.Addr.Protocol]; ok {
		dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()] = req.Cert
	} else {
		dnsRegister[req.Addr.Protocol] = make(map[string]apnams.Certificate)
	}
	reply := &apnams.DNSRegisterReply{
		ErrorCode: apnams.ErrorDNSRegisterOk,
	}
	bench.RegisterTime = time.Since(start)
	dnsRegisterBenchmarks = append(dnsRegisterBenchmarks, bench)
	return reply
}
