package apna

type Addr []byte

type EphID struct {
	HostID    []byte
	Timestamp uint32
	Type      []byte
}

const (
	AddrLen   = 16
	HostIDLen = 3
	TypeLen   = 1
)

func (addr Addr) String() string {
	if len(addr) == 0 {
		return "<nil>"
	}
	return string(addr)
}
