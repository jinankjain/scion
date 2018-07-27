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
	LoadConfig("testdata/apnad.json")
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

func benchmarkCertificateSign(c *Certificate, b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c.Sign()
	}
}

func BenchmarkCertificateSign1(b *testing.B) {
	c := &Certificate{
		Ephid:    xtest.MustParseHexString("76321d3891568e2c9b90daccf63100ba"),
		Pubkey:   xtest.MustParseHexString("d174e21f4b9bfa965e94a2bade1d2f01ec7e8f10de011c0df52aea7b2e33b632"),
		RecvOnly: 1,
		ExpTime:  xtest.MustParseHexString("de610400"),
	}
	setupTest()
	benchmarkCertificateSign(c, b)
}

func BenchmarkCertificateSign2(b *testing.B) {
	c := &Certificate{
		Ephid:    xtest.MustParseHexString("b2c8294b8584d4a8e76a756a3ff65678"),
		Pubkey:   xtest.MustParseHexString("29062cf5c42d1d9f63cd7d91acb07a66225c78e8062ce3e3cd177a25b612d706"),
		RecvOnly: 0,
		ExpTime:  xtest.MustParseHexString("de610400"),
	}
	setupTest()
	benchmarkCertificateSign(c, b)
}
