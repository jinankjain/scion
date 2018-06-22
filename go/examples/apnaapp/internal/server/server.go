package server

import (
	"flag"

	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apnaapp/internal/config"
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
}

var (
	server snet.Addr
)

func main() {
	flag.Parse()
	sciondSock := sciond.GetDefaultSCIONDPath(&server.IA)
	dispatcher := getDefaultDispatcherSock()
	if err := snet.Init(server.IA, sciondSock, dispatcher); err != nil {
		log.Crit("Unable to initialize SCION network", "err", err)
	}
	log.Info("SCION Network successfully initialized")
	sconn, err := snet.ListenSCION("udp4", &server)
	if err != nil {
		panic(err)
	}
	log.Info("Local Ephid", "ephid", sconn.CtrlEphid())
	reply, err := sconn.RegisterWithDNS()
	if err != nil {
		panic(err)
	}
	log.Info("Got reply", "reply", reply)
}

func init() {
	flag.Var((*snet.Addr)(&server), "local", "(Mandatory) address to listen on")
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Server configuration", "conf", conf)
}
