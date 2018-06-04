package apna

import (
	"encoding/json"
	"io/ioutil"

	"github.com/scionproto/scion/go/lib/common"
)

const (
	ConfigFileName = "apna.json"
)

const (
	ErrLoadingConfig = "Error loading configuration"
	ErrParsingConfig = "Error parsing configuration"
)

type Configuration struct {
	SignAlgorithm string          `json:"signAlgo"`
	Pubkey        common.RawBytes `json:"pubkey"`
	Privkey       common.RawBytes `json:"privkey"`
}

func LoadConfig(path string) (*Configuration, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, common.NewBasicError(ErrLoadingConfig, err, "path", path)
	}
	var config *Configuration
	err = json.Unmarshal(data, config)
	if err != nil {
		return nil, common.NewBasicError(ErrParsingConfig, err, "path", path)
	}
	return config, nil
}
