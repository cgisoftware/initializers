# Sistema de Logging Estruturado com OpenTelemetry

Este pacote fornece um sistema de logging estruturado integrado com OpenTelemetry, oferecendo diferentes níveis de log, mascaramento de dados sensíveis e suporte a diferentes tipos de operações.

## Características

- ✅ **Níveis de Log**: DEBUG, INFO, WARN, ERROR, FATAL
- ✅ **Logs Estruturados**: HTTP, Database, Business
- ✅ **Integração OpenTelemetry**: Trace ID e Span ID automáticos
- ✅ **Mascaramento de Dados**: Proteção automática de dados sensíveis
- ✅ **Middleware HTTP**: Logging automático de requisições
- ✅ **Compatibilidade**: Mantém API existente funcionando

## Instalação e Configuração

### Configuração Básica

```go
package main

import (
    "context"
    "github.com/cgisoftware/initializers/opentelemetry"
)

func main() {
    // Configuração do logger
    config := &opentelemetry.LoggerConfig{
        Level:           opentelemetry.INFO,
        Format:          "json",
        IncludeTrace:    true,
        SensitiveFields: []string{"password", "token", "secret"},
        MaxBodySize:     1024,
        ServiceName:     "meu-servico",
    }
    
    // Inicializa o logger global
    opentelemetry.InitializeStructuredLogger(config)
}
```

### Configuração com OpenTelemetry

```go
// Primeiro inicialize o OpenTelemetry
shutdown, err := opentelemetry.Initialize(
    ctx,
    opentelemetry.WithServiceName("meu-servico"),
    opentelemetry.WithOtelCollectorUri("http://localhost:4318"),
)
if err != nil {
    panic(err)
}
defer shutdown(ctx)

// Depois configure o logger estruturado
opentelemetry.InitializeStructuredLogger(&opentelemetry.LoggerConfig{
    ServiceName: "meu-servico",
    Level:       opentelemetry.INFO,
})
```

## Uso Básico

### Logs Simples

```go
ctx := context.Background()

// Logs por nível
opentelemetry.Debug(ctx, "Informação de debug", nil)
opentelemetry.Info(ctx, "Aplicação iniciada", nil)
opentelemetry.Warn(ctx, "Configuração não encontrada", nil)
opentelemetry.Error(ctx, "Erro ao processar", err, nil)
opentelemetry.Fatal(ctx, "Erro crítico", err, nil)
```

### Log de Requisições HTTP

```go
// Usando middleware (recomendado)
func setupServer() {
    logger := opentelemetry.GetStructuredLogger()
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", handleUsers)
    
    // Aplica middleware de logging
    handler := opentelemetry.HTTPLoggingMiddleware(logger)(mux)
    
    http.ListenAndServe(":8080", handler)
}

// Ou log manual
func handleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // ... processa requisição ...
    
    duration := time.Since(start)
    opentelemetry.LogHTTPRequest(r.Context(), r, 200, duration, "response body")
}
```

### Log de Operações de Banco

```go
func executeQuery(ctx context.Context, query string, args ...interface{}) {
    start := time.Now()
    
    // ... executa query ...
    
    duration := time.Since(start)
    rowsAffected := int64(1)
    
    opentelemetry.LogDatabaseQuery(ctx, query, args, duration, rowsAffected, nil)
}
```

### Log de Operações de Negócio

```go
func createUser(ctx context.Context, userID string, email string) {
    metadata := map[string]interface{}{
        "email": email,
        "role":  "user",
    }
    
    opentelemetry.LogBusinessOperation(
        ctx,
        "user_creation",
        userID,
        "user",
        "123",
        metadata,
        nil, // erro se houver
    )
}
```

## Tipos de Log Estruturados

### HTTPLog

Para requisições HTTP:

```go
httpLog := &opentelemetry.HTTPLog{
    BaseLog: opentelemetry.BaseLog{
        Timestamp: time.Now(),
        Level:     opentelemetry.INFO,
        Message:   "HTTP Request",
    },
    Method:      "GET",
    Path:        "/api/users",
    StatusCode:  200,
    Duration:    50 * time.Millisecond,
    UserAgent:   "Mozilla/5.0...",
    RemoteAddr:  "192.168.1.1",
}

opentelemetry.Info(ctx, "Requisição processada", httpLog)
```

### DatabaseLog

Para operações de banco de dados:

