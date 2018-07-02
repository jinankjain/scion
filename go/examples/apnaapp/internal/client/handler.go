package client

import (
	"github.com/scionproto/scion/go/examples/apnaapp/internal/apna"
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/proto"
)

const (
	ErrUnexpectedPacket = "Unexpected Packet sequence"
)

func (c *Client) handshakePartOne() (common.RawBytes, error) {
	log.Info("Start APNA handshake with server by sending its pubkey")
	msg := &apna.Pld{
		Which:       proto.APNAHeader_Which_pubkey,
		LocalEphID:  client.CtrlCertificate.Ephid,
		RemoteEphID: client.ServerCertificate.Ephid,
		NextHeader:  0x00,
		Pubkey:      client.CtrlCertificate.Pubkey,
	}
	ctrlSharedSecret, err := apnad.GenSharedSecret(c.ServerCertificate.Pubkey,
		c.CtrlEphIDPrivkey)
	if err != nil {
		return nil, err
	}
	c.Session = &Session{
		CtrlSharedSecret: ctrlSharedSecret,
	}
	return msg.RawPld()
}

func (c *Client) handshakePartTwo(data *apna.Pld) (common.RawBytes, error) {
	if data.NextHeader != 0x01 {
		return nil, common.NewBasicError(ErrUnexpectedPacket, nil, "expected",
			0x01, "got", data.NextHeader)
	}
	serverSessionCert, err := apna.DecryptCert(c.Session.CtrlSharedSecret, data.Ecert)
	if err != nil {
		return nil, err
	}
	pub, priv, err := apnad.GenKeyPairs()
	if err != nil {
		return nil, err
	}
	c.Session.LocalPrivKey = priv
	clientSessionEphidRequest, err := c.Apnad.EphIDGenerationRequest(apnad.GenerateSessionEphID,
		c.SrvAddr, pub)
	if err != nil {
		return nil, err
	}
	if clientSessionEphidRequest.ErrorCode != apnad.ErrorEphIDGenOk {
		return nil, common.NewBasicError(clientSessionEphidRequest.ErrorCode.String(), nil)
	}
	sessionSharedKey, err := apnad.GenSharedSecret(serverSessionCert.Pubkey, c.Session.LocalPrivKey)
	if err != nil {
		return nil, err
	}
	c.Session.SharedSecret = sessionSharedKey
	c.Session.RemoteEphID = data.LocalEphID
	ecert, err := apna.EncryptCert(c.Session.CtrlSharedSecret, &clientSessionEphidRequest.Cert)
	if err != nil {
		return nil, err
	}
	reply := &apna.Pld{
		Which:       proto.APNAHeader_Which_ecertPubkey,
		LocalEphID:  client.CtrlCertificate.Ephid,
		RemoteEphID: client.ServerCertificate.Ephid,
		NextHeader:  0x02,
		EcertPubkey: apna.EcertPubkey{
			Ecert:  ecert,
			Pubkey: serverSessionCert.Pubkey,
		},
	}
	return reply.RawPld()
}
