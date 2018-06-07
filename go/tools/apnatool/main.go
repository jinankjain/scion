package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	signAlgo  string
	outputDir string
	ip        string
	port      int
)

func init() {
	flag.StringVar(&signAlgo, "signAlgo", crypto.Ed25519, "Sign Algorithm")
	flag.StringVar(&outputDir, "output", ".", "output directory")
	flag.IntVar(&port, "port", 6000, "management service port")
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip address for the management service")
}

func main() {
	flag.Parse()
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		log.Error("Invalid IP address")
		os.Exit(1)
	}
	pubkey, privkey, err := crypto.GenKeyPairs(signAlgo)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	c := apna.Configuration{
		SignAlgorithm: signAlgo,
		Pubkey:        pubkey,
		Privkey:       privkey,
		IP:            parsedIP,
		Port:          port,
	}
	b, err := json.Marshal(c)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Error("output dir does not exists")
	}
	ioutil.WriteFile(filepath.Join(absPath, apna.ConfigFileName), b, 0644)
}
