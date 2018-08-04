package main

import (
	"fmt"
	"net"
	"time"

	"github.com/scionproto/scion/go/apna_srv/conf"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
)

const (
	initAttempts  = 100
	initInterval  = time.Second
	ErrorConf     = "Unable to load configuration"
	ErrorDispInit = "Unable to initialize dispatcher"
	ErrorSNET     = "Unable to create local SCION Network context"
	MaxBufSize    = 2 << 16
)

func setup() error {
	config, err := conf.Load(*id, *confDir)
	if err != nil {
		return common.NewBasicError(ErrorConf, err)
	}
	if err := initSNET(config.PublicAddr.IA, initAttempts, initInterval); err != nil {
		return common.NewBasicError(ErrorSNET, err)
	}
	done := make(chan error)
	go startServer(config.PublicAddr, done)
	go startSVC(config.PublicAddr, config.BindAddr, done)
	err = <-done
	if err != nil {
		return err
	}
	return nil
}

func initSNET(ia addr.IA, attempts int, sleep time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		if err = snet.Init(ia, *sciondPath, *dispPath); err == nil {
			break
		}
		log.Error("Unable to initialize snet", "Retry interval", sleep, "err", err)
		time.Sleep(sleep)
	}
	return err
}

func startSVC(pubAddr, bindAddr *snet.Addr, done chan error) {
	conn, err := snet.ListenSCIONWithBindSVC("udp4", pubAddr, bindAddr,
		addr.SvcAP)
	if err != nil {
		log.Error(err.Error())
		done <- err
	}
	log.Info("Started APNA service on", "addr", pubAddr.String())
	buf := make([]byte, MaxBufSize)
	for {
		len, addr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Error("Unable to read packet from the network", "err", err)
			continue
		}
		log.Info("Message info", "size", len, "addr", addr, "info", string(buf[:len]))
	}
}

func startServer(addr *snet.Addr, done chan error) {
	laddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%v", addr.Host.IP(), addr.L4Port))
	if err != nil {
		log.Error(err.Error())
		done <- err
	}
	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		log.Error(err.Error())
		done <- err
	}
	log.Info("Started UDP server on", "addr", laddr.String())
	buf := make([]byte, MaxBufSize)
	for {
		len, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Error("Unable to read network packet", "err", err)
			continue
		}
		log.Info("Message info", "size", len, "addr", raddr, "info", string(buf[:len]))
	}
}
