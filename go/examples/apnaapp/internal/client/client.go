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
}

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
	client := &Client{}
	network := "udp4"
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
	data := &apna.Pld{
		Which:       proto.APNAHeader_Which_pubkey,
		LocalEphID:  client.CtrlCertificate.Ephid,
		RemoteEphID: client.ServerCertificate.Ephid,
		NextHeader:  0x00,
		Pubkey:      client.CtrlCertificate.Pubkey,
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
}
