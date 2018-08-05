package internal

import (
	"encoding/binary"
	"hash"
	"net"
	"time"

	"github.com/scionproto/scion/go/lib/apna"
	"github.com/scionproto/scion/go/lib/apnams"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/log"

	"github.com/dchest/siphash"
)

var siphasher hash.Hash64
var siphashKey1 uint64
var siphashKey2 uint64
var epoch = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

const (
	CtrlEphIDDuration    = time.Hour
	SessionEphIDDuration = time.Minute * 5
)

func getExpTime(kind uint8) []byte {
	currTime := time.Now()
	switch kind {
	case apna.SessionEphID:
		currTime.Add(SessionEphIDDuration)
	case apna.CtrlEphID:
		currTime.Add(CtrlEphIDDuration)
	}
	timestamp := (currTime.Unix() - epoch) / 60
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(timestamp))
	return bs
}

var mapSiphashToHost map[string]net.IP

func generateHostID(addr net.IP) (common.RawBytes, error) {
	// TODO(jinankjain): Check bound on n
	hash := siphash.Hash(siphashKey1, siphashKey2, addr.To4())
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, hash)
	return b[:apna.HostLen], nil
}

func handleSiphashToHost(req *apnams.SiphashToHostReq) *apnams.SiphashToHostReply {
	log.Debug("Got SiphashToHost Request", "request", req)
	if val, ok := mapSiphashToHost[req.Siphash.String()]; ok {
		reply := &apnams.SiphashToHostReply{
			ErrorCode: apnams.ErrorSiphashToHostOk,
			Host:      val,
		}
		log.Debug("Reply EphIDGeneration sent", "reply", reply)
		return reply
	}
	reply := &apnams.SiphashToHostReply{
		ErrorCode: apnams.ErrorSiphashToHostNotFound,
	}
	log.Debug("Reply EphIDGeneration sent", "reply", reply)
	return reply
}

// handleEphIDGeneration: Generate EphID and sends it as response back
// @param: kind -> Control vs Session EphID generation
// @param: conn -> Connection to the management service of the host requesting EphID generation
// @param: retAddr -> Return address on which response would be send back
// @param: registerAddr -> EphID generation for this address
func handleEphIDGeneration(req *apnams.EphIDGenerationReq) *apnams.EphIDGenerationReply {
	log.Debug("Got EphIDGeneration request", "request", req)
	hostID, err := generateHostID(req.Addr.Addr)
	if err != nil {
		reply := &apnams.EphIDGenerationReply{
			ErrorCode: apnams.ErrorGenerateHostID,
		}
		return reply
	}
	mapSiphashToHost[hostID.String()] = req.Addr.Addr
	expTime := getExpTime(req.Kind)
	hid := apna.GetHID(req.Kind, hostID, expTime)

	ephid, err := apna.EncryptAndSignHostID(hid, apnams.ApnaMSConfig.AESKey,
		apnams.ApnaMSConfig.HMACKey)
	if err != nil {
		reply := &apnams.EphIDGenerationReply{
			ErrorCode: apnams.ErrorEncryptMACEphID,
		}
		return reply
	}

	cert := &apnams.Certificate{
		Ephid:    common.RawBytes(ephid),
		Pubkey:   req.Pubkey,
		RecvOnly: req.Kind,
		ExpTime:  expTime,
	}
	cert.Sign()
	reply := &apnams.EphIDGenerationReply{
		ErrorCode: apnams.ErrorEphIDGenOk,
		Cert:      *cert,
	}
	log.Debug("Reply EphIDGeneration sent", "reply", reply)
	return reply
}
