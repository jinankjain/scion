package main

import (
	"fmt"
	"net"
	"os"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
)

func main() {
	service := apnad.NewService("127.0.0.1", 6000)
	connector, err := service.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	addr := apnad.ServiceAddr{
		Addr:     net.IP{127, 0, 0, 1},
		Protocol: 0x04,
	}
	reply, err := connector.EphIDGenerationRequest(0x00, &addr, common.RawBytes{}, 0x00)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("EphID Reply ", reply)
	dreply, err := connector.DNSRequest(&addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("DNS Reply ", dreply)

}
