package internal

import (
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/proto"
)

type EphIDGenerationReqHandler struct {
	Transport infra.Transport
}

func (h *EphIDGenerationReqHandler) Handle(pld *apnad.Pld, src net.Addr) {
	req := pld.EphIDGenerationReq
	log.Info("Got request", "req", req)
	reply := &apnad.Pld{
		Id:    pld.Id,
		Which: proto.APNADMsg_Which_ephIDGenerationReply,
		EphIDGenerationReply: apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorEncryptEphID,
		},
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMsg reply")
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}

type DNSReqHandler struct {
	Transport infra.Transport
}

func (h *DNSReqHandler) Handle(pld *apnad.Pld, src net.Addr) {
	req := pld.DNSReq
	log.Info("Got Request", "req", req)
	reply := &apnad.Pld{
		Id:    pld.Id,
		Which: proto.APNADMsg_Which_dNSReply,
		DNSReply: apnad.DNSReply{
			ErrorCode: apnad.ErrorNoEntries,
		},
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMsg reply")
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}
