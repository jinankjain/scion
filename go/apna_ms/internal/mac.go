package internal

import (
	"fmt"

	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

var macKeyRegister map[string]common.RawBytes

func handleMacKeyRequest(req *apnams.MacKeyReq) *apnams.MacKeyReply {
	log.Debug("Got MacKey request", "request", req)
	key := fmt.Sprintf("%s:%v", req.Addr, req.Port)
	val, ok := macKeyRegister[key]
	if !ok {
		reply := &apnams.MacKeyReply{
			ErrorCode: apnams.ErrorMacKeyNotFound,
		}
		log.Debug("Reply sent", "reply", reply)
		return reply
	}
	reply := &apnams.MacKeyReply{
		ErrorCode: apnams.ErrorMacKeyOk,
		MacKey:    val,
	}
	log.Debug("MAC Key request Reply sent", "reply", reply)
	return reply
}

func handleMacKeyRegister(req *apnams.MacKeyRegister) *apnams.MacKeyRegisterReply {
	hid, err := generateHostID(req.Addr.To4())
	if err != nil {
		reply := &apnams.MacKeyRegisterReply{
			ErrorCode: apnams.ErrorMacKeyRegister,
		}
		log.Debug("MacKey Register Reply sent", "reply", reply)
		return reply
	}
	log.Debug("Got MacKeyRegister request", "request", req, "hid", hid.String())
	key := fmt.Sprintf("%s:%v", hid, req.Port)
	macKeyRegister[key] = req.Key
	reply := &apnams.MacKeyRegisterReply{
		ErrorCode: apnams.ErrorMacKeyRegisterOk,
	}
	log.Debug("MacKey Register Reply sent", "reply", reply)
	return reply
}
