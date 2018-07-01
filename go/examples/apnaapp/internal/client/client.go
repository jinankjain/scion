package client

import (
	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apnaapp/internal/apna"
	"github.com/scionproto/scion/go/examples/apnaapp/internal/config"
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
	"github.com/scionproto/scion/go/proto"
)

func getDefaultDispatcherSock() string {
	return "/run/shm/dispatcher/default.sock"
}

var Cmd = &cobra.Command{
	Use:   "client",
	Short: "Run apna client",
	Run: func(cmd *cobra.Command, args []string) {
		startClient(args)
	},
}

type Client struct {
	Apnad             apnad.Connector
	CtrlCertificate   apnad.Certificate
	CtrlEphIDPrivkey  common.RawBytes
	ServerSrvAddr     *apnad.ServiceAddr
	ServerCertificate apnad.Certificate
	SrvAddr           *apnad.ServiceAddr
}

type Session struct {
	LocalEphID   common.RawBytes
	RemoteEphID  common.RawBytes
	LocalPrivKey common.RawBytes
	LocalPubKey  common.RawBytes
	RemotePubKey common.RawBytes
	SharedSecret common.RawBytes
}

var client Client

func initApnad(conf *config.Config, client *Client, network string) error {
	var err error
	service := apnad.NewService(conf.IP.String(), conf.Port)
	client.Apnad, err = service.Connect()
	if err != nil {
		return err
	}
	pubkey, privkey, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	client.CtrlEphIDPrivkey = privkey
	proto, err := apnad.ProtocolStringToUint8(network)
	if err != nil {
		return err
	}
	srvAddr := &apnad.ServiceAddr{
		Addr:     config.LocalAddr.Host.IP(),
		Protocol: proto,
	}
	client.SrvAddr = srvAddr
	client.ServerSrvAddr = &apnad.ServiceAddr{
		Addr:     config.RemoteAddr.Host.IP(),
		Protocol: proto,
	}
	reply, err := client.Apnad.EphIDGenerationRequest(apnad.GenerateCtrlEphID,
		srvAddr, pubkey)
	if err != nil {
		return err
	}
	if reply.ErrorCode != apnad.ErrorEphIDGenOk {
		return common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	client.CtrlCertificate = reply.Cert
	dnsreply, err := client.Apnad.DNSRequest(client.ServerSrvAddr)
	if err != nil {
		return err
	}
	if dnsreply.ErrorCode != apnad.ErrorDNSOk {
		return common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	client.ServerCertificate = dnsreply.Certificate
	return nil
}

func startClient(args []string) {
	// 1. Load the config
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Client configuration", "conf", conf)
	// 2. Initialize APNAD deamon
	network := "udp4"
	initApnad(conf, &client, network)
	// 3. Initialize SCION related stuff
	sciondSock := sciond.GetDefaultSCIONDPath(&config.LocalAddr.IA)
	dispatcher := getDefaultDispatcherSock()
	if err := snet.Init(config.LocalAddr.IA, sciondSock, dispatcher); err != nil {
		log.Crit("Unable to initialize SCION network", "err", err)
	}
	log.Info("SCION Network successfully initialized")
	conn, err := snet.DialSCION(network, &config.LocalAddr, &config.RemoteAddr)
	if err != nil {
		panic(err)
	}
	log.Info("connection params", "conn", conn.LocalSnetAddr())
	data := &apna.Pld{
		Which:       proto.APNAHeader_Which_pubkey,
		LocalEphID:  client.CtrlCertificate.Ephid,
		RemoteEphID: client.ServerCertificate.Ephid,
		NextHeader:  0x00,
		Pubkey:      client.CtrlCertificate.Pubkey,
	}
	sharedCtrlSecret, err := crypto.GenSharedSecret(client.ServerCertificate.Pubkey,
		client.CtrlEphIDPrivkey, crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		panic(err)
	}
	marshal, err := proto.PackRoot(data)
	if err != nil {
		panic(err)
	}
	n, err := conn.Write(marshal)
	if err != nil {
		panic(err)
	}
	log.Info("Bytes sent", "len", n)
	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		panic(err)
	}
	log.Info("Bytes received", "len", n)
	pld, err := apna.NewPldFromRaw(buf)
	if err != nil {
		panic(err)
	}
	if pld.NextHeader != 0x01 {
		panic("Broken handshake")
	}
	serverCert, err := apna.DecryptCert(sharedCtrlSecret, pld.Ecert)
	if err != nil {
		panic(err)
	}
	sessionPubkey, sessionPrivkey, err := apnad.GenKeyPairs()
	if err != nil {
		panic(err)
	}
	sessEphIDReply, err := client.Apnad.EphIDGenerationRequest(apnad.GenerateSessionEphID,
		client.SrvAddr, sessionPubkey)
	if err != nil {
		panic(err)
	}
	if sessEphIDReply.ErrorCode != apnad.ErrorEphIDGenOk {
		panic(sessEphIDReply.ErrorCode.String())
	}
	sessSecret, err := crypto.GenSharedSecret(serverCert.Pubkey, sessionPrivkey,
		crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		panic(err)
	}
	sess := &Session{
		LocalEphID:   sessEphIDReply.Cert.Ephid,
		RemoteEphID:  serverCert.Ephid,
		LocalPubKey:  sessionPubkey,
		LocalPrivKey: sessionPrivkey,
		RemotePubKey: serverCert.Pubkey,
		SharedSecret: sessSecret,
	}
	log.Info("Established session", "sess", sess)
	esessCert, err := apna.EncryptCert(sharedCtrlSecret, &sessEphIDReply.Cert)
	if err != nil {
		panic(err)
	}
	partTwoReply := &apna.Pld{
		Which:       proto.APNAHeader_Which_ecertPubkey,
		LocalEphID:  client.CtrlCertificate.Ephid,
		RemoteEphID: client.ServerCertificate.Ephid,
		NextHeader:  0x02,
		EcertPubkey: apna.EcertPubkey{
			Ecert:  esessCert,
			Pubkey: serverCert.Pubkey,
		},
		Pubkey: serverCert.Pubkey,
	}
	partTwoReplyBytes, err := proto.PackRoot(partTwoReply)
	if err != nil {
		panic(err)
	}
	n, err = conn.Write(partTwoReplyBytes)
	if err != nil {
		panic(err)
	}
	log.Info("Number of bytes written", "len", n)
	n, err = conn.Read(buf)
	if err != nil {
		panic(err)
	}
	finalReply, err := apna.NewPldFromRaw(buf)
	if err != nil {
		panic(err)
	}
	decryptData, err := apna.DecryptData(sess.SharedSecret, finalReply.Data)
	if err != nil {
		panic(err)
	}
	log.Info("Finally", "buf", string(decryptData), "len", n)
}
