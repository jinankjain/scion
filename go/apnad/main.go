package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/scionproto/scion/go/apnad/internal"
	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	flagConfig = flag.String("config", "", "Service JSON config file (required)")
	expName    = flag.String("exp", "", "Experiment Name")
)

const logDir = "apna_benchmark"

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
	t := time.Now()
	name := fmt.Sprintf("%s-%s", *expName, t.Format("2006-01-02 15:04:05"))
	log.SetupLogFile(name, logDir, "info", 20, 100, 0)
	internal.Init()
	err = internal.ListenAndServe(apnad.ApnadConfig.IP, apnad.ApnadConfig.Port)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
