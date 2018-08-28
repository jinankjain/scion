package kmsbenchmark

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
	"github.com/scionproto/scion/go/lib/log"
)

var MACRegisterCMD = &cobra.Command{
	Use:   "mac_register_benchmark",
	Short: "Run mac key register benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		startMACRegisterBenchmark(args)
	},
}

var MACRequestCMD = &cobra.Command{
	Use:   "mac_request_benchmark",
	Short: "Run mac key request benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		startMACRequestBenchmark(args)
	},
}

var reqTrials int
var reqRepetitions int

var regTrials int
var regRepetitions int

func init() {
	reqTrials = *MACRequestCMD.PersistentFlags().IntP("trials", "t", 10000,
		"Number of trials in each repetitions")
	reqRepetitions = *MACRequestCMD.PersistentFlags().IntP("repetitions", "n", 5,
		"Number of repetitions")
	regTrials = *MACRegisterCMD.PersistentFlags().IntP("trials", "t", 10000,
		"Number of trials in each repetitions")
	regRepetitions = *MACRegisterCMD.PersistentFlags().IntP("repetitions", "n", 5,
		"Number of repetitions")
}

type KMSBenchmark struct {
	Apnad apnams.Connector
}

var kmsBenchmark KMSBenchmark

func initApnad(conf *config.Config) error {
	var err error
	svc := apnams.NewService(conf.IP.String(), conf.Port, conf.MyIP)
	kmsBenchmark.Apnad, err = svc.Connect()
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

func runMACRegisterBenchmark(conf *config.Config) error {
	srvAddrs, err := generateServiceAddress("127.0.0.1/18")
	if err != nil {
		return err
	}
	for i := 0; i < regTrials; i++ {
		_, err = kmsBenchmark.Apnad.MacKeyRegister(srvAddrs[i].Addr, 5000, conf.HMACKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func startMACRegisterBenchmark(args []string) {
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Server configuration", "conf", conf)
	err = initApnad(conf)
	if err != nil {
		panic(err)
	}
	err = runMACRegisterBenchmark(conf)
	if err != nil {
		panic(err)
	}
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

func runMACRequestBenchmark(conf *config.Config) error {
	srvAddrs, err := generateServiceAddress("127.0.0.1/18")
	if err != nil {
		return err
	}
	for i := 0; i < reqTrials; i++ {
		_, err = kmsBenchmark.Apnad.MacKeyRegister(srvAddrs[i].Addr.To4(), 5000, conf.HMACKey)
		if err != nil {
			return err
		}
	}
	for i := 0; i < reqTrials; i++ {
		hid, err := generateHostID(srvAddrs[i].Addr.To4())
		if err != nil {
			return err
		}
		_, err = kmsBenchmark.Apnad.MacKeyRequest(hid, 5000)
		if err != nil {
			return err
		}
	}
	return nil
}

func startMACRequestBenchmark(args []string) {
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
	err = runMACRequestBenchmark(conf)
	if err != nil {
		panic(err)
	}
}
