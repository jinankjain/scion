package apna

type Addr []byte

const (
	AddrLen = 16
)

func (addr Addr) String() string {
	if len(addr) == 0 {
		return "<nil>"
	}
	return string(addr)
}
