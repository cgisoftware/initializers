package pacific

// PacificError define um erro com status code HTTP
type PacificError struct {
	StatusCode int
	Body       []byte
	Message    string
}

func (e *PacificError) Error() string {
	return e.Message
}
