# Pacote Formatter

O pacote `formatter` fornece funcionalidades para formatação padronizada de erros e respostas em aplicações Go, especialmente útil para APIs REST e sistemas que precisam de tratamento consistente de erros.

## Funcionalidades

### 🚨 Tratamento de Erros
- Formatação padronizada de erros
- Erros pré-definidos comuns
- Encapsulamento de erros (error wrapping)
- Códigos de erro personalizados

### 📝 Formatação de Respostas
- Respostas JSON estruturadas
- Metadados de erro
- Suporte a múltiplos idiomas
- Integração com frameworks web

### 🔧 Utilitários
- Verificação de tipos de erro
- Extração de informações de erro
- Logging estruturado
- Debugging facilitado

## Estruturas Principais

### `FormattedError`
```go
type FormattedError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
    Field   string `json:"field,omitempty"`
}
```

Estrutura principal para representar erros formatados.

### `ErrorResponse`
```go
type ErrorResponse struct {
    Success bool             `json:"success"`
    Error   *FormattedError  `json:"error,omitempty"`
    Errors  []FormattedError `json:"errors,omitempty"`
    Meta    map[string]interface{} `json:"meta,omitempty"`
}
```

Estrutura para respostas de erro completas.

## Erros Pré-definidos

O pacote inclui erros comuns pré-definidos:

```go
var (
    ErrInternalServer = &FormattedError{
        Code:    "INTERNAL_SERVER_ERROR",
        Message: "Erro interno do servidor",
    }
    
    ErrBadRequest = &FormattedError{
        Code:    "BAD_REQUEST",
        Message: "Requisição inválida",
    }
    
    ErrUnauthorized = &FormattedError{
        Code:    "UNAUTHORIZED",
        Message: "Não autorizado",
    }
    
    ErrForbidden = &FormattedError{
        Code:    "FORBIDDEN",
        Message: "Acesso negado",
    }
    
    ErrNotFound = &FormattedError{
        Code:    "NOT_FOUND",
        Message: "Recurso não encontrado",
    }
    
    ErrValidation = &FormattedError{
        Code:    "VALIDATION_ERROR",
        Message: "Erro de validação",
    }
    
    ErrDatabase = &FormattedError{
        Code:    "DATABASE_ERROR",
        Message: "Erro de banco de dados",
    }
    
    ErrTimeout = &FormattedError{
        Code:    "TIMEOUT",
        Message: "Tempo limite excedido",
    }
)
```

## Uso Básico

### Criação de Erros Simples
```go
package main

import (
    "fmt"
    "seu-projeto/initializers/formatter"
)

func main() {
    // Usar erro pré-definido
    err := formatter.ErrNotFound
    fmt.Printf("Erro: %+v\n", err)
    
    // Criar erro customizado
    customErr := &formatter.FormattedError{
        Code:    "CUSTOM_ERROR",
        Message: "Algo deu errado",
        Details: "Detalhes específicos do erro",
    }
    
    fmt.Printf("Erro customizado: %+v\n", customErr)
}
```

### Formatação de Respostas
```go
func handleError(err error) *formatter.ErrorResponse {
    if err == nil {
        return nil
    }
    
    // Verificar se é um erro formatado
    if formattedErr, ok := err.(*formatter.FormattedError); ok {
        return &formatter.ErrorResponse{
            Success: false,
            Error:   formattedErr,
        }
    }
    
    // Erro genérico
    return &formatter.ErrorResponse{
        Success: false,
        Error: &formatter.FormattedError{
            Code:    "UNKNOWN_ERROR",
            Message: err.Error(),
        },
    }
}
```

## Integração com Frameworks Web

