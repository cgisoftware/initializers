package validator

import "encoding/json"

type RequestError struct {
	Fields []Field `json:"fields"`
}

type Field struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Errs    string `json:"errs"`
}

func (r RequestError) Error() string {
	payload, _ := json.Marshal(r)
	return string(payload)
}
