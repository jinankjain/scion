package internal

import (
	"encoding/binary"
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

const (
	logDir = "apna_benchmark"
)

var ephidGenBenchmarks []*apnams.EphIDGenerationBenchmark

func cleanup() {
	for _, e := range ephidGenBenchmarks {
		log.Info(e.String())
	}
}

func InitBenchmark() {
	t := time.Now()
	ephidBench := fmt.Sprintf("%s-%s", "ephid_benchmark", t.Format("2006-01-02 15:04:05"))
	log.SetupLogFile(ephidBench, logDir, "info", 20, 100, 0)
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
	InitBenchmark()
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