```go
dbLog := &opentelemetry.DatabaseLog{
    BaseLog: opentelemetry.BaseLog{
        Timestamp: time.Now(),
        Level:     opentelemetry.INFO,
        Message:   "Database Query",
    },
    Query:        "SELECT * FROM users WHERE id = $1",
    Args:         []interface{}{123},
    Duration:     10 * time.Millisecond,
    RowsAffected: 1,
    Database:     "postgres",
    Operation:    "SELECT",
}

opentelemetry.Info(ctx, "Query executada", dbLog)
```

### BusinessLog

Para operações de negócio:

```go
businessLog := &opentelemetry.BusinessLog{
    BaseLog: opentelemetry.BaseLog{
        Timestamp: time.Now(),
        Level:     opentelemetry.INFO,
        Message:   "User Created",
    },
    Operation:  "user_creation",
    UserID:     "user123",
    EntityType: "user",
    EntityID:   "456",
    Metadata: map[string]interface{}{
        "email": "user@example.com",
        "role":  "admin",
    },
}

opentelemetry.Info(ctx, "Usuário criado", businessLog)
```

## Configuração Avançada

### Níveis de Log

```go
config := &opentelemetry.LoggerConfig{
    Level: opentelemetry.DEBUG, // Só logs DEBUG e acima
}
```

### Mascaramento de Dados Sensíveis

```go
config := &opentelemetry.LoggerConfig{
    SensitiveFields: []string{
        "password",
        "token",
        "secret",
        "authorization",
        "cookie",
        "credit_card",
    },
}
```

### Controle de Tamanho do Body

```go
config := &opentelemetry.LoggerConfig{
    MaxBodySize: 2048, // máximo 2KB de body nos logs
}
```

## Compatibilidade com API Antiga

O sistema mantém compatibilidade com a API existente:

```go
// API antiga ainda funciona
req, _ := http.NewRequest("GET", "/api/test", nil)
httpLog := opentelemetry.NewHttpLog(req, []byte(`{"result": "ok"}`), 200)

opentelemetry.ErrorLog(ctx, "Teste", nil, opentelemetry.WithHttpLog(httpLog))
```

## Exemplo Completo

```go
package main

import (
    "context"
    "net/http"
    "time"
    
    "github.com/cgisoftware/initializers/opentelemetry"
)

func main() {
    ctx := context.Background()
    
    // 1. Inicializa OpenTelemetry
    shutdown, err := opentelemetry.Initialize(
        ctx,
        opentelemetry.WithServiceName("api-service"),
        opentelemetry.WithOtelCollectorUri("http://localhost:4318"),
    )
    if err != nil {
        panic(err)
    }
    defer shutdown(ctx)
    
    // 2. Configura logger estruturado
    opentelemetry.InitializeStructuredLogger(&opentelemetry.LoggerConfig{
        Level:       opentelemetry.INFO,
        ServiceName: "api-service",
    })
    
    // 3. Configura servidor HTTP com middleware
    logger := opentelemetry.GetStructuredLogger()
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", handleUsers)
    
    handler := opentelemetry.HTTPLoggingMiddleware(logger)(mux)
    
    opentelemetry.Info(ctx, "Servidor iniciado na porta 8080", nil)
    http.ListenAndServe(":8080", handler)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Log de operação de negócio
    opentelemetry.LogBusinessOperation(
        ctx,
        "list_users",
        "user123",
        "user",
        "",
        nil,
        nil,
    )
    
    // Simula query de banco
    start := time.Now()
    // ... executa query ...
    duration := time.Since(start)
    
    opentelemetry.LogDatabaseQuery(
        ctx,
        "SELECT * FROM users",
        nil,
        duration,
        10,
        nil,
    )
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"users": []}`))
}
```

## Formato de Saída

Os logs são estruturados em JSON:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "message": "HTTP Request",
  "service": "api-service",
  "trace_id": "abc123...",
  "span_id": "def456...",
  "type": "http",
  "method": "GET",
  "path": "/api/users",
  "status_code": 200,
  "duration_ms": 45,
  "user_agent": "Mozilla/5.0...",
  "remote_addr": "192.168.1.1"
}
```

## Migração da API Antiga

### Antes
```go
opentelemetry.ErrorLog(ctx, "Erro", err, 
    opentelemetry.WithHttpLog(httpLog),
)
```

### Depois
```go
logger := opentelemetry.GetStructuredLogger()
logger.Error(ctx, "Erro", err, httpLog)
```

Ou usando funções globais:
```go
opentelemetry.Error(ctx, "Erro", err, httpLog)
```