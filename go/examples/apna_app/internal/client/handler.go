package client

import (
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/proto"
)

const (
	ErrUnexpectedPacket = "Unexpected Packet sequence"
)

func (c *Client) handshakePartOne() (common.RawBytes, error) {
	log.Info("Start APNA handshake with server by sending its pubkey")
	msg := &apna.Pkt{
		Which:       proto.APNAPkt_Which_pubkey,
		LocalEphID:  client.CtrlCertificate.Ephid,
		RemoteEphID: client.ServerCertificate.Ephid,
		NextHeader:  0x00,
		Pubkey:      client.CtrlCertificate.Pubkey,
	}
	err := msg.Sign(client.Config.HMACKey)
	if err != nil {
		return nil, err
	}
	ctrlSharedSecret, err := apnams.GenSharedSecret(c.ServerCertificate.Pubkey,
		c.CtrlEphIDPrivkey)
	if err != nil {
		return nil, err
	}
	c.Session = &Session{
		CtrlSharedSecret: ctrlSharedSecret,
	}
	return msg.RawPkt()
}

func (c *Client) handshakePartTwo(data *apna.Pkt) (common.RawBytes, error) {
	if data.NextHeader != 0x01 {
		return nil, common.NewBasicError(ErrUnexpectedPacket, nil, "expected",
			0x01, "got", data.NextHeader)
	}
	serverSessionCert, err := apnams.DecryptCert(c.Session.CtrlSharedSecret, data.Ecert)
	if err != nil {
		return nil, err
	}
	pub, priv, err := apnams.GenKeyPairs()
	if err != nil {
		return nil, err
	}
	c.Session.LocalPrivKey = priv
	clientSessionEphidRequest, err := c.ApnaMS.EphIDGenerationRequest(apna.SessionEphID,
		c.SrvAddr, pub)
	if err != nil {
		return nil, err
	}
	if clientSessionEphidRequest.ErrorCode != apnams.ErrorEphIDGenOk {
		return nil, common.NewBasicError(clientSessionEphidRequest.ErrorCode.String(), nil)
	}
	sessionSharedKey, err := apnams.GenSharedSecret(serverSessionCert.Pubkey, c.Session.LocalPrivKey)
	if err != nil {
		return nil, err
	}
	c.Session.SharedSecret = sessionSharedKey
	c.Session.RemoteEphID = data.LocalEphID
	ecert, err := apnams.EncryptCert(c.Session.CtrlSharedSecret, &clientSessionEphidRequest.Cert)
	if err != nil {
		return nil, err
	}
	reply := &apna.Pkt{
		Which:       proto.APNAPkt_Which_ecertPubkey,
		LocalEphID:  client.CtrlCertificate.Ephid,
		RemoteEphID: client.ServerCertificate.Ephid,
		NextHeader:  0x02,
		EcertPubkey: apna.EcertPubkey{
			Ecert:  ecert,
			Pubkey: serverSessionCert.Pubkey,
		},
	}
	err = reply.Sign(client.Config.HMACKey)
	if err != nil {
		return nil, err
	}
	return reply.RawPkt()
}
