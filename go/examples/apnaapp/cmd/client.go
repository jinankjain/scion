package cmd

import (
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
)

func StartClient(client *snet.Addr, server *snet.Addr) {
	// Initialize default SCION networking context
	sciondAddr := sciond.GetDefaultSCIONDPath(&client.IA)
	dispatcher := getDefaultDispatcherSock()
	log.SetupLogConsole("apnaClient")
	if err := snet.Init(client.IA, sciondAddr, dispatcher); err != nil {
		log.Crit("Unable to initialize SCION network", "err", err)
	}
	cconn, err := snet.DialSCION("udp4", client, server)
	if err != nil {
		panic(err)
	}
	log.Debug("Client", "ephID", cconn.GetLocalEphID())
	log.Debug("Server", "ephID", cconn.GetRemoteEphID())
	n, err := cconn.Write([]byte("Hello!"))
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 1024)
	n, err = cconn.Read(buf)
	for n == 0 {
		n, err = cconn.Read(buf)
	}
	log.Info("Client Recived: ", buf[:n])
}
