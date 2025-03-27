package formatter

import (
	"encoding/json"
	"errors"
	"net/http"
)

type errorAPIError struct {
	status int
	err    error
}

func (e errorAPIError) Error() string {
	return e.err.Error()
}

type HttpResponse struct {
	Message string `json:"message"`
}

func (e errorAPIError) HttpErrorResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	payload, _ := json.Marshal(HttpResponse{
		Message: e.err.Error(),
	})

	http.Error(w, string(payload), e.status)
}

func HttpErrorResponse(w http.ResponseWriter, err error, messages ...string) {
	if err == nil {
		return
	}

	var errorMessage string
	if len(messages) > 0 {
		for i, msg := range messages {
			errorMessage += msg
			if i != len(messages)-1 {
				errorMessage += "\n"
			}
		}
	}

	if !isErrorAPIError(err) {
		if errorMessage != "" {
			err = &errorAPIError{status: http.StatusInternalServerError, err: errors.New(errorMessage)}
		}
	}

	if errorMessage != "" {
		err = &errorAPIError{status: err.(*errorAPIError).status, err: errors.New(errorMessage)}
	}

	err.(*errorAPIError).HttpErrorResponse(w)
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

func isErrorAPIError(err error) bool {
	_, ok := err.(*errorAPIError)
	return ok
}
