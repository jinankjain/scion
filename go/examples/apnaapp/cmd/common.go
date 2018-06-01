package cmd

import (
	"net"
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

func issueEphID(conn *net.UDPConn) {
	msg := []byte{0}
	conn.Write(msg)
	finalEphID := make([]byte, 16)
	n, err := conn.Read(finalEphID)
	if err != nil {
		panic(err)
	}
	fmt.Println("ephid: ", finalEphID)
	fmt.Printf("bytes read: %v\n", n)
}
