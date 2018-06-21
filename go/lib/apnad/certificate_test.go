package apnad

import (
	"crypto/rand"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/scionproto/scion/go/lib/xtest"
)

type testCertificate struct {
	scenario  string
	cert      *Certificate
	signErr   string
	verifyErr string
}

func setupTest() {
	LoadConfig("../../apnad/testdata/apnad.json")
}

func generateRandom(t *testing.T, size int) []byte {
	b := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, b)
	xtest.FailOnErr(t, err)
	return b
}

func TestSignCertificate(t *testing.T) {
	Convey("Given a certificate sign and verify it", t, func() {
		setupTest()
		tests := []testCertificate{
			{
				scenario: "With normal parameters",
				cert: &Certificate{
					Ephid:    generateRandom(t, 2*EphIDLen),
					Pubkey:   generateRandom(t, PubkeyLen),
					RecvOnly: 0x00,
					ExpTime:  generateRandom(t, TimestampLen),
				},
				signErr:   "",
				verifyErr: "",
			},
		}
		for _, test := range tests {
			Convey(test.scenario, func() {
				signErr := test.cert.Sign()
				if test.signErr != "" {
					So(signErr.Error(), ShouldEqual, test.signErr)
				} else {
					So(signErr, ShouldBeNil)
					verifyErr := test.cert.Verify()
					if test.verifyErr != "" {
						So(verifyErr.Error(), ShouldEqual, test.verifyErr)
					} else {
						So(verifyErr, ShouldBeNil)
					}
				}
			})
		}

	})
}
