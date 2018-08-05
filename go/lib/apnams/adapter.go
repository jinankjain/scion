package apnams

import (
	"strconv"

	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/infra/disp"
	"github.com/scionproto/scion/go/proto"
)

var _ disp.MessageAdapter = (*Adapter)(nil)

type Adapter struct{}

func (a *Adapter) MsgToRaw(msg proto.Cerealizable) (common.RawBytes, error) {
	apnadMsg := msg.(*Pld)
	return proto.PackRoot(apnadMsg)
}

func (a *Adapter) RawToMsg(b common.RawBytes) (proto.Cerealizable, error) {
	return NewPldFromRaw(b)
}

func (a *Adapter) MsgKey(msg proto.Cerealizable) string {
	apnadMsg := msg.(*Pld)
	return strconv.FormatUint(apnadMsg.Id, 10)
}