### Gin Framework
```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "seu-projeto/initializers/formatter"
)

// Middleware para tratamento de erros
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // Verificar se houve erros
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var statusCode int
            var response *formatter.ErrorResponse
            
            // Determinar código de status baseado no tipo de erro
            if formattedErr, ok := err.(*formatter.FormattedError); ok {
                statusCode = getStatusCodeFromError(formattedErr)
                response = &formatter.ErrorResponse{
                    Success: false,
                    Error:   formattedErr,
                }
            } else {
                statusCode = http.StatusInternalServerError
                response = &formatter.ErrorResponse{
                    Success: false,
                    Error:   formatter.ErrInternalServer,
                }
            }
            
            c.JSON(statusCode, response)
        }
    }
}

func getStatusCodeFromError(err *formatter.FormattedError) int {
    switch err.Code {
    case "BAD_REQUEST", "VALIDATION_ERROR":
        return http.StatusBadRequest
    case "UNAUTHORIZED":
        return http.StatusUnauthorized
    case "FORBIDDEN":
        return http.StatusForbidden
    case "NOT_FOUND":
        return http.StatusNotFound
    case "TIMEOUT":
        return http.StatusRequestTimeout
    default:
        return http.StatusInternalServerError
    }
}

// Handler de exemplo
func GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    if userID == "" {
        c.Error(&formatter.FormattedError{
            Code:    "MISSING_USER_ID",
            Message: "ID do usuário é obrigatório",
            Field:   "id",
        })
        return
    }
    
    user, err := getUserFromDatabase(userID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.Error(formatter.ErrNotFound)
        } else {
            c.Error(formatter.ErrDatabase)
        }
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    user,
    })
}
```

### Echo Framework
```go
package handlers

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "seu-projeto/initializers/formatter"
)

// Middleware para tratamento de erros
func ErrorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        err := next(c)
        if err != nil {
            var statusCode int
            var response *formatter.ErrorResponse
            
            if formattedErr, ok := err.(*formatter.FormattedError); ok {
                statusCode = getStatusCodeFromError(formattedErr)
                response = &formatter.ErrorResponse{
                    Success: false,
                    Error:   formattedErr,
                }
            } else if he, ok := err.(*echo.HTTPError); ok {
                statusCode = he.Code
                response = &formatter.ErrorResponse{
                    Success: false,
                    Error: &formatter.FormattedError{
                        Code:    "HTTP_ERROR",
                        Message: he.Message.(string),
                    },
                }
            } else {
                statusCode = http.StatusInternalServerError
                response = &formatter.ErrorResponse{
                    Success: false,
                    Error:   formatter.ErrInternalServer,
                }
            }
            
            return c.JSON(statusCode, response)
        }
        return nil
    }
}
```

## Encapsulamento de Erros

### Wrapping de Erros
```go
package services

import (
    "fmt"
    "seu-projeto/initializers/formatter"
)

func ProcessUser(userID string) error {
    // Validar entrada
    if userID == "" {
        return &formatter.FormattedError{
            Code:    "INVALID_INPUT",
            Message: "ID do usuário não pode estar vazio",
            Field:   "userID",
        }
    }
    
    // Buscar usuário
    user, err := getUserFromDB(userID)
    if err != nil {
        // Encapsular erro original
        return &formatter.FormattedError{
            Code:    "USER_FETCH_ERROR",
            Message: "Erro ao buscar usuário",
            Details: fmt.Sprintf("Erro original: %v", err),
        }
    }
    
    // Processar usuário
    err = processUserData(user)
    if err != nil {
        return &formatter.FormattedError{
            Code:    "USER_PROCESSING_ERROR",
            Message: "Erro ao processar dados do usuário",
            Details: err.Error(),
        }
    }
    
    return nil
}
```

### Chain de Erros
```go
type ErrorChain struct {
    errors []error
}

func NewErrorChain() *ErrorChain {
    return &ErrorChain{
        errors: make([]error, 0),
    }
}

func (ec *ErrorChain) Add(err error) {
    if err != nil {
        ec.errors = append(ec.errors, err)
    }
}

func (ec *ErrorChain) HasErrors() bool {
    return len(ec.errors) > 0
}

func (ec *ErrorChain) ToFormattedResponse() *formatter.ErrorResponse {
    if !ec.HasErrors() {
        return nil
    }
    
    var formattedErrors []formatter.FormattedError
    
    for _, err := range ec.errors {
        if formattedErr, ok := err.(*formatter.FormattedError); ok {
            formattedErrors = append(formattedErrors, *formattedErr)
        } else {
            formattedErrors = append(formattedErrors, formatter.FormattedError{
                Code:    "GENERIC_ERROR",
                Message: err.Error(),
            })
        }
    }
    
    return &formatter.ErrorResponse{
        Success: false,
        Errors:  formattedErrors,
    }
}

// Uso
func ValidateAndProcessUser(userData map[string]interface{}) *formatter.ErrorResponse {
    errorChain := NewErrorChain()
    
    // Múltiplas validações
    if name, ok := userData["name"].(string); !ok || name == "" {
        errorChain.Add(&formatter.FormattedError{
            Code:    "MISSING_NAME",
            Message: "Nome é obrigatório",
            Field:   "name",
        })
    }
    
    if email, ok := userData["email"].(string); !ok || !isValidEmail(email) {
        errorChain.Add(&formatter.FormattedError{
            Code:    "INVALID_EMAIL",
            Message: "Email inválido",
            Field:   "email",
        })
    }
    
    if age, ok := userData["age"].(float64); !ok || age < 18 {
        errorChain.Add(&formatter.FormattedError{
            Code:    "INVALID_AGE",
            Message: "Idade deve ser maior que 18 anos",
            Field:   "age",
        })
    }
    
    return errorChain.ToFormattedResponse()
}
```

