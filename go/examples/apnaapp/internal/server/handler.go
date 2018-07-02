package server

import (
	"github.com/scionproto/scion/go/examples/apnaapp/internal/apna"
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/proto"
)

func (s Server) handshakePartOne(data *apna.Pld, raddr *snet.Addr) {
	log.Info("First part of the handshake i.e. got a request to initiate new APNA session")
	pub, priv, err := apnad.GenKeyPairs()
	if err != nil {
		panic(err)
	}
	sessionEphIDGenReply, err := s.Apnad.EphIDGenerationRequest(apnad.GenerateSessionEphID, s.SrvAddr, pub)
	if err != nil {
		panic(err)
	}
	if sessionEphIDGenReply.ErrorCode != apnad.ErrorEphIDGenOk {
		panic(sessionEphIDGenReply.ErrorCode.String())
	}
	ctrlEphIDSharedSecret, err := apnad.GenSharedSecret(data.Pubkey, s.CtrlEphIDPrivkey)
	if err != nil {
		panic(err)
	}
	sess := &Session{
		LocalPrivKey:     priv,
		LocalEphID:       sessionEphIDGenReply.Cert.Ephid,
		CtrlSharedSecret: ctrlEphIDSharedSecret,
	}
	s.SessionMap[pub.String()] = sess
	ecert, err := apna.EncryptCert(sess.CtrlSharedSecret, &sessionEphIDGenReply.Cert)
	if err != nil {
		panic(err)
	}
	reply := &apna.Pld{
		Which:       proto.APNAHeader_Which_ecert,
		LocalEphID:  data.RemoteEphID,
		RemoteEphID: data.LocalEphID,
		NextHeader:  0x01,
		Ecert:       ecert,
	}
	rawBytes, err := reply.RawPld()
	if err != nil {
		panic(err)
	}
	_, err = s.conn.WriteTo(rawBytes, raddr)
	if err != nil {
		panic(err)
	}
}

func (s Server) handshakePartTwo(data *apna.Pld, raddr *snet.Addr) {
	log.Info("Second part of the handshake i.e. complete session on server side and sent handshake complete")
	localSessionPubkey := data.EcertPubkey.Pubkey.String()
	localSession, ok := s.SessionMap[localSessionPubkey]
	if !ok {
		panic("Unknown session")
	}
	remoteCert, err := apna.DecryptCert(localSession.CtrlSharedSecret, data.EcertPubkey.Ecert)
	if err != nil {
		panic(err)
	}
	localSession.RemoteEphID = remoteCert.Ephid
	sessionSharedKey, err := apnad.GenSharedSecret(remoteCert.Pubkey, localSession.LocalPrivKey)
	if err != nil {
		panic(err)
	}
	localSession.SessionSharedSecret = sessionSharedKey
	s.SessionMap[localSessionPubkey] = localSession
	msg := []byte("Handshake Done")
	edata, err := apna.EncryptData(localSession.SessionSharedSecret, msg)
	if err != nil {
		panic(err)
	}
	reply := &apna.Pld{
		Which:       proto.APNAHeader_Which_data,
		LocalEphID:  localSession.LocalEphID,
		RemoteEphID: localSession.RemoteEphID,
		NextHeader:  0x03,
		Data:        edata,
	}
	rawBytes, err := reply.RawPld()
	if err != nil {
		panic(err)
	}
	_, err = s.conn.WriteTo(rawBytes, raddr)
	if err != nil {
		panic(err)
	}
}

func (s Server) handleConnection() {
	buf := make([]byte, 1024)
	n, raddr, err := s.conn.ReadFromSCION(buf)
	log.Info("Details", "raddr", raddr, "len", n)
	if err != nil {
		panic(err)
	}
	data, err := apna.NewPldFromRaw(buf)
	if err != nil {
		panic(err)
	}
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
