# Pacote Pacific

O pacote `pacific` fornece funcionalidades para integração com sistemas Pacific, incluindo estruturas de dados padronizadas, tratamento de erros específicos e utilitários para comunicação com APIs Pacific.

## Funcionalidades

### 📊 Estruturas de Dados
- `PacificInput`: Estrutura principal para entrada de dados
- `Dados`: Informações de usuário e programa
- `Param`: Parâmetros configuráveis
- Suporte a diferentes tipos de entrada

### 🔐 Autenticação
- Suporte a senhas de usuário
- Senhas de colaborador
- Validação de credenciais
- Integração com sistemas de autenticação

### 🚨 Tratamento de Erros
- Erros específicos do Pacific
- Logging estruturado
- Códigos de erro padronizados
- Respostas de erro formatadas

### 🔧 Utilitários
- Validação de dados
- Formatação de parâmetros
- Helpers para integração
- Debugging facilitado

## Estruturas Principais

### `PacificInput`
```go
type PacificInput struct {
    Dados  Dados   `json:"dados"`
    Params []Param `json:"params"`
}
```

Estrutura principal para entrada de dados no sistema Pacific.

### `Dados`
```go
type Dados struct {
    Usuario  string `json:"usuario"`
    Senha    string `json:"senha"`
    Programa string `json:"programa"`
}
```

Informações de autenticação e identificação do programa.

### `Param`
```go
type Param struct {
    Nome  string      `json:"nome"`
    Valor interface{} `json:"valor"`
    Tipo  string      `json:"tipo,omitempty"`
}
```

Parâmetro configurável com nome, valor e tipo opcional.

### `LogErroApp`
```go
type LogErroApp struct {
    Codigo    string `json:"codigo"`
    Mensagem  string `json:"mensagem"`
    Detalhes  string `json:"detalhes,omitempty"`
    Timestamp string `json:"timestamp"`
}
```

Estrutura para logging de erros da aplicação.

## Uso Básico

### Criação de PacificInput Simples
```go
package main

import (
    "fmt"
    "seu-projeto/initializers/pacific"
)

func main() {
    // Criar entrada básica
    input := pacific.NewPacificInput("usuario123", "senha456", "PROGRAMA_TESTE")
    
    fmt.Printf("Input criado: %+v\n", input)
    
    // Adicionar parâmetros
    input.Params = append(input.Params, pacific.Param{
        Nome:  "parametro1",
        Valor: "valor1",
        Tipo:  "string",
    })
    
    input.Params = append(input.Params, pacific.Param{
        Nome:  "parametro2",
        Valor: 123,
        Tipo:  "int",
    })
    
    fmt.Printf("Input com parâmetros: %+v\n", input)
}
```

### Criação com Senha de Colaborador
```go
func exemploColaborador() {
    // Criar entrada com senha de colaborador
    input := pacific.NewPacificInputColab("colaborador", "senhaColab", "SISTEMA_ADMIN")
    
    // Adicionar parâmetros específicos
    params := []pacific.Param{
        {
            Nome:  "nivel_acesso",
            Valor: "admin",
            Tipo:  "string",
        },
        {
            Nome:  "departamento",
            Valor: "TI",
            Tipo:  "string",
        },
        {
            Nome:  "ativo",
            Valor: true,
            Tipo:  "boolean",
        },
    }
    
    input.Params = params
    
    fmt.Printf("Input colaborador: %+v\n", input)
}
```

## Tratamento de Erros

