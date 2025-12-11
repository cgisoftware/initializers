package types

import (
	"crypto/x509"
	"time"
)

type A1 struct {
	File     []byte `json:"file"`
	Password string `json:"password"`
}

// InfoCertificado contains information extracted from the certificate
type A1Info struct {
	Subject      string            `json:"subject"`
	Issuer       string            `json:"issuer"`
	SerialNumber string            `json:"serial_number"`
	NotBefore    time.Time         `json:"not_before"`
	NotAfter     time.Time         `json:"not_after"`
	Certificate  *x509.Certificate `json:"-"`
}
