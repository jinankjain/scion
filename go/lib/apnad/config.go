package apnad

import (
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/scionproto/scion/go/lib/common"
)

const (
	ConfigFileName = "apna.json"
)

const (
	ErrLoadingConfig = "Error loading configuration"
	ErrParsingConfig = "Error parsing configuration"
)

type Config struct {
	IP            net.IP          `json:"ip"`
	Port          int             `json:"port"`
	SignAlgorithm string          `json:"signAlgo"`
	Pubkey        common.RawBytes `json:"pubkey"`
	Privkey       common.RawBytes `json:"privkey"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, common.NewBasicError(ErrLoadingConfig, err, "path", path)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, common.NewBasicError(ErrParsingConfig, err, "path", path)
	}
	return &config, nil
}
