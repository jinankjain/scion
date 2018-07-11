package apnad

import (
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
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
	HMACKey       common.RawBytes `json:"hmacKey"`
	AESKey        common.RawBytes `json:"aesKey"`
	SipHashKey    common.RawBytes `json:"siphashKey"`
}

var ApnadConfig Config

func LoadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return common.NewBasicError(ErrLoadingConfig, err, "path", path)
	}
	err = json.Unmarshal(data, &ApnadConfig)
	log.Info("APNA config", "SipHashKey", ApnadConfig.SipHashKey)
	if err != nil {
		return common.NewBasicError(ErrParsingConfig, err, "path", path)
	}
	return nil
}