### Logging de Erros
```go
package services

import (
    "encoding/json"
    "log"
    "time"
    "seu-projeto/initializers/pacific"
)

func processarRequisicao(input *pacific.PacificInput) error {
    // Validar entrada
    if input.Dados.Usuario == "" {
        erro := &pacific.LogErroApp{
            Codigo:    "USUARIO_VAZIO",
            Mensagem:  "Usuário não pode estar vazio",
            Detalhes:  "Campo 'usuario' é obrigatório para autenticação",
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        pacific.LogErr001(erro)
        return fmt.Errorf("erro de validação: %s", erro.Mensagem)
    }
    
    // Validar programa
    if input.Dados.Programa == "" {
        erro := &pacific.LogErroApp{
            Codigo:    "PROGRAMA_VAZIO",
            Mensagem:  "Programa não pode estar vazio",
            Detalhes:  "Campo 'programa' é obrigatório para identificação",
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        pacific.LogErr001(erro)
        return fmt.Errorf("erro de validação: %s", erro.Mensagem)
    }
    
    // Processar requisição
    err := chamarAPIPacific(input)
    if err != nil {
        erro := &pacific.LogErroApp{
            Codigo:    "API_ERROR",
            Mensagem:  "Erro na comunicação com API Pacific",
            Detalhes:  fmt.Sprintf("Erro original: %v", err),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        pacific.LogErr001(erro)
        return err
    }
    
    return nil
}
```

### Verificação de Erros de Resposta
```go
func verificarResposta(response interface{}) bool {
    // Verificar se a resposta contém erro
    isError := pacific.IsResponseErr(response)
    
    if isError {
        log.Printf("Resposta contém erro: %+v", response)
        
        // Log do erro
        erro := &pacific.LogErroApp{
            Codigo:    "RESPONSE_ERROR",
            Mensagem:  "Resposta da API contém erro",
            Detalhes:  fmt.Sprintf("Response: %+v", response),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        
        pacific.LogErr001(erro)
        return false
    }
    
    log.Println("Resposta processada com sucesso")
    return true
}
```

## Manipulação de Parâmetros

### Diferentes Tipos de Parâmetros
```go
func exemploParametros() {
    input := pacific.NewPacificInput("user", "pass", "PROGRAMA")
    
    // Parâmetro string
    input.Params = append(input.Params, pacific.Param{
        Nome:  "nome_cliente",
        Valor: "João Silva",
        Tipo:  "string",
    })
    
    // Parâmetro numérico
    input.Params = append(input.Params, pacific.Param{
        Nome:  "idade",
        Valor: 30,
        Tipo:  "int",
    })
    
    // Parâmetro decimal
    input.Params = append(input.Params, pacific.Param{
        Nome:  "salario",
        Valor: 5500.50,
        Tipo:  "float",
    })
    
    // Parâmetro booleano
    input.Params = append(input.Params, pacific.Param{
        Nome:  "ativo",
        Valor: true,
        Tipo:  "boolean",
    })
    
    // Parâmetro data
    input.Params = append(input.Params, pacific.Param{
        Nome:  "data_nascimento",
        Valor: "1990-05-15",
        Tipo:  "date",
    })
    
    // Parâmetro array
    input.Params = append(input.Params, pacific.Param{
        Nome:  "telefones",
        Valor: []string{"11999999999", "1133333333"},
        Tipo:  "array",
    })
    
    // Parâmetro objeto
    endereco := map[string]interface{}{
        "rua":    "Rua das Flores, 123",
        "cidade": "São Paulo",
        "cep":    "01234-567",
    }
    
    input.Params = append(input.Params, pacific.Param{
        Nome:  "endereco",
        Valor: endereco,
        Tipo:  "object",
    })
    
    fmt.Printf("Input com parâmetros variados: %+v\n", input)
}
```

