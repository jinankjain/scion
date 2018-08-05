package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apna_app/internal/client"
	"github.com/scionproto/scion/go/examples/apna_app/internal/config"
	"github.com/scionproto/scion/go/examples/apna_app/internal/server"
)

var RootCmd = &cobra.Command{
	Use:   "apnaapp",
	Short: "Apna application for launching apna server and client",
	Long: `apnaapp is a tool to launch an apna server and client.
You can specify the address and port on which these services would be running`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&config.ApnadConfigPath, "apnad", "a", ".",
		"apnad config path")
	RootCmd.PersistentFlags().VarP(&config.LocalAddr, "local", "l", "local address")
	RootCmd.PersistentFlags().VarP(&config.RemoteAddr, "remote", "r", "remote address")

	RootCmd.AddCommand(server.Cmd)
	RootCmd.AddCommand(client.Cmd)
}
