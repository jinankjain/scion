package main

import (
	"fmt"
	"os"

	"github.com/scionproto/scion/go/lib/apnad"
)

func main() {
	service := apnad.NewService("127.0.0.1", 6000)
	connector, err := service.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	reply, err := connector.EphIDGenerationRequest(0x02)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(reply)
}