### Helpers para Parâmetros
```go
package helpers

import "seu-projeto/initializers/pacific"

// AddStringParam adiciona parâmetro string
func AddStringParam(input *pacific.PacificInput, nome, valor string) {
    input.Params = append(input.Params, pacific.Param{
        Nome:  nome,
        Valor: valor,
        Tipo:  "string",
    })
}

// AddIntParam adiciona parâmetro inteiro
func AddIntParam(input *pacific.PacificInput, nome string, valor int) {
    input.Params = append(input.Params, pacific.Param{
        Nome:  nome,
        Valor: valor,
        Tipo:  "int",
    })
}

// AddFloatParam adiciona parâmetro decimal
func AddFloatParam(input *pacific.PacificInput, nome string, valor float64) {
    input.Params = append(input.Params, pacific.Param{
        Nome:  nome,
        Valor: valor,
        Tipo:  "float",
    })
}

// AddBoolParam adiciona parâmetro booleano
func AddBoolParam(input *pacific.PacificInput, nome string, valor bool) {
    input.Params = append(input.Params, pacific.Param{
        Nome:  nome,
        Valor: valor,
        Tipo:  "boolean",
    })
}

// FindParam encontra parâmetro por nome
func FindParam(input *pacific.PacificInput, nome string) *pacific.Param {
    for i := range input.Params {
        if input.Params[i].Nome == nome {
            return &input.Params[i]
        }
    }
    return nil
}

// RemoveParam remove parâmetro por nome
func RemoveParam(input *pacific.PacificInput, nome string) bool {
    for i, param := range input.Params {
        if param.Nome == nome {
            input.Params = append(input.Params[:i], input.Params[i+1:]...)
            return true
        }
    }
    return false
}

// UpdateParam atualiza valor de parâmetro existente
func UpdateParam(input *pacific.PacificInput, nome string, novoValor interface{}) bool {
    param := FindParam(input, nome)
    if param != nil {
        param.Valor = novoValor
        return true
    }
    return false
}

// Uso dos helpers
func exemploHelpers() {
    input := pacific.NewPacificInput("user", "pass", "PROGRAMA")
    
    // Adicionar parâmetros usando helpers
    AddStringParam(input, "nome", "Maria")
    AddIntParam(input, "idade", 25)
    AddFloatParam(input, "salario", 3500.00)
    AddBoolParam(input, "ativo", true)
    
    // Buscar parâmetro
    nomeParam := FindParam(input, "nome")
    if nomeParam != nil {
        fmt.Printf("Nome encontrado: %v\n", nomeParam.Valor)
    }
    
    // Atualizar parâmetro
    UpdateParam(input, "idade", 26)
    
    // Remover parâmetro
    RemoveParam(input, "salario")
    
    fmt.Printf("Input após modificações: %+v\n", input)
}
```

## Integração com APIs

