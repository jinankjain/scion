package main

import (
	"flag"

	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "apna.json", "management service configuration")
}

var Config *apna.Configuration

func main() {
	flag.Parse()
	var err error
	Config, err = apna.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}
	log.Info("Manager configuration", "conf", Config)
	err = RunServer(3001)
	if err != nil {
		panic(err)
	}
}
