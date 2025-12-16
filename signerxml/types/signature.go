package types

import (
	"crypto"
	"crypto/x509"
	"time"
)

// Signature represents the data necessary to sign an XML
type Signature struct {
	XMLContent   string            `json:"xml_content"`
	Certificate  *x509.Certificate `json:"certificate"`
	PrivateKey   crypto.PrivateKey `json:"private_key"`
	ElementID    string            `json:"element_id,omitempty"`
	TagSignature string            `json:"tag_signature,omitempty"`
}

// SignatureResult represents the result of the signature
type SignatureResult struct {
	XMLSigned  string    `json:"xml_signed"`
	XMLContent string    `json:"xml_content"`
	Success    bool      `json:"success"`
	Error      error     `json:"error,omitempty"`
	Date       time.Time `json:"date"`
}
