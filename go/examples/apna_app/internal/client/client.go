package client

import (
	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apna_app/internal/config"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
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
	ApnaMS            apnams.Connector
	Config            *config.Config
	CtrlCertificate   apnams.Certificate
	CtrlEphIDPrivkey  common.RawBytes
	ServerSrvAddr     *apnams.ServiceAddr
	ServerCertificate apnams.Certificate
	SrvAddr           *apnams.ServiceAddr
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

func initApnaMS(conf *config.Config, client *Client, network string) error {
	var err error
	service := apnams.NewService(conf.IP.String(), conf.Port)
	client.ApnaMS, err = service.Connect()
	if err != nil {
		return err
	}
	pubkey, privkey, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	client.CtrlEphIDPrivkey = privkey
	proto, err := apnams.ProtocolStringToUint8(network)
	if err != nil {
		return err
	}
	srvAddr := &apnams.ServiceAddr{
		Addr:     config.LocalAddr.Host.IP(),
		Protocol: proto,
	}
	client.SrvAddr = srvAddr
	client.ServerSrvAddr = &apnams.ServiceAddr{
		Addr:     config.RemoteAddr.Host.IP(),
		Protocol: proto,
	}
	reply, err := client.ApnaMS.EphIDGenerationRequest(apna.SessionEphID,
		srvAddr, pubkey)
	if err != nil {
		return err
	}
	if reply.ErrorCode != apnams.ErrorEphIDGenOk {
		return common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	client.CtrlCertificate = reply.Cert
	dnsreply, err := client.ApnaMS.DNSRequest(client.ServerSrvAddr)
	if err != nil {
		return err
	}
	if dnsreply.ErrorCode != apnams.ErrorDNSOk {
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
	client.Config = conf
	err = initApnaMS(conf, client, network)
	if err != nil {
		panic(err)
	}
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
	log.Info("Bytes sent", "len", n)
	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		panic(err)
	}
	log.Info("Bytes received", "len", n)
	pld, err := apna.NewPktFromRaw(buf)
	if err != nil {
		panic(err)
	}
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
	finalReply, err := apna.NewPktFromRaw(buf)
	if err != nil {
		panic(err)
	}
	decryptData, err := apnams.DecryptData(client.Session.SharedSecret, finalReply.Data)
	if err != nil {
		panic(err)
	}
	log.Info("Finally", "buf", string(decryptData), "len", n)
}
