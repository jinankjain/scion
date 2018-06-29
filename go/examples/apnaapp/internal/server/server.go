package server

import (
	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apnaapp/internal/config"
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
	Use:   "server",
	Short: "Run apna server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer(args)
	},
}

type Session struct {
	LocalPubKey         common.RawBytes
	LocalPrivKey        common.RawBytes
	SessionSharedSecret common.RawBytes
	CtrlSharedSecret    common.RawBytes
	RemotePubKey        common.RawBytes
	LocalEphID          common.RawBytes
	RemoteEphID         common.RawBytes
}

type Server struct {
	Apnad            apnad.Connector
	CtrlCertificate  apnad.Certificate
	CtrlEphIDPrivkey common.RawBytes
	SrvAddr          *apnad.ServiceAddr
	SessionMap       map[string]*Session
}

var server Server

func initApnad(conf *config.Config, server *Server, network string) error {
	var err error
	service := apnad.NewService(conf.IP.String(), conf.Port)
	server.Apnad, err = service.Connect()
	server.SessionMap = make(map[string]*Session)
	if err != nil {
		return err
	}
	pubkey, privkey, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	server.CtrlEphIDPrivkey = privkey
	proto, err := apnad.ProtocolStringToUint8(network)
	if err != nil {
		return err
	}
	srvAddr := &apnad.ServiceAddr{
		Addr:     config.LocalAddr.Host.IP(),
		Protocol: proto,
	}
	server.SrvAddr = srvAddr
	reply, err := server.Apnad.EphIDGenerationRequest(apnad.GenerateCtrlEphID,
		srvAddr, pubkey)
	if err != nil {
		return err
	}
	if reply.ErrorCode != apnad.ErrorEphIDGenOk {
		return common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	server.CtrlCertificate = reply.Cert
	dnsreply, err := server.Apnad.DNSRegister(srvAddr, server.CtrlCertificate)
	if err != nil {
		return err
	}
	if dnsreply.ErrorCode != apnad.ErrorDNSRegisterOk {
		return common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	return nil
}

func startServer(args []string) {
	// 1. Load config
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Server configuration", "conf", conf)
	// 2. Initialize APNAD deamon
	network := "udp4"
	initApnad(conf, &server, network)
	// 3. Initialize SCION related stuff
	sciondSock := sciond.GetDefaultSCIONDPath(&config.LocalAddr.IA)
	dispatcher := getDefaultDispatcherSock()
	if err := snet.Init(config.LocalAddr.IA, sciondSock, dispatcher); err != nil {
		log.Crit("Unable to initialize SCION network", "err", err)
	}
	log.Info("SCION Network successfully initialized")
	sconn, err := snet.ListenSCION(network, &config.LocalAddr)
	if err != nil {
		panic(err)
	}
	log.Info("connection params", "conn", sconn.LocalSnetAddr())
	for /* ever */ {
		server.handleConnection(sconn)
	}
}