### Cliente HTTP para Pacific
```go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "seu-projeto/initializers/pacific"
)

type PacificClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}

func NewPacificClient(baseURL string, timeout time.Duration) *PacificClient {
    return &PacificClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: timeout,
        },
        timeout: timeout,
    }
}

func (pc *PacificClient) ExecuteRequest(ctx context.Context, input *pacific.PacificInput) (interface{}, error) {
    // Serializar input
    jsonData, err := json.Marshal(input)
    if err != nil {
        erro := &pacific.LogErroApp{
            Codigo:    "JSON_MARSHAL_ERROR",
            Mensagem:  "Erro ao serializar dados de entrada",
            Detalhes:  fmt.Sprintf("Erro: %v", err),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        pacific.LogErr001(erro)
        return nil, fmt.Errorf("erro de serialização: %v", err)
    }
    
    // Criar requisição
    req, err := http.NewRequestWithContext(ctx, "POST", pc.baseURL+"/execute", bytes.NewBuffer(jsonData))
    if err != nil {
        erro := &pacific.LogErroApp{
            Codigo:    "REQUEST_CREATE_ERROR",
            Mensagem:  "Erro ao criar requisição HTTP",
            Detalhes:  fmt.Sprintf("Erro: %v", err),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        pacific.LogErr001(erro)
        return nil, fmt.Errorf("erro ao criar requisição: %v", err)
    }
    
    // Headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("User-Agent", "Pacific-Client/1.0")
    
    // Executar requisição
    resp, err := pc.httpClient.Do(req)
    if err != nil {
        erro := &pacific.LogErroApp{
            Codigo:    "HTTP_REQUEST_ERROR",
            Mensagem:  "Erro na requisição HTTP",
            Detalhes:  fmt.Sprintf("URL: %s, Erro: %v", req.URL.String(), err),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        pacific.LogErr001(erro)
        return nil, fmt.Errorf("erro na requisição: %v", err)
    }
    defer resp.Body.Close()
    
    // Verificar status code
    if resp.StatusCode != http.StatusOK {
        erro := &pacific.LogErroApp{
            Codigo:    "HTTP_STATUS_ERROR",
            Mensagem:  "Status HTTP inválido",
            Detalhes:  fmt.Sprintf("Status: %d, URL: %s", resp.StatusCode, req.URL.String()),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        pacific.LogErr001(erro)
        return nil, fmt.Errorf("status HTTP inválido: %d", resp.StatusCode)
    }
    
    // Decodificar resposta
    var response interface{}
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        erro := &pacific.LogErroApp{
            Codigo:    "JSON_DECODE_ERROR",
            Mensagem:  "Erro ao decodificar resposta JSON",
            Detalhes:  fmt.Sprintf("Erro: %v", err),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        pacific.LogErr001(erro)
        return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
    }
    
    // Verificar se resposta contém erro
    if pacific.IsResponseErr(response) {
        erro := &pacific.LogErroApp{
            Codigo:    "PACIFIC_API_ERROR",
            Mensagem:  "API Pacific retornou erro",
            Detalhes:  fmt.Sprintf("Response: %+v", response),
            Timestamp: time.Now().Format(time.RFC3339),
        }
        pacific.LogErr001(erro)
        return response, fmt.Errorf("erro da API Pacific")
    }
    
    return response, nil
}

// Método específico para consultas
func (pc *PacificClient) Query(ctx context.Context, usuario, senha, programa string, params []pacific.Param) (interface{}, error) {
    input := &pacific.PacificInput{
        Dados: pacific.Dados{
            Usuario:  usuario,
            Senha:    senha,
            Programa: programa,
        },
        Params: params,
    }
    
    return pc.ExecuteRequest(ctx, input)
}

// Método para operações de colaborador
func (pc *PacificClient) ColabOperation(ctx context.Context, colaborador, senhaColab, programa string, params []pacific.Param) (interface{}, error) {
    input := pacific.NewPacificInputColab(colaborador, senhaColab, programa)
    input.Params = params
    
    return pc.ExecuteRequest(ctx, input)
}
```

### Exemplo de Uso do Cliente
```go
func exemploCliente() {
    ctx := context.Background()
    
    // Criar cliente
    client := NewPacificClient("https://api.pacific.com", time.Second*30)
    
    // Parâmetros da consulta
    params := []pacific.Param{
        {
            Nome:  "codigo_cliente",
            Valor: "12345",
            Tipo:  "string",
        },
        {
            Nome:  "incluir_historico",
            Valor: true,
            Tipo:  "boolean",
        },
    }
    
    // Executar consulta
    response, err := client.Query(ctx, "usuario", "senha", "CONSULTA_CLIENTE", params)
    if err != nil {
        log.Printf("Erro na consulta: %v", err)
        return
    }
    
    fmt.Printf("Resposta da consulta: %+v\n", response)
}
```

## Validação de Dados

