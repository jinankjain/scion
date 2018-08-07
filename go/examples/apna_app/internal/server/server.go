package server

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
	Use:   "server",
	Short: "Run apna server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer(args)
	},
}

type Server struct {
	conn             *snet.Conn
	Config           *config.Config
	ApnaMS           apnams.Connector
	CtrlCertificate  apnams.Certificate
	CtrlEphIDPrivkey common.RawBytes
	SrvAddr          *apnams.ServiceAddr
	SessionMap       map[string]*Session
}

var server Server

func initApnaMS(conf *config.Config, server *Server, network string) error {
	var err error
	service := apnams.NewService(conf.IP.String(), conf.Port)
	server.ApnaMS, err = service.Connect()
	server.SessionMap = make(map[string]*Session)
	if err != nil {
		return err
	}
	pubkey, privkey, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	server.CtrlEphIDPrivkey = privkey
	proto, err := apnams.ProtocolStringToUint8(network)
	if err != nil {
		return err
	}
	srvAddr := &apnams.ServiceAddr{
		Addr:     config.LocalAddr.Host.IP(),
		Protocol: proto,
	}
	server.SrvAddr = srvAddr
	reply, err := server.ApnaMS.EphIDGenerationRequest(apna.CtrlEphID,
		srvAddr, pubkey)
	if err != nil {
		return err
	}
	if reply.ErrorCode != apnams.ErrorEphIDGenOk {
		return common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	server.CtrlCertificate = reply.Cert
	dnsreply, err := server.ApnaMS.DNSRegister(srvAddr, server.CtrlCertificate)
	if err != nil {
		return err
	}
	if dnsreply.ErrorCode != apnams.ErrorDNSRegisterOk {
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
	server.Config = conf
	log.Info("Server configuration", "conf", conf)
	// 2. Initialize apnams deamon
	network := "udp4"
	err = initApnaMS(conf, &server, network)
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
	sconn, err := snet.ListenSCION(network, &config.LocalAddr)
	if err != nil {
		panic(err)
	}
	server.conn = sconn
	log.Info("connection params", "conn", sconn.LocalSnetAddr())
	for /* ever */ {
		server.handleConnection()
	}
}
