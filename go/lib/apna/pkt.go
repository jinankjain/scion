package apna

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/proto"
)

var _ proto.Cerealizable = (*Pkt)(nil)

type Pkt struct {
	Which       proto.APNAPkt_Which
	LocalEphID  common.RawBytes
	RemoteEphID common.RawBytes
	RemotePort  uint16
	LocalPort   uint16
	PacketMAC   common.RawBytes
	NextHeader  uint8
	Pubkey      common.RawBytes
	Ecert       common.RawBytes
	Data        common.RawBytes
	EcertPubkey EcertPubkey
}

type EcertPubkey struct {
	Ecert  common.RawBytes
	Pubkey common.RawBytes
}

func (p *Pkt) ProtoId() proto.ProtoIdType {
	return proto.APNAPkt_TypeID
}

func (p *Pkt) union() (interface{}, error) {
	switch p.Which {
	case proto.APNAPkt_Which_pubkey:
		return p.Pubkey, nil
	case proto.APNAPkt_Which_ecert:
		return p.Ecert, nil
	case proto.APNAPkt_Which_data:
		return p.Data, nil
	case proto.APNAPkt_Which_ecertPubkey:
		return p.EcertPubkey, nil
	default:
		return nil, common.NewBasicError("Unsupported APNA union type", nil, "type", p.Which)
	}
}

func (p *Pkt) String() string {
	desc := []string{
		fmt.Sprintf("LocalEphID: %s, RemoteEphID: %s, LocalPort %d, RemotePort %d, MAC: %s, NextHeader: %d Union: ",
			p.LocalEphID, p.RemoteEphID, p.LocalPort, p.RemotePort, p.PacketMAC, p.NextHeader),
	}
	u1, err := p.union()
	if err != nil {
		desc = append(desc, err.Error())
	} else {
		desc = append(desc, fmt.Sprintf("%+v", u1))
	}
	return strings.Join(desc, "")
}

func (p *Pkt) RawPkt() (common.RawBytes, error) {
	return proto.PackRoot(p)
}

func NewPktFromRaw(b common.RawBytes) (*Pkt, error) {
	p := &Pkt{}
	return p, proto.ParseFromRaw(p, p.ProtoId(), b)
}

func (p *Pkt) bytes() common.RawBytes {
	var buf common.RawBytes
	switch p.Which {
	case proto.APNAPkt_Which_pubkey:
		buf = append(buf, p.Pubkey...)
	case proto.APNAPkt_Which_ecert:
		buf = append(buf, p.Ecert...)
	case proto.APNAPkt_Which_data:
		buf = append(buf, p.Data...)
	case proto.APNAPkt_Which_ecertPubkey:
		buf = append(buf, p.EcertPubkey.Ecert...)
		buf = append(buf, p.EcertPubkey.Pubkey...)
	}
	return buf
}

func (p *Pkt) Sign(key common.RawBytes) error {
	mac := hmac.New(sha256.New, key)
	msg := p.bytes()
	_, err := mac.Write(msg)
	if err != nil {
		return err
	}
	p.PacketMAC = mac.Sum(nil)
	return nil
}

func (p *Pkt) Verify(key common.RawBytes) (bool, error) {
	mac := hmac.New(sha256.New, key)
	msg := p.bytes()
	_, err := mac.Write(msg)
	if err != nil {
		return false, err
	}
	expectedMac := mac.Sum(nil)
	hmac.Equal(expectedMac, p.PacketMAC)
	return true, nil
}
