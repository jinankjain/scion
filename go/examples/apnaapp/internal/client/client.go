package client

import (
	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apnaapp/internal/config"
	"github.com/scionproto/scion/go/lib/log"
)

var Cmd = &cobra.Command{
	Use:   "client",
	Short: "Run apna client",
}

func init() {
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Client configuration", "conf", conf)
}
