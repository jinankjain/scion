package apnad

import (
	"fmt"
	"net"

	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/snet"
)

var Config *apna.Configuration

type ManagementService struct {
	conn *net.Conn
}

const (
	UnableToConnectToMS = "Unable to connect to APNA Management Service"
	MSNotRunning        = "Management Service is not running"
)

func NewManagementService() (*ManagementService, error) {
	if Config == nil {
		return nil, common.NewBasicError(MSNotRunning, nil)
	}
	msconn, err := net.Dial("udp", fmt.Sprintf("%s:%v", Config.IP.String(), Config.Port))
	if err != nil {
		return nil, common.NewBasicError(UnableToConnectToMS, err)
	}
	return &ManagementService{conn: &msconn}, nil
}

func (m *ManagementService) GenerateEphID() {
}

func (m *ManagementService) DNSRequest(raddr *snet.Addr) {
}
