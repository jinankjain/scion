package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/scionproto/scion/go/apnad/internal"
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	flagConfig = flag.String("config", "", "Service JSON config file (required)")
)

func main() {
	flag.Parse()
	if *flagConfig == "" {
		fmt.Fprintln(os.Stderr, "Missing config file")
		flag.Usage()
		os.Exit(1)
	}
	err := apnad.LoadConfig(*flagConfig)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	internal.Init()
	err = internal.ListenAndServe(apnad.ApnadConfig.IP, apnad.ApnadConfig.Port)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
