package postgres

import "net/http"

type PostgresError struct {
	Message    string
	StatusCode int
}

func (e *PostgresError) Error() string {
	return e.Message
}

var ErrNotFound = &PostgresError{"Recurso não encontrado", http.StatusNotFound}
var ErrInternal = &PostgresError{"Erro interno no servidor", http.StatusInternalServerError}
