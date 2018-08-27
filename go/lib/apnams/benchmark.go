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

func (b *DNSRegisterBenchmark) String() string {
	return fmt.Sprintf("%d", b.RegisterTime)
}

type DNSRequestBenchmark struct {
	RequestTime time.Duration
}

func (b *DNSRequestBenchmark) String() string {
	return fmt.Sprintf("%d", b.RequestTime)
}

type MACRegisterBenchmark struct {
	RegisterTime time.Duration
}

func (b *MACRegisterBenchmark) String() string {
	return fmt.Sprintf("%d", b.RegisterTime)
}

type MACRequestBenchmark struct {
	RequestTime time.Duration
}

func (b *MACRequestBenchmark) String() string {
	return fmt.Sprintf("%d", b.RequestTime)
}

type SiphashBenchmark struct {
	SiphashTime time.Duration
}

func (b *SiphashBenchmark) String() string {
	return fmt.Sprintf("%d", b.SiphashTime)
}
