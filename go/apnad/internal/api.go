package internal

import (
	"bytes"
	"context"
	"net"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/proto"
)

// API is a APAND API server running on top of a Transport
type API struct {
	Transport infra.Transport

	// State for request handlers
	handlers map[proto.APNADMsg_Which]handler
}

type handler interface {
	Handle(pld *apnad.Pld, src net.Addr)
}

func NewAPI(transport infra.Transport) *API {
	return &API{
		Transport: transport,
		handlers: map[proto.APNADMsg_Which]handler{
			proto.APNADMsg_Which_ephIDGenerationReq: &EphIDGenerationReqHandler{
				Transport: transport,
			},
			proto.APNADMsg_Which_dNSReq: &DNSReqHandler{
				Transport: transport,
			},
			proto.APNADMsg_Which_dNSRegister: &DNSRegisterHandler{
				Transport: transport,
			},
			proto.APNADMsg_Which_siphashToHostReq: &SiphashToHostHandler{
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
	p := &apnad.Pld{}
	if err := proto.ParseFromReader(p, proto.APNADMsg_TypeID, bytes.NewReader(b)); err != nil {
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