## Logging Estruturado

### Integração com Loggers
```go
package logging

import (
    "encoding/json"
    "log"
    "seu-projeto/initializers/formatter"
)

type StructuredLogger struct {
    logger *log.Logger
}

func NewStructuredLogger() *StructuredLogger {
    return &StructuredLogger{
        logger: log.New(os.Stdout, "", log.LstdFlags),
    }
}

func (sl *StructuredLogger) LogError(err error, context map[string]interface{}) {
    logEntry := map[string]interface{}{
        "timestamp": time.Now().UTC(),
        "level":     "error",
        "context":   context,
    }
    
    if formattedErr, ok := err.(*formatter.FormattedError); ok {
        logEntry["error"] = map[string]interface{}{
            "code":    formattedErr.Code,
            "message": formattedErr.Message,
            "details": formattedErr.Details,
            "field":   formattedErr.Field,
        }
    } else {
        logEntry["error"] = map[string]interface{}{
            "message": err.Error(),
        }
    }
    
    jsonData, _ := json.Marshal(logEntry)
    sl.logger.Println(string(jsonData))
}

// Uso
func handleRequest(userID string) {
    logger := NewStructuredLogger()
    
    user, err := getUserFromDB(userID)
    if err != nil {
        logger.LogError(err, map[string]interface{}{
            "operation": "get_user",
            "user_id":   userID,
            "request_id": getRequestID(),
        })
        return
    }
    
    // ... continuar processamento
}
```

## Utilitários

### Verificação de Tipos de Erro
```go
package utils

import "seu-projeto/initializers/formatter"

// IsFormattedError verifica se o erro é do tipo FormattedError
func IsFormattedError(err error) bool {
    _, ok := err.(*formatter.FormattedError)
    return ok
}

// GetErrorCode extrai o código do erro se for FormattedError
func GetErrorCode(err error) string {
    if formattedErr, ok := err.(*formatter.FormattedError); ok {
        return formattedErr.Code
    }
    return "UNKNOWN"
}

// IsErrorCode verifica se o erro tem um código específico
func IsErrorCode(err error, code string) bool {
    return GetErrorCode(err) == code
}

// HasErrorField verifica se o erro tem um campo específico
func HasErrorField(err error, field string) bool {
    if formattedErr, ok := err.(*formatter.FormattedError); ok {
        return formattedErr.Field == field
    }
    return false
}

// Uso
func processError(err error) {
    if IsFormattedError(err) {
        code := GetErrorCode(err)
        
        switch code {
        case "VALIDATION_ERROR":
            // Tratar erro de validação
        case "DATABASE_ERROR":
            // Tratar erro de banco
        case "TIMEOUT":
            // Tratar timeout
        default:
            // Tratar erro genérico
        }
    }
}
```

### Conversores
```go
package converters

import (
    "encoding/json"
    "seu-projeto/initializers/formatter"
)

// ErrorToJSON converte erro para JSON
func ErrorToJSON(err error) ([]byte, error) {
    if formattedErr, ok := err.(*formatter.FormattedError); ok {
        return json.Marshal(formattedErr)
    }
    
    genericErr := &formatter.FormattedError{
        Code:    "GENERIC_ERROR",
        Message: err.Error(),
    }
    
    return json.Marshal(genericErr)
}

// JSONToError converte JSON para erro
func JSONToError(data []byte) error {
    var formattedErr formatter.FormattedError
    if err := json.Unmarshal(data, &formattedErr); err != nil {
        return err
    }
    
    return &formattedErr
}

// ErrorResponseToJSON converte resposta de erro para JSON
func ErrorResponseToJSON(response *formatter.ErrorResponse) ([]byte, error) {
    return json.Marshal(response)
}
```

## Internacionalização

