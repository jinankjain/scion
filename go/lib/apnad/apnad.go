package apnad

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra/disp"
	"github.com/scionproto/scion/go/lib/infra/transport"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/proto"
)

type Service interface {
	Connect() (Connector, error)
}

type service struct {
	ip   string
	port int
}

func NewService(ip string, port int) Service {
	return &service{
		ip:   ip,
		port: port,
	}
}

func (s *service) Connect() (Connector, error) {
	return connect(s.ip, s.port)
}

func connect(ip string, port int) (*connector, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", port))
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		return nil, err
	}
	return &connector{
		dispatcher: disp.New(
			transport.NewPacketTransport(conn),
			&Adapter{},
			log.Root(),
		),
		addr: addr,
	}, nil
}

type Connector interface {
	EphIDGenerationRequest(kind byte,
		addr *ServiceAddr,
		pubkey common.RawBytes,
	) (*EphIDGenerationReply, error)
	DNSRequest(addr *ServiceAddr) (*DNSReply, error)
	DNSRegister(addr *ServiceAddr,
		cert Certificate) (*DNSRegisterReply, error)
	SiphashToHostRequest(siphash common.RawBytes) (*SiphashToHostReply, error)
}

type connector struct {
	dispatcher *disp.Dispatcher
	requestID  uint64
	addr       net.Addr
}

func (c *connector) nextID() uint64 {
	return atomic.AddUint64(&c.requestID, 1)
}

func (c *connector) EphIDGenerationRequest(kind byte,
	addr *ServiceAddr,
	pubkey common.RawBytes,
) (*EphIDGenerationReply, error) {
	reply, err := c.dispatcher.Request(
		context.Background(),
		&Pld{
			Id:    c.nextID(),
			Which: proto.APNADMsg_Which_ephIDGenerationReq,
			EphIDGenerationReq: EphIDGenerationReq{
				Kind:   uint8(kind),
				Addr:   *addr,
				Pubkey: pubkey,
			},
		},
		c.addr,
	)
	if err != nil {
		return nil, err
	}
	return &reply.(*Pld).EphIDGenerationReply, nil
}

func (c *connector) DNSRequest(addr *ServiceAddr) (*DNSReply, error) {
	reply, err := c.dispatcher.Request(
		context.Background(),
		&Pld{
			Id:    c.nextID(),
			Which: proto.APNADMsg_Which_dNSReq,
			DNSReq: DNSReq{
				Addr: *addr,
			},
		},
		c.addr,
	)
	if err != nil {
		return nil, err
	}
	return &reply.(*Pld).DNSReply, nil
}

func (c *connector) DNSRegister(addr *ServiceAddr,
	cert Certificate) (*DNSRegisterReply, error) {
	reply, err := c.dispatcher.Request(
		context.Background(),
		&Pld{
			Id:    c.nextID(),
			Which: proto.APNADMsg_Which_dNSRegister,
			DNSRegister: DNSRegister{
				Addr: *addr,
				Cert: cert,
			},
		},
		c.addr,
	)
	if err != nil {
		return nil, err
	}
	return &reply.(*Pld).DNSRegisterReply, nil
}

func (c *connector) SiphashToHostRequest(siphash common.RawBytes) (*SiphashToHostReply, error) {
	reply, err := c.dispatcher.Request(
		context.Background(),
		&Pld{
			Id:    c.nextID(),
			Which: proto.APNADMsg_Which_siphashToHostReq,
			SiphashToHostReq: SiphashToHostReq{
				Siphash: siphash,
			},
		},
		c.addr,
	)
	if err != nil {
		return nil, err
	}
	return &reply.(*Pld).SiphashToHostReply, nil
}
