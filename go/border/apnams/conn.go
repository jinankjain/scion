package apnams

import (
	"github.com/scionproto/scion/go/lib/apnad"
)

func InitApnad(conf string) {
	apnad.LoadConfig(conf)
}