### Validadores
```go
package validators

import (
    "errors"
    "fmt"
    "regexp"
    "strings"
    "seu-projeto/initializers/pacific"
)

// ValidatePacificInput valida estrutura de entrada
func ValidatePacificInput(input *pacific.PacificInput) error {
    if input == nil {
        return errors.New("input não pode ser nil")
    }
    
    // Validar dados
    if err := ValidateDados(&input.Dados); err != nil {
        return fmt.Errorf("erro nos dados: %v", err)
    }
    
    // Validar parâmetros
    if err := ValidateParams(input.Params); err != nil {
        return fmt.Errorf("erro nos parâmetros: %v", err)
    }
    
    return nil
}

// ValidateDados valida estrutura de dados
func ValidateDados(dados *pacific.Dados) error {
    if dados.Usuario == "" {
        return errors.New("usuário é obrigatório")
    }
    
    if dados.Senha == "" {
        return errors.New("senha é obrigatória")
    }
    
    if dados.Programa == "" {
        return errors.New("programa é obrigatório")
    }
    
    // Validar formato do usuário
    if len(dados.Usuario) < 3 {
        return errors.New("usuário deve ter pelo menos 3 caracteres")
    }
    
    // Validar formato do programa
    programaRegex := regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
    if !programaRegex.MatchString(dados.Programa) {
        return errors.New("programa deve conter apenas letras maiúsculas, números e underscore")
    }
    
    return nil
}

// ValidateParams valida lista de parâmetros
func ValidateParams(params []pacific.Param) error {
    nomes := make(map[string]bool)
    
    for i, param := range params {
        // Validar nome
        if param.Nome == "" {
            return fmt.Errorf("parâmetro %d: nome é obrigatório", i)
        }
        
        // Verificar duplicatas
        if nomes[param.Nome] {
            return fmt.Errorf("parâmetro duplicado: %s", param.Nome)
        }
        nomes[param.Nome] = true
        
        // Validar formato do nome
        nomeRegex := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
        if !nomeRegex.MatchString(param.Nome) {
            return fmt.Errorf("parâmetro %s: nome deve conter apenas letras minúsculas, números e underscore", param.Nome)
        }
        
        // Validar valor
        if param.Valor == nil {
            return fmt.Errorf("parâmetro %s: valor não pode ser nil", param.Nome)
        }
        
        // Validar tipo se especificado
        if param.Tipo != "" {
            if err := ValidateParamType(param); err != nil {
                return fmt.Errorf("parâmetro %s: %v", param.Nome, err)
            }
        }
    }
    
    return nil
}

// ValidateParamType valida tipo do parâmetro
func ValidateParamType(param pacific.Param) error {
    switch param.Tipo {
    case "string":
        if _, ok := param.Valor.(string); !ok {
            return fmt.Errorf("valor deve ser string, recebido: %T", param.Valor)
        }
    case "int":
        switch param.Valor.(type) {
        case int, int32, int64:
            // OK
        case float64:
            // JSON unmarshaling pode converter int para float64
            if v := param.Valor.(float64); v != float64(int(v)) {
                return errors.New("valor deve ser inteiro")
            }
        default:
            return fmt.Errorf("valor deve ser inteiro, recebido: %T", param.Valor)
        }
    case "float":
        switch param.Valor.(type) {
        case float32, float64, int, int32, int64:
            // OK
        default:
            return fmt.Errorf("valor deve ser numérico, recebido: %T", param.Valor)
        }
    case "boolean":
        if _, ok := param.Valor.(bool); !ok {
            return fmt.Errorf("valor deve ser boolean, recebido: %T", param.Valor)
        }
    case "array":
        switch param.Valor.(type) {
        case []interface{}, []string, []int, []float64:
            // OK
        default:
            return fmt.Errorf("valor deve ser array, recebido: %T", param.Valor)
        }
    case "object":
        if _, ok := param.Valor.(map[string]interface{}); !ok {
            return fmt.Errorf("valor deve ser object, recebido: %T", param.Valor)
        }
    case "date":
        if dateStr, ok := param.Valor.(string); ok {
            if !isValidDate(dateStr) {
                return errors.New("formato de data inválido, use YYYY-MM-DD")
            }
        } else {
            return fmt.Errorf("data deve ser string, recebido: %T", param.Valor)
        }
    default:
        return fmt.Errorf("tipo não suportado: %s", param.Tipo)
    }
    
    return nil
}

// isValidDate verifica formato de data
func isValidDate(dateStr string) bool {
    dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
    return dateRegex.MatchString(dateStr)
}

// Exemplo de uso
func exemploValidacao() {
    input := pacific.NewPacificInput("user123", "pass456", "PROGRAMA_TESTE")
    
    // Adicionar parâmetros
    input.Params = []pacific.Param{
        {
            Nome:  "nome_cliente",
            Valor: "João Silva",
            Tipo:  "string",
        },
        {
            Nome:  "idade",
            Valor: 30,
            Tipo:  "int",
        },
    }
    
    // Validar
    if err := ValidatePacificInput(input); err != nil {
        log.Printf("Erro de validação: %v", err)
        return
    }
    
    fmt.Println("Input válido!")
}
```

