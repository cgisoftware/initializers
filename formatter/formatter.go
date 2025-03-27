package formatter

import (
	"errors"
	"fmt"
	"net/http"
)

type errorAPIError struct {
	status int
	err    error
}

func (e errorAPIError) Error() string {
	return e.err.Error()
}

func (e errorAPIError) HttpError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	payload := fmt.Sprintf(`{"message": "%s"}`, e.err)
	http.Error(w, payload, e.status)
}

var (
	ErrAuth                = &errorAPIError{status: http.StatusUnauthorized, err: errors.New("não autorizado")}
	ErrNotFound            = &errorAPIError{status: http.StatusNotFound, err: errors.New("não existe")}
	ErrDuplicate           = &errorAPIError{status: http.StatusBadRequest, err: errors.New("duplicado")}
	ErrInternalServer      = &errorAPIError{status: http.StatusInternalServerError, err: errors.New("erro no servidor")}
	ErrBadRequest          = &errorAPIError{status: http.StatusBadRequest, err: errors.New("preencha os campos")}
	ErrIDNotFound          = &errorAPIError{status: http.StatusBadRequest, err: errors.New("é necessário informar o id e precisa ser um número válido")}
	ErrAPITokenKeyNotFound = &errorAPIError{status: http.StatusBadRequest, err: errors.New("é necessário informar a chave do cliente e o token da api")}
	ErrCodMenuKeyNotFound  = &errorAPIError{status: http.StatusBadRequest, err: errors.New("é necessário informar a chave do cliente e o codigo do menu")}
)

func IsErrorAPIError(err error) bool {
	_, ok := err.(*errorAPIError)
	return ok
}
