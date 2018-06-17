package internal

import (
	"crypto/rand"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/apnad"
	"github.com/scionproto/scion/go/lib/xtest"
)

type testEncryption struct {
	scenario   string
	ephid      apnad.EphID
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

func setupTest() {
	apnad.LoadConfig("../testdata/apnad.json")
	Init()
}

func TestEncryptEphID(t *testing.T) {
	Convey("Given an EphID", t, func() {
		setupTest()
		tests := []testEncryption{
			{
				scenario:   "With normal parameters",
				ephid:      apnad.EphID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				encryptErr: "",
				decryptErr: "",
			},
		}
		for _, test := range tests {
			Convey(test.scenario, func() {
				iv, finalEphID, encryptErr := encryptEphID(&test.ephid)
				if test.encryptErr != "" {
					So(encryptErr.Error(), ShouldEqual, test.encryptErr)
				} else {
					So(encryptErr, ShouldBeNil)
					ephid, decryptErr := decryptEphID(iv, finalEphID)
					if test.decryptErr != "" {
						So(decryptErr.Error(), ShouldEqual, test.decryptErr)
					} else {
						So(decryptErr, ShouldBeNil)
					}
					So(*ephid, ShouldEqual, test.ephid)
				}
			})
		}
	})
}

func generateRandom(t *testing.T, size int) []byte {
	b := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, b)
	xtest.FailOnErr(t, err)
	return b
}

func TestMacComputation(t *testing.T) {
	Convey("Given IV and FinalEphID", t, func() {
		setupTest()
		tests := []testMac{
			{
				scenario:   "With normal parameters",
				finalEphID: generateRandom(t, apnad.EphIDLen),
				iv:         generateRandom(t, ivLen),
				computeErr: "",
				verifyErr:  "",
			},
		}
		for _, test := range tests {
			Convey(test.scenario, func() {
				mac, macErr := computeMac(test.iv, test.finalEphID)
				if test.computeErr != "" {
					So(macErr.Error(), ShouldEqual, test.computeErr)
				} else {
					So(macErr, ShouldBeNil)
					msg := append(test.iv, test.finalEphID...)
					result, verifyErr := verifyMac(msg, mac)
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
