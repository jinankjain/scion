package apnams

import (
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/scionproto/scion/go/lib/common"
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

var ApnaMSConfig Config

func LoadConfig(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return common.NewBasicError(ErrLoadingConfig, err, "path", path)
	}
	err = json.Unmarshal(data, &ApnaMSConfig)
	if err != nil {
		return common.NewBasicError(ErrParsingConfig, err, "path", path)
	}
	return nil
}
