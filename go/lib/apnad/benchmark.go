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

type DNSReplyBenchmark struct {
	ReplyTime time.Duration
}

func (b *DNSReplyBenchmark) String() string {
	return fmt.Sprintf("%d", b.ReplyTime)
}

type DNSRegisterBenchmark struct {
	RegisterTime time.Duration
}

func (b *DNSRegisterBenchmark) String() string {
	return fmt.Sprintf("%d", b.RegisterTime)
}

type BorderRouterBenchmark struct {
	EphidDecryptionTime time.Duration
	SiphashToHost       time.Duration
	MacVerificationTime time.Duration
}

func (b *BorderRouterBenchmark) String() string {
	return fmt.Sprintf("%d %d %d", b.EphidDecryptionTime, b.SiphashToHost, b.MacVerificationTime)
}
