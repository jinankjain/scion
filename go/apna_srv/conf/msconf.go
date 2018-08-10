package conf

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

type MSConf struct {
	IP      net.IP          `json:"ip"`
	Port    int             `json:"port"`
	HMACKey common.RawBytes `json:"hmacKey"`
	AESKey  common.RawBytes `json:"aesKey"`
	MyIP    net.IP          `json:"myip"`
}

func loadMSConfig(path string) (*MSConf, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, common.NewBasicError(ErrLoadingConfig, err, "path", path)
	}
	msconf := &MSConf{}
	err = json.Unmarshal(data, msconf)
	if err != nil {
		return nil, common.NewBasicError(ErrParsingConfig, err, "path", path)
	}
	return msconf, nil
}
