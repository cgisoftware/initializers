package pacific

import (
	"encoding/json"
)

type PacificInput struct {
	UID    string  `json:"uid"`
	PWD    string  `json:"pwd"`
	PtoP   string  `json:"ptoP"`
	Params []Param `json:"params"`
}

type Param struct {
	Parametro string `json:"parametro"`
	DataType  string `json:"data_type"`
	ParamType string `json:"param_type"`
	Valor     string `json:"valor"`
}

type Dados struct {
	Usuario  string
	Senha    string
	Programa string
	Metodo   string
	Valor    string
	IsGed    bool
}

func NewPacificInput(usuario, senha, programa, metodo, valor string) PacificInput {
	input := PacificInput{
		UID:  usuario,
		PWD:  senha,
		PtoP: programa,
		Params: []Param{
			{
				Parametro: "pcMetodo",
				DataType:  "char",
				ParamType: "input",
				Valor:     metodo,
			},
			{
				Parametro: "pcParametros",
				DataType:  "longchar",
				ParamType: "input",
				Valor:     string(valor),
			},
			{
				Parametro: "pcRetorno",
				DataType:  "longchar",
				ParamType: "output",
				Valor:     "",
			},
		},
	}

	return input
}

type LogErroApp struct {
	LogErroApp []LogErroAppElement `json:"logErroApp"`
}

type LogErroAppElement struct {
	ID   int64  `json:"id"`
	Erro string `json:"erro"`
}

func (log LogErroApp) IsErr() bool {
	return log.LogErroApp[0].Erro != ""
}

func IsResponseErr(body []byte) bool {
	return isLogAppErro(body) || isLogErr001(body)
}

func isLogAppErro(body []byte) bool {
	logErroApp := LogErroApp{}
	json.Unmarshal(body, &logErroApp)
	return len(logErroApp.LogErroApp) > 0
}

func isLogErr001(body []byte) bool {
	logErro001 := LogErr001{}
	json.Unmarshal(body, &logErro001)

	return logErro001.Status == "ERRO"
}

type LogErr001 struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}
