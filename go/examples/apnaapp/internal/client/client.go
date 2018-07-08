package client

import (
	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apnaapp/internal/config"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
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
	Session           *Session
}

type Session struct {
	LocalEphID       common.RawBytes
	RemoteEphID      common.RawBytes
	LocalPrivKey     common.RawBytes
	SharedSecret     common.RawBytes
	CtrlSharedSecret common.RawBytes
}

var client *Client

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
	reply, err := client.Apnad.EphIDGenerationRequest(apnad.GenerateSessionEphID,
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
	client = &Client{}
	initApnad(conf, client, network)
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
	msgPartOne, err := client.handshakePartOne()
	if err != nil {
		panic(err)
	}
	n, err := conn.Write(msgPartOne)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		panic(err)
	}
	pld, err := apna.NewPldFromRaw(buf)
	if err != nil {
		panic(err)
	}
	log.Debug("Client recieving server credentials", "server", pld)
	msgPartTwo, err := client.handshakePartTwo(pld)
	if err != nil {
		panic(err)
	}
	n, err = conn.Write(msgPartTwo)
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
	decryptData, err := apna.DecryptData(client.Session.SharedSecret, finalReply.Data)
	if err != nil {
		panic(err)
	}
	log.Info("Finally", "buf", string(decryptData), "len", n)
}
