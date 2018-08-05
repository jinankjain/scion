package internal

import (
	"bytes"
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/proto"
)

// API is a APAND API server running on top of a Transport
type API struct {
	Transport infra.Transport

	// State for request handlers
	handlers map[proto.APNAMSMsg_Which]handler
}

type handler interface {
	Handle(pld *apnams.Pld, src net.Addr)
}

func NewAPI(transport infra.Transport) *API {
	return &API{
		Transport: transport,
		handlers: map[proto.APNAMSMsg_Which]handler{
			proto.APNAMSMsg_Which_ephIDGenerationReq: &EphIDGenerationReqHandler{
				Transport: transport,
			},
			proto.APNAMSMsg_Which_dNSReq: &DNSReqHandler{
				Transport: transport,
			},
			proto.APNAMSMsg_Which_dNSRegister: &DNSRegisterHandler{
				Transport: transport,
			},
			proto.APNAMSMsg_Which_siphashToHostReq: &SiphashToHostHandler{
				Transport: transport,
			},
		},
	}
}

func (srv *API) Serve() error {
	for {
		b, addr, err := srv.Transport.RecvFrom(context.Background())
		if err != nil {
			return err
		}
		srv.Handle(b, addr)
	}
}

func (srv *API) Handle(b common.RawBytes, addr net.Addr) {
	p := &apnams.Pld{}
	if err := proto.ParseFromReader(p, proto.APNAMSMsg_TypeID, bytes.NewReader(b)); err != nil {
		log.Error("capnp error", "err", err)
		return
	}
	handler, ok := srv.handlers[p.Which]
	if !ok {
		log.Error("handler not found for capnp message", "which", p.Which)
		return
	}
	handler.Handle(p, addr)
}
