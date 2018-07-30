package internal

import (
	"encoding/binary"
	"hash"
	"net"
	"time"

	"github.com/scionproto/scion/go/lib/apnad"
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

var mapSiphashToHost map[string]net.IP

func generateHostID(addr net.IP) (common.RawBytes, error) {
	// TODO(jinankjain): Check bound on n
	hash := siphash.Hash(siphashKey1, siphashKey2, addr.To4())
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, hash)
	return b[:apnad.HostIDLen], nil
}

func handleSiphashToHost(req *apnad.SiphashToHostReq) *apnad.SiphashToHostReply {
	log.Debug("Got SiphashToHost Request", "request", req)
	if val, ok := mapSiphashToHost[req.Siphash.String()]; ok {
		reply := &apnad.SiphashToHostReply{
			ErrorCode: apnad.ErrorSiphashToHostOk,
			Host:      val,
		}
		log.Debug("Reply EphIDGeneration sent", "reply", reply)
		return reply
	}
	reply := &apnad.SiphashToHostReply{
		ErrorCode: apnad.ErrorSiphashToHostNotFound,
	}
	log.Debug("Reply EphIDGeneration sent", "reply", reply)
	return reply
}

// handleEphIDGeneration: Generate EphID and sends it as response back
// @param: kind -> Control vs Session EphID generation
// @param: conn -> Connection to the management service of the host requesting EphID generation
// @param: retAddr -> Return address on which response would be send back
// @param: registerAddr -> EphID generation for this address
func handleEphIDGeneration(req *apnad.EphIDGenerationReq) *apnad.EphIDGenerationReply {
	var ephID apnad.EphID
	// 1. Copy the kind inside ephID
	copy(ephID[apnad.TypeOffset:apnad.HostIDOffset], []byte{req.Kind})

	// 2. Generate hostID and put it inside ephID
	ephidBenchmark := &apnad.EphIDBenchmark{}
	start := time.Now()
	hostID, err := generateHostID(req.Addr.Addr)
	if err != nil {
		reply := &apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorGenerateHostID,
		}
		return reply
	}
	mapSiphashToHost[hostID.String()] = req.Addr.Addr
	copy(ephID[apnad.HostIDOffset:apnad.TimestampOffset], hostID)
	ephidBenchmark.HostIDGenerationTime = time.Since(start)

	// 3. Get the expiration time and append to ephID
	start = time.Now()
	expTime := getExpTime(req.Kind)
	copy(ephID[apnad.TimestampOffset:], expTime)
	ephidBenchmark.ExpTimeGenerationTime = time.Since(start)

	start = time.Now()
	iv, encryptedEphID, err := apnad.EncryptEphID(&ephID)
	if err != nil {
		reply := &apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorEncryptEphID,
		}
		return reply
	}
	response := append(iv, encryptedEphID...)
	ephidBenchmark.EncryptEphidTime = time.Since(start)

	start = time.Now()
	mac, err := apnad.ComputeMac(iv, encryptedEphID)
	if err != nil {
		reply := &apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorMACCompute,
		}
		return reply
	}
	response = append(response, mac...)
	ephidBenchmark.MacComputeTime = time.Since(start)

	start = time.Now()
	cert := &apnad.Certificate{
		Ephid:    response,
		Pubkey:   req.Pubkey,
		RecvOnly: req.Kind,
		ExpTime:  expTime,
	}
	err = cert.Sign()
	if err != nil {
		reply := &apnad.EphIDGenerationReply{
			ErrorCode: apnad.ErrorSignCertificate,
		}
		return reply
	}
	reply := &apnad.EphIDGenerationReply{
		ErrorCode: apnad.ErrorEphIDGenOk,
		Cert:      *cert,
	}
	ephidBenchmark.CertificateSignTime = time.Since(start)
	log.Info(ephidBenchmark.String())
	return reply
}
