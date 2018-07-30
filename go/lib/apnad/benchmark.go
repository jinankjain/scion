package apnad

import (
	"fmt"
	"time"
)

type EphIDBenchmark struct {
	HostIDGenerationTime  time.Duration
	ExpTimeGenerationTime time.Duration
	EncryptEphidTime      time.Duration
	MacComputeTime        time.Duration
	CertificateSignTime   time.Duration
}

func (b *EphIDBenchmark) String() string {
	return fmt.Sprintf("%d %d %d %d %d", b.HostIDGenerationTime, b.ExpTimeGenerationTime,
		b.EncryptEphidTime, b.MacComputeTime, b.CertificateSignTime)
}
