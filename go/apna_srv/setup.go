package main

import (
	"time"

	"github.com/scionproto/scion/go/apna_srv/conf"
	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/snet"
)

const (
	initAttempts = 100
	initInterval = time.Second

	ErrorConf     = "Unable to load configuration"
	ErrorDispInit = "Unable to initialize dispatcher"
	ErrorSNET     = "Unable to create local SCION Network context"
)

func setup() error {
	config, err := conf.Load(*id, *confDir)
	if err != nil {
		return common.NewBasicError(ErrorConf, err)
	}
	if err := initSNET(config.PublicAddr.IA, initAttempts, initInterval); err != nil {
		return common.NewBasicError(ErrorSNET, err)
	}
	if *server {
		conn, err := snet.ListenSCIONWithBindSVC("udp4", config.PublicAddr, config.BindAddr,
			addr.SvcAP)
		if err != nil {
			return err
		}
		buf := make([]byte, 1024)
		for {
			len, addr, err := conn.ReadFrom(buf)
			if err != nil {
				return err
			}
			log.Info("Message info", "size", len, "addr", addr, "info", string(buf[:len]))
		}
	} else {
		raddrIA, _ := addr.IAFromString("2-ff00:0:222")
		raddr := &snet.Addr{IA: raddrIA, Host: addr.HostFromIP([]byte{127, 0, 1, 1}),
			L4Port: uint16(30075)}
		conn, err := snet.DialSCIONWithBindSVC("udp4", config.PublicAddr, raddr, config.BindAddr,
			addr.SvcAP)
		if err != nil {
			return err
		}
		conn.Write([]byte("Hello World!"))
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
