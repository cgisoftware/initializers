package pacific

import "fmt"

// PacificError define um erro com status code HTTP
type PacificError struct {
	StatusCode int
	Body       []byte
	Message    string
	Err        error // erro original (rede, parse, etc), quando houver
}

func (e PacificError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap permite o uso de errors.Is/errors.As sobre o erro original.
func (e PacificError) Unwrap() error {
	return e.Err
}
