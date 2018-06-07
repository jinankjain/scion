package cmd

import (
	"fmt"
	"log"

	"github.com/scionproto/scion/go/lib/addr"
	//"github.com/scionproto/scion/go/lib/crypto"
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
		log.Fatal("Unable to initialize SCION network", "err", err)
	}
	log.Print("SCION Network successfully initialized")

	// Connect to management service
	apnaConn := connectToApnaManager()

	// Issue CtrlEphID
	issueCtrlEphID(apnaConn)

	verify := make([]byte, 2)
	verify[0] = 0x03
	verify[1] = 0x00
	addr := []byte(apnaConn.LocalAddr().String())
	verify = append(verify, addr...)
	fmt.Println("msg to be send: ", verify)
	apnaConn.Write(verify)

	sconn, err := snet.ListenSCION("udp4", server)
	if err != nil {
		panic(err)
	}
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
	log.Print("Data Received: ", buf[:n])
	n, err = conn.WriteToSCION([]byte("Bye!"), raddr)
	if err != nil {
		panic(err)
	}
	log.Print("Reply Sent of size: ", n)
}
