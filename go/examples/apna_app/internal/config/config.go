package config

import (
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/snet"
)

type Config struct {
	ApnadPubkey  common.RawBytes `json:"apnadPubkey"`
	SymmetricKey common.RawBytes `json:"symmetricKey"`
	HMACKey      common.RawBytes `json:"hmacKey"`
	IP           net.IP          `json:"ip"`
	Port         int             `json:"port"`
	MyIP         net.IP          `json:"myip"`
}

const (
	ErrReadingFile = "Error while reading file"
	ErrJSONMarshal = "Error while marshaling data into json"
)

var (
	ApnadConfigPath string
	LocalAddr       snet.Addr
	RemoteAddr      snet.Addr
)

func LoadConfig() (*Config, error) {
	data, err := ioutil.ReadFile(ApnadConfigPath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