## Serialização e Deserialização

### JSON Handling
```go
package serialization

import (
    "encoding/json"
    "fmt"
    "seu-projeto/initializers/pacific"
)

// ToJSON converte PacificInput para JSON
func ToJSON(input *pacific.PacificInput) ([]byte, error) {
    return json.MarshalIndent(input, "", "  ")
}

// FromJSON converte JSON para PacificInput
func FromJSON(data []byte) (*pacific.PacificInput, error) {
    var input pacific.PacificInput
    err := json.Unmarshal(data, &input)
    if err != nil {
        return nil, fmt.Errorf("erro ao deserializar JSON: %v", err)
    }
    return &input, nil
}

// ToJSONString converte para string JSON
func ToJSONString(input *pacific.PacificInput) (string, error) {
    data, err := ToJSON(input)
    if err != nil {
        return "", err
    }
    return string(data), nil
}

// FromJSONString converte string JSON para PacificInput
func FromJSONString(jsonStr string) (*pacific.PacificInput, error) {
    return FromJSON([]byte(jsonStr))
}

// Exemplo de uso
func exemploSerialization() {
    // Criar input
    input := pacific.NewPacificInput("usuario", "senha", "PROGRAMA")
    input.Params = []pacific.Param{
        {
            Nome:  "parametro1",
            Valor: "valor1",
            Tipo:  "string",
        },
    }
    
    // Converter para JSON
    jsonData, err := ToJSON(input)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("JSON:\n%s\n", string(jsonData))
    
    // Converter de volta
    inputFromJSON, err := FromJSON(jsonData)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Input deserializado: %+v\n", inputFromJSON)
}
```

## Middleware e Integração Web

### Middleware Gin
```go
package middleware

import (
    "encoding/json"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "seu-projeto/initializers/pacific"
)

// PacificMiddleware middleware para requisições Pacific
func PacificMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Verificar Content-Type
        if c.GetHeader("Content-Type") != "application/json" {
            erro := &pacific.LogErroApp{
                Codigo:    "INVALID_CONTENT_TYPE",
                Mensagem:  "Content-Type deve ser application/json",
                Timestamp: time.Now().Format(time.RFC3339),
            }
            pacific.LogErr001(erro)
            
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Content-Type inválido",
            })
            c.Abort()
            return
        }
        
        // Parse do body
        var input pacific.PacificInput
        if err := c.ShouldBindJSON(&input); err != nil {
            erro := &pacific.LogErroApp{
                Codigo:    "JSON_PARSE_ERROR",
                Mensagem:  "Erro ao fazer parse do JSON",
                Detalhes:  err.Error(),
                Timestamp: time.Now().Format(time.RFC3339),
            }
            pacific.LogErr001(erro)
            
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "JSON inválido",
            })
            c.Abort()
            return
        }
        
        // Validar input
        if err := ValidatePacificInput(&input); err != nil {
            erro := &pacific.LogErroApp{
                Codigo:    "VALIDATION_ERROR",
                Mensagem:  "Erro de validação",
                Detalhes:  err.Error(),
                Timestamp: time.Now().Format(time.RFC3339),
            }
            pacific.LogErr001(erro)
            
            c.JSON(http.StatusBadRequest, gin.H{
                "error": err.Error(),
            })
            c.Abort()
            return
        }
        
        // Adicionar input ao contexto
        c.Set("pacific_input", &input)
        
        c.Next()
    }
}

// Handler de exemplo
func PacificHandler(c *gin.Context) {
    // Obter input do contexto
    inputInterface, exists := c.Get("pacific_input")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Input não encontrado no contexto",
        })
        return
    }
    
    input := inputInterface.(*pacific.PacificInput)
    
    // Processar requisição
    response, err := processarRequisicaoPacific(input)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Erro no processamento",
        })
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// Configuração das rotas
func setupRoutes() {
    r := gin.Default()
    
    // Aplicar middleware
    r.Use(PacificMiddleware())
    
    // Rotas
    r.POST("/pacific/execute", PacificHandler)
    r.POST("/pacific/query", PacificQueryHandler)
    r.POST("/pacific/colab", PacificColabHandler)
    
    r.Run(":8080")
}
```

