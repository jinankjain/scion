package hidbenchmark

import (
	"encoding/binary"
	"encoding/hex"
	"net"

	"github.com/dchest/siphash"
	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apna_app/internal/config"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
)

var Cmd = &cobra.Command{
	Use:   "hid_benchmark",
	Short: "Run hid benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		startHidBenchmark(args)
	},
}

var trials int
var repetitions int

func init() {
	trials = *Cmd.PersistentFlags().IntP("trials", "t", 10000, "Number of trials in each repetitions")
	repetitions = *Cmd.PersistentFlags().IntP("repetitions", "n", 5, "Number of repetitions")
}

type HIDBenchmark struct {
	Apnad apnams.Connector
}

var hidBenchmark HIDBenchmark

func initApnad(conf *config.Config) error {
	var err error
	svc := apnams.NewService(conf.IP.String(), conf.Port, conf.MyIP)
	hidBenchmark.Apnad, err = svc.Connect()
	if err != nil {
		return err
	}
	return nil
}

func generateServiceAddress(cidr string) ([]*apnams.ServiceAddr, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	network := "udp4"
	proto, err := apnams.ProtocolStringToUint8(network)
	if err != nil {
		return nil, err
	}
	var srvAddrs []*apnams.ServiceAddr
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		srvAddr := &apnams.ServiceAddr{
			Protocol: proto,
		}
		srvAddr.Addr = make([]byte, len(ip))
		copy(srvAddr.Addr, ip)
		srvAddrs = append(srvAddrs, srvAddr)
	}
	return srvAddrs, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func runHIDBenchmark() error {
	pubkey, _, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	srvAddrs, err := generateServiceAddress("10.0.0.1/18")
	if err != nil {
		return err
	}
	for i := 0; i < trials; i++ {
		_, err := hidBenchmark.Apnad.EphIDGenerationRequest(apna.CtrlEphID, srvAddrs[i],
			pubkey)
		if err != nil {
			return err
		}
	}
	for i := 0; i < trials; i++ {
		hid, err := generateHostID(srvAddrs[i].Addr.To4())
		if err != nil {
			return err
		}
		_, err = hidBenchmark.Apnad.SiphashToHostRequest(hid)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateHostID(addr net.IP) (common.RawBytes, error) {
	// TODO(jinankjain): Check bound on n
	hash := siphash.Hash(siphashKey1, siphashKey2, addr.To4())
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, hash)
	return b[:apna.HostLen], nil
}

var (
	siphashKey1 uint64
	siphashKey2 uint64
)

func startHidBenchmark(args []string) {
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	siphashKey, _ := hex.DecodeString("dae7ace5b7723bd4ec5986a8d25f12c6")
	siphashKey1 = binary.LittleEndian.Uint64(siphashKey[:8])
	siphashKey2 = binary.LittleEndian.Uint64(siphashKey[8:])

	log.Info("Server configuration", "conf", conf)
	err = initApnad(conf)
	if err != nil {
		panic(err)
	}
	err = runHIDBenchmark()
	if err != nil {
		panic(err)
	}
}
