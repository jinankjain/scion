package server

import (
	"github.com/scionproto/scion/go/examples/apnaapp/internal/apna"
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/proto"
)

func (s Server) handleConnection(conn *snet.Conn) {
	buf := make([]byte, 1024)
	n, raddr, err := conn.ReadFromSCION(buf)
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
		log.Info("Got request to initiate new APNA session")
		pub, priv, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
		if err != nil {
			panic(err)
		}
		reply, err := s.Apnad.EphIDGenerationRequest(apnad.GenerateSessionEphID, s.SrvAddr, pub)
		if err != nil {
			panic(err)
		}
		if reply.ErrorCode != apnad.ErrorEphIDGenOk {
			panic(reply.ErrorCode.String())
		}
		sharedKey, err := crypto.GenSharedSecret(data.Pubkey, s.CtrlEphIDPrivkey,
			crypto.Curve25519xSalsa20Poly1305)
		if err != nil {
			panic(err)
		}
		sess := Session{
			LocalPubKey:      pub,
			LocalPrivKey:     priv,
			LocalEphID:       reply.Cert.Ephid,
			CtrlSharedSecret: sharedKey,
		}
		s.SessionMap[pub.String()] = &sess
		ecert := &apna.Ecert{
			Cert:      &reply.Cert,
			SharedKey: sharedKey,
		}
		edata, err := ecert.Encrypt()
		if err != nil {
			panic(err)
		}
		sreply := &apna.Pld{
			Which:       proto.APNAHeader_Which_ecert,
			LocalEphID:  data.RemoteEphID,
			RemoteEphID: data.LocalEphID,
			NextHeader:  0x01,
			Ecert:       edata,
		}
		sreplyToBytes, err := proto.PackRoot(sreply)
		if err != nil {
			panic(err)
		}
		n, err := conn.WriteTo(sreplyToBytes, raddr)
		if err != nil {
			panic(err)
		}
		log.Info("bytes sent", "len", n)
	case 0x02:
		log.Info("Got request for second part of handshake")
		sessPubkey := data.EcertPubkey.Pubkey.String()
		sessId := *s.SessionMap[sessPubkey]
		ecert := &apna.Ecert{
			SharedKey: sessId.CtrlSharedSecret,
		}
		clientCert, err := ecert.Decrypt(data.EcertPubkey.Ecert)
		if err != nil {
			panic(err)
		}
		sessId.RemoteEphID = clientCert.Pubkey
		sessId.RemotePubKey = clientCert.Pubkey
		sessSharedKey, err := crypto.GenSharedSecret(clientCert.Pubkey,
			s.SessionMap[sessPubkey].LocalPrivKey,
			crypto.Curve25519xSalsa20Poly1305)
		if err != nil {
			panic(err)
		}
		sessId.SessionSharedSecret = sessSharedKey
		s.SessionMap[sessPubkey] = &sessId
		data := []byte("Handshake Done")
		reply := &apna.Pld{
			Which:       proto.APNAHeader_Which_data,
			LocalEphID:  s.SessionMap[sessPubkey].LocalEphID,
			RemoteEphID: s.SessionMap[sessPubkey].RemoteEphID,
			NextHeader:  0x03,
			Data:        data,
		}
		replyRaw, err := proto.PackRoot(reply)
		if err != nil {
			panic(err)
		}
		n, err := conn.WriteTo(replyRaw, raddr)
		if err != nil {
			panic(err)
		}
		log.Info("bytes sent in last phase", "len", n)
	default:
		log.Error("Unsupported next header")
	}
	log.Info("Recieved", "data", data)
}
