package dnsbenchmark

import (
	"net"

	"github.com/spf13/cobra"

	"github.com/scionproto/scion/go/examples/apna_app/internal/config"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
)

var DNSRegisterCMD = &cobra.Command{
	Use:   "dns_register_benchmark",
	Short: "Run dns register benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		startDNSRegisterBenchmark(args)
	},
}

var DNSRequestCMD = &cobra.Command{
	Use:   "dns_request_benchmark",
	Short: "Run dns request benchmark",
	Run: func(cmd *cobra.Command, args []string) {
		startDNSRequestBenchmark(args)
	},
}

var reqTrials int
var reqRepetitions int

var regTrials int
var regRepetitions int

func init() {
	reqTrials = *DNSRequestCMD.PersistentFlags().IntP("trials", "t", 10000,
		"Number of trials in each repetitions")
	reqRepetitions = *DNSRequestCMD.PersistentFlags().IntP("repetitions", "n", 5,
		"Number of repetitions")
	regTrials = *DNSRegisterCMD.PersistentFlags().IntP("trials", "t", 10000,
		"Number of trials in each repetitions")
	regRepetitions = *DNSRegisterCMD.PersistentFlags().IntP("repetitions", "n", 5,
		"Number of repetitions")
}

type DNSBenchmark struct {
	Apnad apnams.Connector
}

var dnsBenchmark DNSBenchmark

func initApnad(conf *config.Config) error {
	var err error
	svc := apnams.NewService(conf.IP.String(), conf.Port, conf.MyIP)
	dnsBenchmark.Apnad, err = svc.Connect()
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
		ip = ip.To4()
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

func runDNSRegisterBenchmark() error {
	pubkey, _, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	srvAddrs, err := generateServiceAddress("127.0.0.1/18")
	if err != nil {
		return err
	}
	for i := 0; i < regTrials; i++ {
		ephidCert, err := dnsBenchmark.Apnad.EphIDGenerationRequest(apna.CtrlEphID, srvAddrs[i],
			pubkey)
		if err != nil {
			return err
		}
		_, err = dnsBenchmark.Apnad.DNSRegister(srvAddrs[i], ephidCert.Cert)
		if err != nil {
			return err
		}
	}
	return nil
}

func startDNSRegisterBenchmark(args []string) {
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Server configuration", "conf", conf)
	err = initApnad(conf)
	if err != nil {
		panic(err)
	}
	err = runDNSRegisterBenchmark()
	if err != nil {
		panic(err)
	}
}

func runDNSRequestBenchmark() error {
	pubkey, _, err := crypto.GenKeyPairs(crypto.Curve25519xSalsa20Poly1305)
	if err != nil {
		return err
	}
	srvAddrs, err := generateServiceAddress("127.0.0.1/18")
	if err != nil {
		return err
	}
	for i := 0; i < reqTrials; i++ {
		ephidCert, err := dnsBenchmark.Apnad.EphIDGenerationRequest(apna.CtrlEphID, srvAddrs[i],
			pubkey)
		if err != nil {
			return err
		}
		_, err = dnsBenchmark.Apnad.DNSRegister(srvAddrs[i], ephidCert.Cert)
		if err != nil {
			return err
		}
	}
	for i := 0; i < reqTrials; i++ {
		_, err := dnsBenchmark.Apnad.DNSRequest(srvAddrs[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func startDNSRequestBenchmark(args []string) {
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	log.Info("Server configuration", "conf", conf)
	err = initApnad(conf)
	if err != nil {
		panic(err)
	}
	err = runDNSRequestBenchmark()
	if err != nil {
		panic(err)
	}
}
