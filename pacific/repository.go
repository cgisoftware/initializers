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
func (repository *pacificHttpRepository) Send(ctx context.Context, url string, input PacificInput, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	http.DefaultTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	payload, err := json.Marshal(input)

	if err != nil {
		errMessage := "Não foi possível realizar o parse do JSON"
		return nil, PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    errMessage,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))

	if err != nil {
		errMessage := "Não foi possível criar a REQUEST"
		return nil, PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    errMessage,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errMessage := "Não foi possível enviar a REQUEST"
		return nil, PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    errMessage,
		}
	}

	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			errMessage := "Erro ao ler o body da resposta"
			return nil, PacificError{
				StatusCode: http.StatusInternalServerError,
				Message:    errMessage,
			}
		}

		errMessage := "Erro interno do PACIFIC"
		return nil, PacificError{
			StatusCode: resp.StatusCode,
			Message:    errMessage,
			Body:       body,
		}
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		errMessage := "Erro ao ler o body da resposta"
		return nil, PacificError{
			StatusCode: http.StatusInternalServerError,
			Message:    errMessage,
			Body:       nil,
		}
	}

	if IsResponseErr(body) {
		errMessage := "PACIFIC retornou erro, mas com status 200"
		return nil, PacificError{
			StatusCode: http.StatusInternalServerError,
			Message:    errMessage,
			Body:       body,
		}
	}

	return body, nil
}

func NewPacificHttpRepository() PacificHttpRepository {
	return &pacificHttpRepository{}
}
