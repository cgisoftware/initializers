package pacific

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type pacificHttpRepository struct{}

// Send implements domain.PacificHttpRepository.
func (repository *pacificHttpRepository) Send(ctx context.Context, url string, input PacificInput) ([]byte, *PacificError) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	http.DefaultTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	payload, err := json.Marshal(input)

	if err != nil {
		return nil, &PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    "failed to parse json",
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))

	if err != nil {
		return nil, &PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    "failed to create request",
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    "failed to send request",
		}
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, &PacificError{
			StatusCode: resp.StatusCode,
			Message:    "pacific internal error",
			Body:       body,
		}
	}

	if IsResponseErr(body) {
		return nil, &PacificError{
			StatusCode: http.StatusInternalServerError,
			Message:    "error returned from api with 200 status code",
			Body:       body,
		}
	}

	return body, nil
}

func NewPacificHttpRepository() PacificHttpRepository {
	return &pacificHttpRepository{}
}
