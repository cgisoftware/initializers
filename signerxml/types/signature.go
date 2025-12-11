package types

import "time"

// Signature represents the data necessary to sign an XML
type Signature struct {
	XMLContent   string `json:"xml_content"`
	Certificate  A1     `json:"certificate"`
	ElementID    string `json:"element_id,omitempty"`
	TagSignature string `json:"tag_signature,omitempty"`
}

// SignatureResult represents the result of the signature
type SignatureResult struct {
	XMLSigned string    `json:"xml_signed"`
	Success   bool      `json:"success"`
	Error     error     `json:"error,omitempty"`
	Date      time.Time `json:"date"`
}
