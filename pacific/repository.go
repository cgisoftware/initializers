package pacific

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type pacificHttpRepository struct {
	client *http.Client
	logger *slog.Logger
}

// RepositoryOption customiza a criação de PacificHttpRepository.
type RepositoryOption func(*pacificHttpRepository)

// WithLogger define o *slog.Logger usado para registrar as requisições
// enviadas ao Pacific. Padrão: slog.Default().
func WithLogger(logger *slog.Logger) RepositoryOption {
	return func(r *pacificHttpRepository) {
		r.logger = logger
	}
}

// WithHTTPClient permite customizar o *http.Client usado para as
// requisições. Padrão: um cliente dedicado (não o http.DefaultClient) com
// InsecureSkipVerify habilitado, mantendo compatibilidade com o
// comportamento histórico deste pacote para os endpoints Pacific.
func WithHTTPClient(client *http.Client) RepositoryOption {
	return func(r *pacificHttpRepository) {
		r.client = client
	}
}

func newDefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// NewPacificHttpRepository cria um PacificHttpRepository com um *http.Client
// próprio (isolado do http.DefaultClient/DefaultTransport do processo) e um
// logger estruturado para as chamadas realizadas.
func NewPacificHttpRepository(opts ...RepositoryOption) PacificHttpRepository {
	r := &pacificHttpRepository{
		client: newDefaultHTTPClient(),
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Send implements domain.PacificHttpRepository.
func (repository *pacificHttpRepository) Send(ctx context.Context, url string, input PacificInput, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Nunca logar UID/PWD/Params: podem conter credenciais e dados de negócio sensíveis.
	logger := repository.logger.With("url", url, "programa", input.PtoP)

	payload, err := json.Marshal(input)
	if err != nil {
		logger.ErrorContext(ctx, "pacific: falha ao serializar payload", "error", err)
		return nil, PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    "Não foi possível realizar o parse do JSON",
			Err:        err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		logger.ErrorContext(ctx, "pacific: falha ao criar requisição", "error", err)
		return nil, PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    "Não foi possível criar a REQUEST",
			Err:        err,
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, value := range input.Headers {
		req.Header.Set(key, value)
	}

	start := time.Now()
	logger.DebugContext(ctx, "pacific: enviando requisição")

	resp, err := repository.client.Do(req)
	if err != nil {
		logger.ErrorContext(ctx, "pacific: falha ao enviar requisição", "error", err, "duration_ms", time.Since(start).Milliseconds())
		return nil, PacificError{
			StatusCode: http.StatusBadGateway,
			Message:    "Não foi possível enviar a REQUEST",
			Err:        err,
		}
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorContext(ctx, "pacific: falha ao ler o body da resposta", "error", err, "status_code", resp.StatusCode, "duration_ms", duration.Milliseconds())
		return nil, PacificError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Erro ao ler o body da resposta",
			Err:        err,
		}
	}

	if resp.StatusCode != http.StatusOK {
		logger.ErrorContext(ctx, "pacific: status HTTP inesperado", "status_code", resp.StatusCode, "duration_ms", duration.Milliseconds())
		return nil, PacificError{
			StatusCode: resp.StatusCode,
			Message:    "Erro interno do PACIFIC",
			Body:       body,
		}
	}

	if IsResponseErr(body) {
		logger.ErrorContext(ctx, "pacific: resposta contém erro de aplicação apesar do status 200", "duration_ms", duration.Milliseconds())
		return nil, PacificError{
			StatusCode: http.StatusInternalServerError,
			Message:    "PACIFIC retornou erro, mas com status 200",
			Body:       body,
		}
	}

	logger.InfoContext(ctx, "pacific: requisição concluída com sucesso", "status_code", resp.StatusCode, "duration_ms", duration.Milliseconds())

	return body, nil
}