## Testes

### Testes Unitários
```go
package pacific_test

import (
    "encoding/json"
    "testing"
    "seu-projeto/initializers/pacific"
)

func TestNewPacificInput(t *testing.T) {
    input := pacific.NewPacificInput("user", "pass", "PROGRAMA")
    
    if input.Dados.Usuario != "user" {
        t.Errorf("Usuário esperado: user, obtido: %s", input.Dados.Usuario)
    }
    
    if input.Dados.Senha != "pass" {
        t.Errorf("Senha esperada: pass, obtida: %s", input.Dados.Senha)
    }
    
    if input.Dados.Programa != "PROGRAMA" {
        t.Errorf("Programa esperado: PROGRAMA, obtido: %s", input.Dados.Programa)
    }
    
    if len(input.Params) != 0 {
        t.Errorf("Parâmetros deveriam estar vazios, obtido: %d", len(input.Params))
    }
}

func TestNewPacificInputColab(t *testing.T) {
    input := pacific.NewPacificInputColab("colab", "senhaColab", "SISTEMA")
    
    if input.Dados.Usuario != "colab" {
        t.Errorf("Colaborador esperado: colab, obtido: %s", input.Dados.Usuario)
    }
    
    if input.Dados.Senha != "senhaColab" {
        t.Errorf("Senha esperada: senhaColab, obtida: %s", input.Dados.Senha)
    }
}

func TestJSONSerialization(t *testing.T) {
    input := pacific.NewPacificInput("user", "pass", "PROGRAMA")
    input.Params = []pacific.Param{
        {
            Nome:  "teste",
            Valor: "valor",
            Tipo:  "string",
        },
    }
    
    // Serializar
    jsonData, err := json.Marshal(input)
    if err != nil {
        t.Fatal("Erro na serialização:", err)
    }
    
    // Deserializar
    var inputFromJSON pacific.PacificInput
    err = json.Unmarshal(jsonData, &inputFromJSON)
    if err != nil {
        t.Fatal("Erro na deserialização:", err)
    }
    
    // Verificar
    if inputFromJSON.Dados.Usuario != input.Dados.Usuario {
        t.Error("Usuário não coincide após serialização")
    }
    
    if len(inputFromJSON.Params) != len(input.Params) {
        t.Error("Número de parâmetros não coincide")
    }
}

func TestLogErroApp(t *testing.T) {
    erro := &pacific.LogErroApp{
        Codigo:    "TEST_ERROR",
        Mensagem:  "Erro de teste",
        Detalhes:  "Detalhes do erro",
        Timestamp: "2023-01-01T00:00:00Z",
    }
    
    // Testar serialização
    jsonData, err := json.Marshal(erro)
    if err != nil {
        t.Fatal("Erro na serialização do erro:", err)
    }
    
    // Verificar se contém campos esperados
    var erroMap map[string]interface{}
    err = json.Unmarshal(jsonData, &erroMap)
    if err != nil {
        t.Fatal("Erro na deserialização:", err)
    }
    
    if erroMap["codigo"] != "TEST_ERROR" {
        t.Error("Código do erro não coincide")
    }
}

func BenchmarkNewPacificInput(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _ = pacific.NewPacificInput("user", "pass", "PROGRAMA")
    }
}

func BenchmarkJSONMarshal(b *testing.B) {
    input := pacific.NewPacificInput("user", "pass", "PROGRAMA")
    input.Params = []pacific.Param{
        {Nome: "param1", Valor: "valor1", Tipo: "string"},
        {Nome: "param2", Valor: 123, Tipo: "int"},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := json.Marshal(input)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Configuração e Deployment

### Variáveis de Ambiente
```bash
# Configurações Pacific
export PACIFIC_API_URL="https://api.pacific.com"
export PACIFIC_TIMEOUT="30s"
export PACIFIC_RETRY_COUNT="3"
export PACIFIC_LOG_LEVEL="info"

