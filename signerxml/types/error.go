package types

type SignerXmlError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e SignerXmlError) Error() string {
	return e.Message
}

var (
	ErrInvalidInput       = SignerXmlError{Message: "Invalid input data", Code: 400}
	ErrInvalidCertificate = SignerXmlError{Message: "Invalid certificate", Code: 400}
)
