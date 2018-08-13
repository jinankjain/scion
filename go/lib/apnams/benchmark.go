package apnams

import (
	"fmt"
	"time"
)

type EphIDGenerationBenchmark struct {
	HostID  time.Duration
	ExpTime time.Duration
	Encrypt time.Duration
	Mac     time.Duration
}

func (b *EphIDGenerationBenchmark) String() string {
	return fmt.Sprintf("%d %d %d %d", b.HostID, b.ExpTime, b.Encrypt, b.Mac)
}

type DNSRegisterBenchmark struct {
	RegisterTime time.Duration
}

type DNSRequestBenchmark struct {
	RequestTime time.Duration
}

type MACRegisterBenchmark struct {
	RegisterTime time.Duration
}

type MACRequestBenchmark struct {
	RequestTime time.Duration
}

type SiphashBenchmark struct {
	SiphashTime time.Duration
}
