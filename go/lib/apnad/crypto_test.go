package apnad

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
		setupTest()
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
					fmt.Printf("%x", ephid[:])
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
