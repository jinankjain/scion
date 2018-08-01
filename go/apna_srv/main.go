package main

import (
	"flag"
	"os"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
)

var (
	id         = flag.String("id", "", "Element ID (Required. E.g. 'ap4-ff00:0:2f')")
	sciondPath = flag.String("sciond", sciond.GetDefaultSCIONDPath(nil), "SCIOND socket path")
	dispPath   = flag.String("dispatcher", "", "SCION Dispatcher path")
	confDir    = flag.String("confd", "", "Configuration directory (Required)")
	prom       = flag.String("prom", "127.0.0.1:1282", "Address to export prometheus metrics on")
)

func main() {
	var err error
	flag.Parse()
	if *id == "" {
		log.Crit("No element ID specified")
		flag.Usage()
		os.Exit(1)
	}
	defer log.LogPanicAndExit()
	if err = checkFlags(); err != nil {
		fatal(err.Error())
	}
	log.Info("Started")
}

func checkFlags() error {
	if *sciondPath == "" {
		flag.Usage()
		return common.NewBasicError("No SCIOND path specified", nil)
	}
	if *confDir == "" {
		flag.Usage()
		return common.NewBasicError("No configuration directory specified", nil)
	}
	return nil
}

func fatal(msg string, args ...interface{}) {
	log.Crit(msg, args...)
	log.Flush()
	os.Exit(1)
}
