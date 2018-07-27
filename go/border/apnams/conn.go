package apnams

import (
	"net"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

var ApnadConn apnad.Connector
var mapSiphashToHost map[string]net.IP

func InitApnad(conf string) error {
	mapSiphashToHost = make(map[string]net.IP)
	err := apnad.LoadConfig(conf)
	if err != nil {
		return err
	}
	service := apnad.NewService(apnad.ApnadConfig.IP.String(), apnad.ApnadConfig.Port)
	ApnadConn, err = service.Connect()
	if err != nil {
		return err
	}
	return nil
}

func SiphashToHost(siphash common.RawBytes) (net.IP, error) {
	if val, ok := mapSiphashToHost[siphash.String()]; ok {
		return val, nil
	}
	reply, err := ApnadConn.SiphashToHostRequest(siphash)
	if err != nil {
		return nil, err
	}
	log.Info("QQQQ", "reply", reply)
	if reply.ErrorCode != apnad.ErrorSiphashToHostOk {
		return nil, common.NewBasicError(reply.ErrorCode.String(), nil)
	}
	mapSiphashToHost[siphash.String()] = reply.Host
	return reply.Host, nil
}
