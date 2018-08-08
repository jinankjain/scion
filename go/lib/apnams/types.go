package apnams

import (
	"fmt"
	"net"
	"strings"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

type MacKeyErrorCode uint8

const (
	ErrorMacKeyOk MacKeyErrorCode = iota
	ErrorMacKeyNotFound
)

func (c MacKeyErrorCode) String() string {
	switch c {
	case ErrorMacKeyOk:
		return "OK"
	case ErrorMacKeyNotFound:
		return "No mackey found"
	default:
		return fmt.Sprintf("Unknown error (%v)", uint8(c))
	}
}

type DNSErrorCode uint8

const (
	ErrorDNSOk DNSErrorCode = iota
	ErrorNoEntries
)

func (c DNSErrorCode) String() string {
	switch c {
	case ErrorDNSOk:
		return "OK"
	case ErrorNoEntries:
		return "No entries found"
	default:
		return fmt.Sprintf("Unknown error (%v)", uint8(c))
	}
}

type EphIDGenerationErrorCode uint8

const (
	ErrorEphIDGenOk EphIDGenerationErrorCode = iota
	ErrorGenerateHostID
	ErrorEncryptMACEphID
)

func (c EphIDGenerationErrorCode) String() string {
	switch c {
	case ErrorEphIDGenOk:
		return "OK"
	case ErrorGenerateHostID:
		return "Unable to generate HostID using siphash"
	case ErrorEncryptMACEphID:
		return "Error while encrypting and compute EphID"
	default:
		return fmt.Sprintf("Unknown error (%v)", uint8(c))
	}
}

type SiphashToHostErrorCode uint8

const (
	ErrorSiphashToHostOk SiphashToHostErrorCode = iota
	ErrorSiphashToHostNotFound
)

func (c SiphashToHostErrorCode) String() string {
	switch c {
	case ErrorSiphashToHostOk:
		return "OK"
	case ErrorSiphashToHostNotFound:
		return "Siphash Not Found"
	default:
		return fmt.Sprintf("Unknown error (%v)", uint8(c))
	}
}

type MacKeyRegisterErrorCode uint8

const (
	ErrorMacKeyRegisterOk MacKeyRegisterErrorCode = iota
	ErrorMacKeyRegister
)

func (c MacKeyRegisterErrorCode) String() string {
	switch c {
	case ErrorMacKeyRegisterOk:
		return "OK"
	case ErrorMacKeyRegister:
		return "Error while registering mac key"
	default:
		return fmt.Sprintf("Unknown error (%v)", uint8(c))
	}
}

type DNSRegisterErrorCode uint8

const (
	ErrorDNSRegisterOk DNSRegisterErrorCode = iota
	ErrorDNSRegister
)

func (c DNSRegisterErrorCode) String() string {
	switch c {
	case ErrorDNSRegisterOk:
		return "OK"
	case ErrorDNSRegister:
		return "Error while registering certificate"
	default:
		return fmt.Sprintf("Unknown error (%v)", uint8(c))
	}
}

var _ proto.Cerealizable = (*Pld)(nil)

type Pld struct {
	Id                   uint64
	Which                proto.APNAMSMsg_Which
	EphIDGenerationReq   EphIDGenerationReq
	EphIDGenerationReply EphIDGenerationReply
	DNSReq               DNSReq
	DNSReply             DNSReply
	DNSRegister          DNSRegister
	DNSRegisterReply     DNSRegisterReply
	SiphashToHostReq     SiphashToHostReq
	SiphashToHostReply   SiphashToHostReply
	MacKeyRegister       MacKeyRegister
	MacKeyRegisterReply  MacKeyRegisterReply
	MacKeyReq            MacKeyReq
	MacKeyReply          MacKeyReply
}

func NewPldFromRaw(b common.RawBytes) (*Pld, error) {
	p := &Pld{}
	return p, proto.ParseFromRaw(p, p.ProtoId(), b)
}

type SiphashToHostReq struct {
	Siphash common.RawBytes
}

type SiphashToHostReply struct {
	ErrorCode SiphashToHostErrorCode
	Host      net.IP
}

type EphIDGenerationReq struct {
	Kind   uint8
	Addr   ServiceAddr
	Pubkey common.RawBytes
}

type EphIDGenerationReply struct {
	ErrorCode EphIDGenerationErrorCode
	Cert      Certificate
}

type DNSReq struct {
	Addr ServiceAddr
}

type DNSReply struct {
	ErrorCode   DNSErrorCode
	Certificate Certificate
}

type DNSRegister struct {
	Addr ServiceAddr
	Cert Certificate
}

type DNSRegisterReply struct {
	ErrorCode DNSRegisterErrorCode
}

type ServiceAddr struct {
	Addr     net.IP
	Protocol uint8
}

type MacKeyReq struct {
	Addr common.RawBytes
	Port uint16
}

type MacKeyReply struct {
	ErrorCode MacKeyErrorCode
	MacKey    common.RawBytes
}

type MacKeyRegister struct {
	Addr net.IP
	Port uint16
	Key  common.RawBytes
}

type MacKeyRegisterReply struct {
	ErrorCode MacKeyRegisterErrorCode
}

func (p *Pld) ProtoId() proto.ProtoIdType {
	return proto.APNAMSMsg_TypeID
}

func (p *Pld) String() string {
	desc := []string{fmt.Sprintf("Apnad: Id: %d Union: ", p.Id)}
	u1, err := p.union()
	if err != nil {
		desc = append(desc, err.Error())
	} else {
		desc = append(desc, fmt.Sprintf("%+v", u1))
	}
	return strings.Join(desc, "")
}

func (p *Pld) union() (interface{}, error) {
	switch p.Which {
	case proto.APNAMSMsg_Which_ephIDGenerationReq:
		return p.EphIDGenerationReq, nil
	case proto.APNAMSMsg_Which_ephIDGenerationReply:
		return p.EphIDGenerationReply, nil
	case proto.APNAMSMsg_Which_dNSReq:
		return p.DNSReq, nil
	case proto.APNAMSMsg_Which_dNSReply:
		return p.DNSReply, nil
	case proto.APNAMSMsg_Which_dNSRegister:
		return p.DNSReq, nil
	case proto.APNAMSMsg_Which_dNSRegisterReply:
		return p.DNSReply, nil
	case proto.APNAMSMsg_Which_siphashToHostReq:
		return p.SiphashToHostReq, nil
	case proto.APNAMSMsg_Which_siphashToHostReply:
		return p.SiphashToHostReply, nil
	default:
		return nil, common.NewBasicError("Unsupported APNAD union type", nil, "type", p.Which)
	}
}

func (s *ServiceAddr) String() string {
	return fmt.Sprintf("Addr: % x, Protocol: %d", s.Addr, s.Protocol)
}

func (s *DNSReply) String() string {
	return fmt.Sprintf("ErrorCode %s, Ephid %v", s.ErrorCode, s.Certificate)
}

func (s *EphIDGenerationReply) String() string {
	return fmt.Sprintf("ErrorCode %s, Ephid %v", s.ErrorCode, s.Cert)
}

func (s *DNSRegisterReply) String() string {
	return fmt.Sprintf("ErrorCode %s", s.ErrorCode)
}

func (s *SiphashToHostReq) String() string {
	return fmt.Sprintf("Siphash %s", s.Siphash)
}
