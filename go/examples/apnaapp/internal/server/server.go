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

type Server struct {
	Apnad            apnad.Connector
	CtrlCertificate  apnad.Certificate
	CtrlEphIDPrivkey common.RawBytes
}

func initApnad(conf *config.Config, server *Server, network string) error {
	var err error
	service := apnad.NewService(conf.IP.String(), conf.Port)
	server.Apnad, err = service.Connect()
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
	reply, err := server.Apnad.EphIDGenerationRequest(apnad.GenerateCtrlEphID,
		srvAddr, pubkey)
	if err != nil {
		return err
	}
	server.CtrlCertificate = reply.Cert
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
	server := &Server{}
	network := "udp4"
	initApnad(conf, server, network)
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
}
