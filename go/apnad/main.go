package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	flagConfig = flag.String("config", "", "Service TOML config file (required)")
)

func main() {
	flag.Parse()
	if *flagConfig == "" {
		fmt.Fprintln(os.Stderr, "Missing config file")
		flag.Usage()
		os.Exit(1)
	}
	config, err := apnad.LoadConfig(*flagConfig)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println(config)
}
