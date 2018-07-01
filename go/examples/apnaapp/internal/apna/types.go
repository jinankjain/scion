package apna

import (
	"fmt"
	"strings"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*Pld)(nil)

type Pld struct {
	Which       proto.APNAHeader_Which
	LocalEphID  common.RawBytes
	RemoteEphID common.RawBytes
	NextHeader  uint8
	Pubkey      common.RawBytes
	Ecert       common.RawBytes
	Data        common.RawBytes
	EcertPubkey EcertPubkey
}

func (p *Pld) ProtoId() proto.ProtoIdType {
	return proto.APNAHeader_TypeID
}

type EcertPubkey struct {
	Ecert  common.RawBytes
	Pubkey common.RawBytes
}

func (p *Pld) String() string {
	desc := []string{
		fmt.Sprintf("LocalEphID: %s, RemoteEphID: %s, NextHeader: %d Union: ",
			p.LocalEphID, p.RemoteEphID, p.NextHeader),
	}
	u1, err := p.union()
	if err != nil {
		desc = append(desc, err.Error())
	} else {
		desc = append(desc, fmt.Sprintf("%+v", u1))
	}
	return strings.Join(desc, "")
}

func (p *Pld) RawPld() (common.RawBytes, error) {
	return proto.PackRoot(p)
}

func (p *Pld) union() (interface{}, error) {
	switch p.Which {
	case proto.APNAHeader_Which_pubkey:
		return p.Pubkey, nil
	case proto.APNAHeader_Which_ecert:
		return p.Ecert, nil
	case proto.APNAHeader_Which_data:
		return p.Data, nil
	case proto.APNAHeader_Which_ecertPubkey:
		return p.EcertPubkey, nil
	default:
		return nil, common.NewBasicError("Unsupported APNA union type", nil, "type", p.Which)
	}
}

func NewPldFromRaw(b common.RawBytes) (*Pld, error) {
	p := &Pld{}
	return p, proto.ParseFromRaw(p, p.ProtoId(), b)
}

func (e *EcertPubkey) String() string {
	return fmt.Sprintf("Ecert: %s, Pubkey: %s", e.Ecert, e.Pubkey)
}
