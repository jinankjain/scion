package apnamgr

import (
	"fmt"
	"net"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
	"github.com/scionproto/scion/go/lib/sciond"
)

type AP struct {
	sciondConn sciond.Connector
	log.Logger
}

const (
	ErrConnectSCIOND = "Error connecting to SCIOND"
)

func New(srvc sciond.Service, logger log.Logger) (*AP, error) {
	sciondConn, err := srvc.Connect()
	if err != nil {
		return nil, common.NewBasicError(ErrConnectSCIOND, err)
	}
	apnamgr := &AP{
		sciondConn: sciondConn,
		Logger:     logger.New("lib", "APNASvcResolver"),
	}
	return apnamgr, nil
}

func (a *AP) Query() (*net.UDPAddr, error) {
	reply, err := a.sciondConn.SVCInfo([]sciond.ServiceType{sciond.SvcAP})
	if err != nil {
		return nil, common.NewBasicError("Unable to find APNA SVC address", err)
	}
	if len(reply.Entries) == 0 {
		return nil, common.NewBasicError("Reply contains no APNA Svc address", nil)
	}
	found := false
	var hostInfo sciond.HostInfo
	for _, e := range reply.Entries {
		if found {
			break
		}
		for _, h := range e.HostInfos {
			hostInfo = h
			found = true
			break
		}
	}
	if !found {
		return nil, common.NewBasicError("Found no SVC listening", nil)
	}
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%v", hostInfo.Host().IP(),
		hostInfo.Port))
	if err != nil {
		return nil, err
	}
	return addr, nil
}
