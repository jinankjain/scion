package cmd

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"github.com/scionproto/scion/go/lib/crypto"
)

const (
	RandomKeySize   = 32
	apnaManagerPort = 3001
)

func connectToApnaManager() *net.UDPConn {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", apnaManagerPort))
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}
	return conn
}

func generateRandomKeyForHost() []byte {
	randomBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		panic(err)
	}
	return randomBytes
}

func issueSessionEphID(conn *net.UDPConn) {
	msg := []byte{1}
	pub, _, err := crypto.GenKeyPairs(crypto.Ed25519)
	if err != nil {
		panic(err)
	}
	msg = append(msg, pub...)
	conn.Write(msg)

}

func issueCtrlEphID(conn *net.UDPConn) []byte {
	msg := []byte{0}
	conn.Write(msg)
	finalEphID := make([]byte, 16)
	n, err := conn.Read(finalEphID)
	if err != nil {
		panic(err)
	}
	fmt.Println("ephid: ", finalEphID)
	fmt.Printf("bytes read: %v\n", n)
	return finalEphID
}

func getHostEphID(conn *net.UDPConn, serverAddr string) []byte {
	msg := make([]byte, 2)
	msg[0] = 0x03
	msg[1] = 0x00
	msg = append(msg, serverAddr...)
	n, err := conn.Write(msg)
	if err != nil {
		panic(err)
	}
	ephID := make([]byte, 8)
	n, err = conn.Read(ephID)
	if err != nil {
		panic(err)
	}
	if n != 8 {
		panic("Error in receiving EPHID")
	}
	return ephID
}

func performKeyExchange(server *net.Addr) {

}