### Suporte a Múltiplos Idiomas
```go
package i18n

import "seu-projeto/initializers/formatter"

type ErrorTranslator struct {
    translations map[string]map[string]string
}

func NewErrorTranslator() *ErrorTranslator {
    return &ErrorTranslator{
        translations: map[string]map[string]string{
            "en": {
                "VALIDATION_ERROR": "Validation error",
                "NOT_FOUND":        "Resource not found",
                "UNAUTHORIZED":     "Unauthorized access",
            },
            "pt": {
                "VALIDATION_ERROR": "Erro de validação",
                "NOT_FOUND":        "Recurso não encontrado",
                "UNAUTHORIZED":     "Acesso não autorizado",
            },
            "es": {
                "VALIDATION_ERROR": "Error de validación",
                "NOT_FOUND":        "Recurso no encontrado",
                "UNAUTHORIZED":     "Acceso no autorizado",
            },
        },
    }
}

func (et *ErrorTranslator) Translate(err *formatter.FormattedError, language string) *formatter.FormattedError {
    if translations, ok := et.translations[language]; ok {
        if message, exists := translations[err.Code]; exists {
            return &formatter.FormattedError{
                Code:    err.Code,
                Message: message,
                Details: err.Details,
                Field:   err.Field,
            }
        }
    }
    
    // Retornar erro original se tradução não encontrada
    return err
}

// Middleware para tradução automática
func TranslationMiddleware(translator *ErrorTranslator) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            language := c.GetHeader("Accept-Language")
            if language == "" {
                language = "en" // padrão
            }
            
            err := c.Errors.Last().Err
            if formattedErr, ok := err.(*formatter.FormattedError); ok {
                translatedErr := translator.Translate(formattedErr, language)
                
                response := &formatter.ErrorResponse{
                    Success: false,
                    Error:   translatedErr,
                }
                
                c.JSON(getStatusCodeFromError(translatedErr), response)
            }
        }
    }
}
```

## Testes

### Testes Unitários
```go
package formatter_test

import (
    "encoding/json"
    "testing"
    "seu-projeto/initializers/formatter"
)

func TestFormattedError(t *testing.T) {
    err := &formatter.FormattedError{
        Code:    "TEST_ERROR",
        Message: "Erro de teste",
        Details: "Detalhes do erro",
        Field:   "test_field",
    }
    
    if err.Code != "TEST_ERROR" {
        t.Errorf("Código esperado: TEST_ERROR, obtido: %s", err.Code)
    }
    
    if err.Message != "Erro de teste" {
        t.Errorf("Mensagem esperada: Erro de teste, obtida: %s", err.Message)
    }
}

func TestErrorResponse(t *testing.T) {
    response := &formatter.ErrorResponse{
        Success: false,
        Error: &formatter.FormattedError{
            Code:    "TEST_ERROR",
            Message: "Erro de teste",
        },
    }
    
    if response.Success {
        t.Error("Success deveria ser false")
    }
    
    if response.Error.Code != "TEST_ERROR" {
        t.Errorf("Código de erro esperado: TEST_ERROR, obtido: %s", response.Error.Code)
    }
}

func TestJSONSerialization(t *testing.T) {
    err := &formatter.FormattedError{
        Code:    "JSON_TEST",
        Message: "Teste de JSON",
    }
    
    // Serializar
    jsonData, jsonErr := json.Marshal(err)
    if jsonErr != nil {
        t.Fatal(jsonErr)
    }
    
    // Deserializar
    var deserializedErr formatter.FormattedError
    if jsonErr := json.Unmarshal(jsonData, &deserializedErr); jsonErr != nil {
        t.Fatal(jsonErr)
    }
    
    if deserializedErr.Code != err.Code {
        t.Errorf("Código não coincide após serialização/deserialização")
    }
}

func TestPredefinedErrors(t *testing.T) {
    tests := []struct {
        name string
        err  *formatter.FormattedError
        code string
    }{
        {"NotFound", formatter.ErrNotFound, "NOT_FOUND"},
        {"BadRequest", formatter.ErrBadRequest, "BAD_REQUEST"},
        {"Unauthorized", formatter.ErrUnauthorized, "UNAUTHORIZED"},
        {"InternalServer", formatter.ErrInternalServer, "INTERNAL_SERVER_ERROR"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.err.Code != tt.code {
                t.Errorf("Código esperado: %s, obtido: %s", tt.code, tt.err.Code)
            }
        })
    }
}
```

