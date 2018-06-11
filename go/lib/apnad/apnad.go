package apnad

import (
	"net"
	"os"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/common"
	//	"github.com/scionproto/scion/go/lib/log"
)

var Config *apna.Configuration

type ManagementService struct {
	conn *net.UDPConn
}

const (
	UnableToConnectToMS = "Unable to connect to APNA Management Service"
	MSNotRunning        = "Management Service is not running"
	NotConnectedToMS    = "Not connected to Management Service"
	UnsupportedOpcode   = "Unsuported opcode"
	FailedToWrite       = "Failed to write request to Management Service"
	FailedToRead        = "Failed to read response from Management Service"
)

const (
	GenerateCtrlEphID    = 0x00
	GenerateSessionEphID = 0x01
	DNSRequest           = 0x03
)

const (
	UDPProtocol = 0x00
)

func NewManagementService() (*ManagementService, error) {
	var err error
	Config, err = apna.LoadConfig(os.Getenv("SC") + "/apnad/apna.json")
	if Config == nil {
		return nil, common.NewBasicError(MSNotRunning, nil)
	}
	raddr := &net.UDPAddr{IP: Config.IP, Port: Config.Port}
	msconn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, common.NewBasicError(UnableToConnectToMS, err)
	}
	return &ManagementService{conn: msconn}, nil
}

func (m *ManagementService) GenerateEphID(opcode byte) (addr.HostAPNA, error) {
	if m.conn == nil {
		return nil, common.NewBasicError(NotConnectedToMS, nil)
	}
	if opcode != GenerateCtrlEphID && opcode != GenerateSessionEphID {
		return nil, common.NewBasicError(UnsupportedOpcode, nil, "opcode ", opcode)
	}
	msg := []byte{opcode}
	n, err := m.conn.Write(msg)
	if err != nil {
		return nil, common.NewBasicError(FailedToWrite, err)
	}
	if n != len(msg) {
		return nil, common.NewBasicError(FailedToWrite, err, "bytes ", n)
	}
	response := make([]byte, apna.AddrLen)
	n, err = m.conn.Read(response)
	if err != nil {
		return nil, common.NewBasicError(FailedToRead, err)
	}
	if n != apna.AddrLen {
		return nil, common.NewBasicError(FailedToRead, nil, "bytes ", n)
	}
	return response, nil
}

func (m *ManagementService) DNSRequest(raddr string) (addr.HostAPNA, error) {
	if m.conn == nil {
		return nil, common.NewBasicError(NotConnectedToMS, nil)
	}
	msg := []byte{DNSRequest, UDPProtocol}
	raddrBytes := []byte(raddr)
	msg = append(msg, raddrBytes...)
	n, err := m.conn.Write(msg)
	if err != nil {
		return nil, common.NewBasicError(FailedToWrite, err)
	}
	if n != len(msg) {
		return nil, common.NewBasicError(FailedToWrite, err, "bytes ", n)
	}
	response := make([]byte, apna.AddrLen)
	n, err = m.conn.Read(response)
	if err != nil {
		return nil, common.NewBasicError(FailedToRead, err)
	}
	if n != apna.AddrLen {
		return nil, common.NewBasicError(FailedToRead, nil, "bytes ", n)
	}
	return response, nil

}
