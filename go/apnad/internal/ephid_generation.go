package internal

import (
	"encoding/binary"
	"hash"
	"time"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"
)

var siphasher hash.Hash64
var epoch = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

const (
	CtrlEphIDDuration    = time.Hour
	SessionEphIDDuration = time.Minute * 5
)

func getExpTime(kind uint8) []byte {
	currTime := time.Now()
	switch kind {
	case apnad.GenerateSessionEphID:
		currTime.Add(SessionEphIDDuration)
	case apnad.GenerateCtrlEphID:
		currTime.Add(CtrlEphIDDuration)
	}
	timestamp := (currTime.Unix() - epoch) / 60
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(timestamp))
	return bs
}

func generateHostID(addr []byte) ([]byte, error) {
	// TODO(jinankjain): Check bound on n
	_, err := siphasher.Write(addr)
	if err != nil {
		return nil, common.NewBasicError(apnad.ErrorGenerateHostID.String(), err)
	}
	return siphasher.Sum(nil)[:apnad.HostIDLen], nil
}

// handleEphIDGeneration: Generate EphID and sends it as response back
// @param: kind -> Control vs Session EphID generation
// @param: conn -> Connection to the management service of the host requesting EphID generation
// @param: retAddr -> Return address on which response would be send back
// @param: registerAddr -> EphID generation for this address
func handleEphIDGeneration(req *apnad.EphIDGenerationReq) *apnad.EphIDGenerationReply {
	log.Debug("Got request", "request", req)
	var ephID apnad.EphID
	// 1. Copy the kind inside ephID
	copy(ephID[apnad.TypeOffset:apnad.HostIDOffset], []byte{req.Kind})
	// 2. Generate hostID and put it inside ephID
	hostID, err := generateHostID(req.Addr.Addr)
	if err != nil {
		reply := &apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorGenerateHostID,
		}
		return reply
	}
	copy(ephID[apnad.HostIDOffset:apnad.TimestampOffset], hostID)
	// 3. Get the expiration time and append to ephID
	copy(ephID[apnad.TimestampOffset:], getExpTime(req.Kind))
	iv, encryptedEphID, err := encryptEphID(&ephID)
	if err != nil {
		reply := &apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorEncryptEphID,
		}
		log.Debug("Reply sent", "reply", reply)
		return reply
	}
	mac, err := computeMac(iv, encryptedEphID)
	if err != nil {
		reply := &apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorMACCompute,
		}
		log.Debug("Reply sent", "reply", reply)
		return reply
	}
	response := append(iv, encryptedEphID...)
	response = append(response, mac...)
	reply := &apnad.EphIDGenerationReply{
		ErrorCode: apnad.ErrorEphIDGenOk,
		Ephid:     response,
	}
	dnsRegister[req.Addr.Protocol] = make(map[string][]byte)
	dnsRegister[req.Addr.Protocol][req.Addr.Addr.String()] = response
	log.Debug("Reply sent", "reply", reply)
	return reply
}
