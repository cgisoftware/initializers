package pacific

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/cgisoftware/initializers/opentelemetry"
)

type pacificHttpRepository struct{}

// Send implements domain.PacificHttpRepository.
func (repository *pacificHttpRepository) Send(ctx context.Context, url string, input PacificInput) ([]byte, *PacificError) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	ctx, span := opentelemetry.StartTracing(ctx, "paficicHttpRepository.Send")
	defer span.End()

	http.DefaultTransport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	payload, err := json.Marshal(input)

	if err != nil {
		errMessage := "Não foi possível realizar o parse do JSON"
		opentelemetry.ErrorLog(ctx, errMessage, err)
		return nil, &PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    errMessage,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))

	if err != nil {
		errMessage := "Não foi possível criar a REQUEST"
		opentelemetry.ErrorLog(ctx, errMessage, err)
		return nil, &PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    errMessage,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errMessage := "Não foi possível enviar a REQUEST"
		opentelemetry.ErrorLog(ctx, errMessage, err, opentelemetry.WithHttpLog(opentelemetry.NewHttpLog(
			req, nil, http.StatusBadGateway,
		)))
		return nil, &PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    errMessage,
		}
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		errMessage := "Erro interno do PACIFIC"
		opentelemetry.ErrorLog(ctx, errMessage, err, opentelemetry.WithHttpLog(opentelemetry.NewHttpLog(
			req, body, int64(resp.StatusCode),
		)))
		return nil, &PacificError{
			StatusCode: resp.StatusCode,
			Message:    errMessage,
			Body:       body,
		}
	}

	if IsResponseErr(body) {
		errMessage := "PACIFIC retornou erro, mas com status 200"
		opentelemetry.ErrorLog(ctx, errMessage, err, opentelemetry.WithHttpLog(opentelemetry.NewHttpLog(
			req, body, http.StatusInternalServerError,
		)))
		return nil, &PacificError{
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
