package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/scionproto/scion/go/lib/addr"
	//"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/snet"
)

var apnaManagerPort = 3001

func getDefaultSCIONDPath(ia addr.IA) string {
	return fmt.Sprintf("/run/shm/sciond/sd%s.sock", ia.FileFmt(false))
}

func getDefaultDispatcherSock() string {
	return "/run/shm/dispatcher/default.sock"
}

func StartServer(server *snet.Addr) {
	// Initialize default SCION networking context
	sciond := getDefaultSCIONDPath(server.IA)
	dispatcher := getDefaultDispatcherSock()
	if err := snet.Init(server.IA, sciond, dispatcher); err != nil {
		log.Fatal("Unable to initialize SCION network", "err", err)
	}
	log.Print("SCION Network successfully initialized")

	// Connect to management service
	apnaConn := connectToApnaManager()

	// Issue CtrlEphID
	issueEphID(apnaConn)

	verify := make([]byte, 2)
	verify[0] = 0x03
	verify[1] = 0x00
	addr := []byte(conn.LocalAddr().String())
	verify = append(verify, addr...)
	fmt.Println("msg to be send: ", verify)
	conn.Write(verify)

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
