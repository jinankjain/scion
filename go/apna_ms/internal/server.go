package internal

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra/transport"
	"github.com/scionproto/scion/go/lib/log"
)

func Init() {
	dnsRegister = make(map[uint8]map[string]apnams.Certificate)
	mapSiphashToHost = make(map[string]net.IP)
	macKeyRegister = make(map[string]common.RawBytes)
	siphashKey1 = binary.LittleEndian.Uint64(apnams.ApnaMSConfig.SipHashKey[:8])
	siphashKey2 = binary.LittleEndian.Uint64(apnams.ApnaMSConfig.SipHashKey[8:])
}

var (
	expName = flag.String("expName", "", "ExpName Required")
)

const (
	logDir = "apna_benchmark"
)

var ephidGenBenchmarks []*apnams.EphIDGenerationBenchmark
var dnsRegisterBenchmarks []*apnams.DNSRegisterBenchmark
var dnsRequestBenchmarks []*apnams.DNSRequestBenchmark
var macRegisterBenchmarks []*apnams.MACRegisterBenchmark
var macRequestBenchmarks []*apnams.MACRequestBenchmark
var siphashBenchmarks []*apnams.SiphashBenchmark

func cleanup() {
	t := time.Now()
	benchFile := fmt.Sprintf("%s-%s", *expName, t.Format("2006-01-02 15:04:05"))
	log.SetupLogFile(benchFile, logDir, "info", 20, 100, 0)
	switch *expName {
	case "ephid":
		for _, e := range ephidGenBenchmarks {
			log.Info(e.String())
		}
	case "dnsRegister":
		for _, e := range dnsRegisterBenchmarks {
			log.Info(e.String())
		}
	case "dnsRequest":
		for _, e := range dnsRequestBenchmarks {
			log.Info(e.String())
		}
	case "macRegister":
		for _, e := range macRegisterBenchmarks {
			log.Info(e.String())
		}
	case "macRequest":
		for _, e := range macRequestBenchmarks {
			log.Info(e.String())
		}
	case "siphash":
		for _, e := range siphashBenchmarks {
			log.Info(e.String())
		}
	}
}

func ListenAndServe(ip net.IP, port int) error {
	serverAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%v", ip, port))
	if err != nil {
		return err
	}
	serverConn, err := net.ListenUDP("udp4", serverAddr)
	if err != nil {
		return err
	}
	log.Info("Started APNA Management Service", "addr", serverConn.LocalAddr())
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()
	NewAPI(transport.NewPacketTransport(serverConn)).Serve()
	return nil
}