# Credenciais (usar secrets em produção)
export PACIFIC_DEFAULT_USER="sistema"
export PACIFIC_DEFAULT_PROGRAM="DEFAULT_PROGRAM"
```

### Configuração via Arquivo
```yaml
# config/pacific.yaml
pacific:
  api:
    url: "https://api.pacific.com"
    timeout: "30s"
    retry_count: 3
  
  logging:
    level: "info"
    format: "json"
    output: "stdout"
  
  defaults:
    program: "DEFAULT_PROGRAM"
    timeout: "10s"
  
  validation:
    strict_mode: true
    required_fields: ["usuario", "senha", "programa"]
```

## Melhores Práticas

### 1. Segurança
```go
// ✅ Não log senhas
func logSafeInput(input *pacific.PacificInput) {
    safeInput := *input
    safeInput.Dados.Senha = "***"
    log.Printf("Input: %+v", safeInput)
}

// ✅ Validar entrada
func processInput(input *pacific.PacificInput) error {
    if err := ValidatePacificInput(input); err != nil {
        return err
    }
    // ... processar
}

// ❌ Evitar log de dados sensíveis
// log.Printf("Input completo: %+v", input) // Pode expor senhas
```

### 2. Performance
```go
// ✅ Reutilizar clientes HTTP
var pacificClient = NewPacificClient("https://api.pacific.com", time.Second*30)

// ✅ Pool de conexões
func setupHTTPClient() *http.Client {
    return &http.Client{
        Timeout: time.Second * 30,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
        },
    }
}
```

### 3. Tratamento de Erros
```go
// ✅ Erros específicos e informativos
func validateUser(usuario string) error {
    if usuario == "" {
        return &pacific.LogErroApp{
            Codigo:    "USUARIO_VAZIO",
            Mensagem:  "Usuário é obrigatório",
            Timestamp: time.Now().Format(time.RFC3339),
        }
    }
    return nil
}

// ✅ Log estruturado
func logError(err error, context map[string]interface{}) {
    if logErr, ok := err.(*pacific.LogErroApp); ok {
        pacific.LogErr001(logErr)
    } else {
        pacific.LogErr001(&pacific.LogErroApp{
            Codigo:    "GENERIC_ERROR",
            Mensagem:  err.Error(),
            Timestamp: time.Now().Format(time.RFC3339),
        })
    }
}
```

## Dependências

- `encoding/json` - Serialização JSON
- `fmt` - Formatação de strings
- `time` - Manipulação de tempo
- `context` - Controle de contexto
- `net/http` - Cliente HTTP

## Veja Também

- [Pacote Auth](../auth/README.md) - Para autenticação
- [Pacote Formatter](../formatter/README.md) - Para formatação de erros
- [Pacote Validator](../validator/README.md) - Para validação de dados
- [Pacote OpenTelemetry](../opentelemetry/README.md) - Para observabilidade

---

**Nota**: Este pacote é específico para integração com sistemas Pacific. Certifique-se de ter as credenciais e permissões adequadas antes de usar em produção.