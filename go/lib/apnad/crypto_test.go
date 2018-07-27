package apnad

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/common"
)

type testEncryption struct {
	scenario   string
	ephid      EphID
	encryptErr string
	decryptErr string
}

type testMac struct {
	scenario   string
	finalEphID []byte
	iv         []byte
	computeErr string
	verifyErr  string
}

func setupCryptoTest() {
	LoadConfig("testdata/apnad.json")
}

func TestEncryptEphID(t *testing.T) {
	Convey("Given an EphID", t, func() {
		setupCryptoTest()
		tests := []testEncryption{
			{
				scenario:   "With normal parameters",
				ephid:      EphID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				encryptErr: "",
				decryptErr: "",
			},
		}
		for _, test := range tests {
			Convey(test.scenario, func() {
				iv, finalEphID, encryptErr := EncryptEphID(&test.ephid)
				if test.encryptErr != "" {
					So(encryptErr.Error(), ShouldEqual, test.encryptErr)
				} else {
					So(encryptErr, ShouldBeNil)
					ephid, decryptErr := DecryptEphID(iv, finalEphID)
					if test.decryptErr != "" {
						So(decryptErr.Error(), ShouldEqual, test.decryptErr)
					} else {
						So(decryptErr, ShouldBeNil)
					}
					So(*ephid, ShouldResemble, test.ephid)
				}
			})
		}
	})
}

func TestMacComputation(t *testing.T) {
	Convey("Given IV and FinalEphID", t, func() {
		setupTest()
		tests := []testMac{
			{
				scenario:   "With normal parameters",
				finalEphID: generateRandom(t, EphIDLen),
				iv:         generateRandom(t, IvLen),
				computeErr: "",
				verifyErr:  "",
			},
		}
		for _, test := range tests {
			Convey(test.scenario, func() {
				mac, macErr := ComputeMac(test.iv, test.finalEphID)
				if test.computeErr != "" {
					So(macErr.Error(), ShouldEqual, test.computeErr)
				} else {
					So(macErr, ShouldBeNil)
					msg := append(test.iv, test.finalEphID...)
					result, verifyErr := VerifyMac(msg, mac)
					if test.verifyErr != "" {
						So(verifyErr.Error(), ShouldEqual, test.verifyErr)
						So(result, ShouldBeFalse)
					} else {
						So(verifyErr, ShouldBeNil)
						So(result, ShouldBeTrue)
					}
				}
			})
		}
	})
}

var resultIV common.RawBytes

func benchmarkEncryptEphid(ephid *EphID, b *testing.B) {
	b.ResetTimer()
	var result common.RawBytes
	for n := 0; n < b.N; n++ {
		result, _, _ = EncryptEphID(ephid)
	}
	resultIV = result
}

func BenchmarkEncryptEphID1(b *testing.B) {
	ephid := &EphID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	setupCryptoTest()
	benchmarkEncryptEphid(ephid, b)
}

func BenchmarkEncryptEphID2(b *testing.B) {
	ephid := &EphID{0xf1, 0xd2, 0xe3, 0xc4, 0x2b, 0xa6, 0x27, 0x18}
	setupCryptoTest()
	benchmarkEncryptEphid(ephid, b)
}

var resultEphid *EphID

func benchmarkDecryptEphid(iv common.RawBytes, finalEphID common.RawBytes, b *testing.B) {
	b.ResetTimer()
	var result *EphID
	for n := 0; n < b.N; n++ {
		result, _ = DecryptEphID(iv, finalEphID)
	}
	resultEphid = result
}

func BenchmarkDecryptEphID1(b *testing.B) {
	iv := []byte{0x01, 0x02, 0x03, 0x04}
	finalEphID := []byte{0x2c, 0x4f, 0x2e, 0x84, 0x7a, 0x3c, 0x31, 0x51}
	setupCryptoTest()
	benchmarkDecryptEphid(iv, finalEphID, b)
}

func BenchmarkDecryptEphID2(b *testing.B) {
	iv := []byte{0xa1, 0xb2, 0xc3, 0xd4}
	finalEphID := []byte{0x2c, 0x4f, 0x2e, 0x84, 0x7a, 0x3c, 0xf1, 0x21}
	setupCryptoTest()
	benchmarkDecryptEphid(iv, finalEphID, b)
}

func benchmarkComputeMac(iv common.RawBytes, finalEphID common.RawBytes, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ComputeMac(iv, finalEphID)
	}
}

func BenchmarkComputeMac1(b *testing.B) {
	iv := []byte{0x01, 0x02, 0x03, 0x04}
	finalEphID := []byte{0x2c, 0x4f, 0x2e, 0x84, 0x7a, 0x3c, 0x31, 0x51}
	setupCryptoTest()
	benchmarkComputeMac(iv, finalEphID, b)
}

func BenchmarkComputeMac2(b *testing.B) {
	iv := []byte{0xa1, 0xb2, 0xc3, 0xd4}
	finalEphID := []byte{0x2c, 0x4f, 0x2e, 0x84, 0x7a, 0x3c, 0xf1, 0x21}
	setupCryptoTest()
	benchmarkComputeMac(iv, finalEphID, b)
}

func benchmarkVerifyMac(msg common.RawBytes, mac common.RawBytes, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		VerifyMac(msg, mac)
	}
}

func BenchmarkVerifyMac1(b *testing.B) {
	//76321d3891568e2c9b90daccf63100ba
	msg := []byte{0x76, 0x32, 0x1d, 0x38, 0x91, 0x56, 0x8e, 0x2c, 0x9b, 0x90, 0xda, 0xcc}
	mac := []byte{0xf6, 0x31, 0x00, 0xba}
	setupCryptoTest()
	benchmarkVerifyMac(msg, mac, b)
}

func BenchmarkVerifyMac2(b *testing.B) {
	// 5e1a8299b7a3b2e1477984c2460b54df
	msg := []byte{0x5e, 0x1a, 0x82, 0x99, 0xb7, 0xa3, 0xb2, 0xe1, 0x47, 0x79, 0x84, 0xc2}
	mac := []byte{0x46, 0x0b, 0x54, 0xdf}
	setupCryptoTest()
	benchmarkVerifyMac(msg, mac, b)
}
