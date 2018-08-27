package server

import (
	"github.com/scionproto/scion/go/examples/apna_app/internal/config"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/proto"
)

func (s Server) handshakePartOne(data *apna.Pkt, raddr *snet.Addr) {
	log.Info("First part of the handshake i.e. got a request to initiate new APNA session")
	pub, priv, err := apnams.GenKeyPairs()
	if err != nil {
		panic(err)
	}
	sessionEphIDGenReply, err := s.ApnaMS.EphIDGenerationRequest(apna.SessionEphID, s.SrvAddr, pub)
	if err != nil {
		panic(err)
	}
	if sessionEphIDGenReply.ErrorCode != apnams.ErrorEphIDGenOk {
		panic(sessionEphIDGenReply.ErrorCode.String())
	}
	ctrlEphIDSharedSecret, err := apnams.GenSharedSecret(data.Pubkey, s.CtrlEphIDPrivkey)
	if err != nil {
		panic(err)
	}
	sess := &Session{
		LocalPrivKey:     priv,
		LocalEphID:       sessionEphIDGenReply.Cert.Ephid,
		CtrlSharedSecret: ctrlEphIDSharedSecret,
	}
	s.SessionMap[pub.String()] = sess
	ecert, err := apnams.EncryptCert(sess.CtrlSharedSecret, &sessionEphIDGenReply.Cert)
	if err != nil {
		panic(err)
	}
	reply := &apna.Pkt{
		Which:       proto.APNAPkt_Which_ecert,
		LocalEphID:  data.RemoteEphID,
		RemoteEphID: data.LocalEphID,
		RemotePort:  data.LocalPort,
		LocalPort:   config.LocalAddr.L4Port,
		NextHeader:  0x01,
		Ecert:       ecert,
	}
	err = reply.Sign(server.Config.HMACKey)
	if err != nil {
		panic(err)
	}
	_, err = s.conn.WriteApnaTo(reply, raddr)
	if err != nil {
		panic(err)
	}
}

func (s Server) handshakePartTwo(data *apna.Pkt, raddr *snet.Addr) {
	log.Info("Second part of the handshake i.e. complete session on server side and sent handshake complete")
	localSessionPubkey := data.EcertPubkey.Pubkey.String()
	localSession, ok := s.SessionMap[localSessionPubkey]
	if !ok {
		panic("Unknown session")
	}
	remoteCert, err := apnams.DecryptCert(localSession.CtrlSharedSecret, data.EcertPubkey.Ecert)
	if err != nil {
		panic(err)
	}
	localSession.RemoteEphID = remoteCert.Ephid
	sessionSharedKey, err := apnams.GenSharedSecret(remoteCert.Pubkey, localSession.LocalPrivKey)
	if err != nil {
		panic(err)
	}
	localSession.SessionSharedSecret = sessionSharedKey
	s.SessionMap[localSessionPubkey] = localSession
	msg := []byte("Handshake Done")
	edata, err := apnams.EncryptData(localSession.SessionSharedSecret, msg)
	if err != nil {
		panic(err)
	}
	reply := &apna.Pkt{
		Which:       proto.APNAPkt_Which_data,
		LocalEphID:  localSession.LocalEphID,
		RemoteEphID: localSession.RemoteEphID,
		LocalPort:   config.LocalAddr.L4Port,
		RemotePort:  data.LocalPort,
		NextHeader:  0x03,
		Data:        edata,
	}
	err = reply.Sign(server.Config.HMACKey)
	if err != nil {
		panic(err)
	}
	_, err = s.conn.WriteApnaTo(reply, raddr)
	if err != nil {
		panic(err)
	}
}

func (s Server) handleConnection() {
	data, raddr, err := s.conn.ReadApna()
	if err != nil {
		panic(err)
	}
	log.Info("Details", "raddr", raddr)
	switch data.NextHeader {
	case 0x00:
		s.handshakePartOne(data, raddr)
	case 0x02:
		s.handshakePartTwo(data, raddr)
	default:
		log.Error("Unsupported next header")
	}
	log.Info("Recieved", "data", data)
}