package cmd

import (
	//"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
	"github.com/scionproto/scion/go/lib/snet"
)

func getDefaultDispatcherSock() string {
	return "/run/shm/dispatcher/default.sock"
}

func StartServer(server *snet.Addr) {
	// Initialize default SCION networking context
	sciondSock := sciond.GetDefaultSCIONDPath(&server.IA)
	dispatcher := getDefaultDispatcherSock()
	if err := snet.Init(server.IA, sciondSock, dispatcher); err != nil {
		log.Crit("Unable to initialize SCION network", "err", err)
	}
	log.Info("SCION Network successfully initialized")
	sconn, err := snet.ListenSCION("udp4", server)
	if err != nil {
		panic(err)
	}
	log.Debug("Server", "ephID", sconn.GetLocalEphID().Pack())
	for /* ever */ {
		handleConnection(sconn)
	}
}

func handleConnection(conn *snet.Conn) {
	buf := make([]byte, 1024)
	n, raddr, err := conn.ReadFromSCION(buf)
	if err != nil {
		panic(err)
	}
	log.Info("Data Received: ", "buf", buf[:n])
	n, err = conn.WriteToSCION([]byte("Bye!"), raddr)
	if err != nil {
		panic(err)
	}
	log.Info("Reply Sent of size: ", "n", n)
}
