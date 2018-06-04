package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/crypto"
	"github.com/scionproto/scion/go/lib/log"
)

var (
	signAlgo  string
	outputDir string
)

func init() {
	flag.StringVar(&signAlgo, "signAlgo", crypto.Ed25519, "Sign Algorithm")
	flag.StringVar(&outputDir, "output", ".", "output directory")
}

func main() {
	flag.Parse()
	pubkey, privkey, err := crypto.GenKeyPairs(signAlgo)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	c := apna.Configuration{
		SignAlgorithm: signAlgo,
		Pubkey:        pubkey,
		Privkey:       privkey,
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
