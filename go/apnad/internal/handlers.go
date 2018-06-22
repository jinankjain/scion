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
	ephidReply := handleEphIDGeneration(&req)
	reply := &apnad.Pld{
		Id:                   pld.Id,
		Which:                proto.APNADMsg_Which_ephIDGenerationReply,
		EphIDGenerationReply: *ephidReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMsg reply", "err", err)
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}

type DNSReqHandler struct {
	Transport infra.Transport
}

func (h *DNSReqHandler) Handle(pld *apnad.Pld, src net.Addr) {
	req := pld.DNSReq
	dnsReply := handleDNSRequest(&req)
	reply := &apnad.Pld{
		Id:       pld.Id,
		Which:    proto.APNADMsg_Which_dNSReply,
		DNSReply: *dnsReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMsg reply")
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}

type DNSRegisterHandler struct {
	Transport infra.Transport
}

func (h *DNSRegisterHandler) Handle(pld *apnad.Pld, src net.Addr) {
	req := pld.DNSRegister
	dnsRegisterReply := handleDNSRegister(&req)
	reply := &apnad.Pld{
		Id:               pld.Id,
		Which:            proto.APNADMsg_Which_dNSRegisterReply,
		DNSRegisterReply: *dnsRegisterReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMsg reply")
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}
