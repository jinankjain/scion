package server

import (
	"github.com/scionproto/scion/go/examples/apnaapp/internal/apna"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
)

func handleConnection(conn *snet.Conn) {
	buf := make([]byte, 100)
	n, raddr, err := conn.ReadFromSCION(buf)
	log.Info("Details", "raddr", raddr, "len", n)
	if err != nil {
		panic(err)
	}
	data, err := apna.NewPldFromRaw(buf)
	if err != nil {
		panic(err)
	}
	log.Info("Recieved", "data", data)
}
