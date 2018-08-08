package internal

import (
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/proto"
)

type EphIDGenerationReqHandler struct {
	Transport infra.Transport
}

func (h *EphIDGenerationReqHandler) Handle(pld *apnams.Pld, src net.Addr) {
	req := pld.EphIDGenerationReq
	ephidReply := handleEphIDGeneration(&req)
	reply := &apnams.Pld{
		Id:                   pld.Id,
		Which:                proto.APNAMSMsg_Which_ephIDGenerationReply,
		EphIDGenerationReply: *ephidReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMsg reply", "err", err)
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}

type SiphashToHostHandler struct {
	Transport infra.Transport
}

func (h *SiphashToHostHandler) Handle(pld *apnams.Pld, src net.Addr) {
	req := pld.SiphashToHostReq
	siphashReply := handleSiphashToHost(&req)
	reply := &apnams.Pld{
		Id:                 pld.Id,
		Which:              proto.APNAMSMsg_Which_siphashToHostReply,
		SiphashToHostReply: *siphashReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize apnamsMsg reply", "err", err)
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}

type DNSReqHandler struct {
	Transport infra.Transport
}

func (h *DNSReqHandler) Handle(pld *apnams.Pld, src net.Addr) {
	req := pld.DNSReq
	dnsReply := handleDNSRequest(&req)
	reply := &apnams.Pld{
		Id:       pld.Id,
		Which:    proto.APNAMSMsg_Which_dNSReply,
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

func (h *DNSRegisterHandler) Handle(pld *apnams.Pld, src net.Addr) {
	req := pld.DNSRegister
	dnsRegisterReply := handleDNSRegister(&req)
	reply := &apnams.Pld{
		Id:               pld.Id,
		Which:            proto.APNAMSMsg_Which_dNSRegisterReply,
		DNSRegisterReply: *dnsRegisterReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMsg reply")
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}

type MacKeyRegisterHandler struct {
	Transport infra.Transport
}

func (h *MacKeyRegisterHandler) Handle(pld *apnams.Pld, src net.Addr) {
	req := pld.MacKeyRegister
	macKeyRegisterReply := handleMacKeyRegister(&req)
	reply := &apnams.Pld{
		Id:                  pld.Id,
		Which:               proto.APNAMSMsg_Which_macKeyRegisterReply,
		MacKeyRegisterReply: *macKeyRegisterReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMSMsg reply")
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}

type MacKeyRequestHandler struct {
	Transport infra.Transport
}

func (h *MacKeyRequestHandler) Handle(pld *apnams.Pld, src net.Addr) {
	req := pld.MacKeyReq
	macKeyReply := handleMacKeyRequest(&req)
	reply := &apnams.Pld{
		Id:          pld.Id,
		Which:       proto.APNAMSMsg_Which_macKeyReply,
		MacKeyReply: *macKeyReply,
	}
	b, err := proto.PackRoot(reply)
	if err != nil {
		log.Error("unable to serialize APNAMSMsg reply")
	}
	h.Transport.SendMsgTo(context.Background(), b, src)
}
