package ephidbenchmark

import (
	"net"

	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apna_app/internal/config"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
)

var Cmd = &cobra.Command{
	Use:   "ephid_benchmark",
	Short: "Run ephid benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		startEphidBenchmark(args)
	},
}

var trials int
var repetitions int

func init() {
	trials = *Cmd.PersistentFlags().IntP("trials", "t", 10000, "Number of trials in each repetitions")
	repetitions = *Cmd.PersistentFlags().IntP("repetitions", "n", 5, "Number of repetitions")
}

type EphIDBenchmark struct {
	Apnad apnams.Connector
}

var ephidBenchmark EphIDBenchmark

func initApnad(conf *config.Config) error {
	var err error
	svc := apnams.NewService(conf.IP.String(), conf.Port, conf.MyIP)
	ephidBenchmark.Apnad, err = svc.Connect()
	if err != nil {
		return err
	}
	return nil
}

func runBenchmark() error {
	pubkey, _, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	network := "udp4"
	proto, err := apnams.ProtocolStringToUint8(network)
	if err != nil {
		return err
	}
	srvAddr := &apnams.ServiceAddr{
		Addr:     net.IP{127, 0, 0, 1},
		Protocol: proto,
	}
	for i := 0; i < trials; i++ {
		_, err := ephidBenchmark.Apnad.EphIDGenerationRequest(apna.CtrlEphID, srvAddr, pubkey)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func startEphidBenchmark(args []string) {
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Server configuration", "conf", conf)
	err = initApnad(conf)
	if err != nil {
		panic(err)
	}
	err = runBenchmark()
	if err != nil {
		panic(err)
	}
}