### Testes de Integração
```go
package integration_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "seu-projeto/initializers/formatter"
)

func TestErrorHandlerMiddleware(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    r := gin.New()
    r.Use(ErrorHandler())
    
    r.GET("/test-error", func(c *gin.Context) {
        c.Error(&formatter.FormattedError{
            Code:    "TEST_ERROR",
            Message: "Erro de teste",
        })
    })
    
    req, _ := http.NewRequest("GET", "/test-error", nil)
    w := httptest.NewRecorder()
    
    r.ServeHTTP(w, req)
    
    if w.Code != http.StatusInternalServerError {
        t.Errorf("Status esperado: %d, obtido: %d", http.StatusInternalServerError, w.Code)
    }
    
    var response formatter.ErrorResponse
    if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
        t.Fatal(err)
    }
    
    if response.Success {
        t.Error("Success deveria ser false")
    }
    
    if response.Error.Code != "TEST_ERROR" {
        t.Errorf("Código esperado: TEST_ERROR, obtido: %s", response.Error.Code)
    }
}
```

## Melhores Práticas

### 1. Consistência de Códigos
```go
// ✅ Use códigos consistentes e descritivos
const (
    ErrorCodeValidation    = "VALIDATION_ERROR"
    ErrorCodeNotFound      = "NOT_FOUND"
    ErrorCodeUnauthorized  = "UNAUTHORIZED"
    ErrorCodeDatabase      = "DATABASE_ERROR"
)

// ❌ Evite códigos genéricos ou inconsistentes
// "ERROR", "ERR1", "FAIL"
```

### 2. Mensagens Claras
```go
// ✅ Mensagens claras e acionáveis
&formatter.FormattedError{
    Code:    "INVALID_EMAIL",
    Message: "O email fornecido não é válido",
    Details: "Email deve ter formato: usuario@dominio.com",
    Field:   "email",
}

// ❌ Mensagens vagas
&formatter.FormattedError{
    Code:    "ERROR",
    Message: "Algo deu errado",
}
```

### 3. Não Exposição de Informações Sensíveis
```go
// ✅ Erro seguro
&formatter.FormattedError{
    Code:    "DATABASE_ERROR",
    Message: "Erro interno do servidor",
    // Details não inclui informações sensíveis
}

// ❌ Exposição de informações internas
&formatter.FormattedError{
    Code:    "DATABASE_ERROR",
    Message: "Erro interno do servidor",
    Details: "Connection failed: password authentication failed for user 'admin'",
}
```

### 4. Logging Adequado
```go
func handleDatabaseError(err error, operation string) *formatter.FormattedError {
    // Log completo para debugging interno
    log.Printf("Database error in %s: %v", operation, err)
    
    // Retorno seguro para cliente
    return &formatter.FormattedError{
        Code:    "DATABASE_ERROR",
        Message: "Erro interno do servidor",
    }
}
```

## Performance

### Otimizações

1. **Pool de Objetos**
   ```go
   var errorResponsePool = sync.Pool{
       New: func() interface{} {
           return &formatter.ErrorResponse{}
       },
   }
   
   func getErrorResponse() *formatter.ErrorResponse {
       return errorResponsePool.Get().(*formatter.ErrorResponse)
   }
   
   func putErrorResponse(resp *formatter.ErrorResponse) {
       resp.Success = false
       resp.Error = nil
       resp.Errors = nil
       resp.Meta = nil
       errorResponsePool.Put(resp)
   }
   ```

2. **Cache de Traduções**
   ```go
   var translationCache = make(map[string]*formatter.FormattedError)
   var cacheMutex sync.RWMutex
   
   func getCachedTranslation(code, language string) *formatter.FormattedError {
       key := fmt.Sprintf("%s:%s", code, language)
       
       cacheMutex.RLock()
       if cached, exists := translationCache[key]; exists {
           cacheMutex.RUnlock()
           return cached
       }
       cacheMutex.RUnlock()
       
       // Traduzir e cachear
       // ...
   }
   ```

## Dependências

- `encoding/json` - Serialização JSON
- `fmt` - Formatação de strings
- `errors` - Manipulação de erros padrão

## Veja Também

- [Pacote Validator](../validator/README.md) - Para validação com erros formatados
- [Pacote Auth](../auth/README.md) - Para autenticação com tratamento de erros
- [Pacote OpenTelemetry](../opentelemetry/README.md) - Para logging estruturado

---

**Nota**: Este pacote segue as melhores práticas de tratamento de erros em Go e fornece uma base sólida para aplicações que precisam de respostas de erro consistentes e bem estruturadas.