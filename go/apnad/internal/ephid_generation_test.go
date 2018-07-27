package internal

import (
	"encoding/binary"
	"net"
	"testing"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/common"
)

var resultExpTime []byte

func benchmarkGetExpTime(kind uint8, b *testing.B) {
	var result []byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result = getExpTime(kind)
	}
	resultExpTime = result
}

func BenchmarkCtrlEphIDExptime(b *testing.B) {
	benchmarkGetExpTime(apnad.GenerateCtrlEphID, b)
}

func BenchmarkSessionEphIDExptime(b *testing.B) {
	benchmarkGetExpTime(apnad.GenerateSessionEphID, b)
}

func setupTestForGenerateHostID() {
	apnad.LoadConfig("../testdata/apnad.json")
	siphashKey1 = binary.LittleEndian.Uint64(apnad.ApnadConfig.SipHashKey[:8])
	siphashKey2 = binary.LittleEndian.Uint64(apnad.ApnadConfig.SipHashKey[8:])
}

var resultHostID common.RawBytes

func benchmarkGenerateHostID(addr net.IP, b *testing.B) {
	var result common.RawBytes
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, _ = generateHostID(addr)
	}
	resultHostID = result
}

func BenchmarkGenerateHostID(b *testing.B) {
	setupTestForGenerateHostID()
	addr := net.IP{127, 0, 0, 1}
	benchmarkGenerateHostID(addr, b)
}
