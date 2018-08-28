package internal

import (
	"fmt"
	"time"

	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
)

var macKeyRegister map[string]common.RawBytes

func handleMacKeyRequest(req *apnams.MacKeyReq) *apnams.MacKeyReply {
	bench := &apnams.MACRequestBenchmark{}
	start := time.Now()
	key := fmt.Sprintf("%s:%v", req.Addr, req.Port)
	val, ok := macKeyRegister[key]
	if !ok {
		reply := &apnams.MacKeyReply{
			ErrorCode: apnams.ErrorMacKeyNotFound,
		}
		return reply
	}
	reply := &apnams.MacKeyReply{
		ErrorCode: apnams.ErrorMacKeyOk,
		MacKey:    val,
	}
	bench.RequestTime = time.Since(start)
	macRequestBenchmarks = append(macRequestBenchmarks, bench)
	return reply
}

func handleMacKeyRegister(req *apnams.MacKeyRegister) *apnams.MacKeyRegisterReply {
	bench := &apnams.MACRegisterBenchmark{}
	start := time.Now()
	hid, err := generateHostID(req.Addr.To4())
	if err != nil {
		reply := &apnams.MacKeyRegisterReply{
			ErrorCode: apnams.ErrorMacKeyRegister,
		}
		return reply
	}
	key := fmt.Sprintf("%s:%v", hid, req.Port)
	macKeyRegister[key] = req.Key
	reply := &apnams.MacKeyRegisterReply{
		ErrorCode: apnams.ErrorMacKeyRegisterOk,
	}
	bench.RegisterTime = time.Since(start)
	macRegisterBenchmarks = append(macRegisterBenchmarks, bench)
	return reply
}
