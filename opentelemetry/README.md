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

### Logs Dinâmicos

Para casos onde você precisa de máxima flexibilidade com campos customizáveis:

```go
// Log dinâmico simples
opentelemetry.LogDynamicInfo(ctx, "Operação realizada", map[string]interface{}{
    "user_id":    "12345",
    "action":     "create_user",
    "ip_address": "192.168.1.1",
    "duration":   "150ms",
})

// Log dinâmico com dados complexos
opentelemetry.LogDynamicWarn(ctx, "Rate limit atingido", map[string]interface{}{
    "user_id":       "67890",
    "endpoint":      "/api/data",
    "requests_count": 1000,
    "limit":         500,
    "reset_time":    time.Now().Add(1 * time.Hour),
    "client_info": map[string]interface{}{
        "user_agent": "MyApp/1.0",
        "platform":   "iOS",
        "version":    "14.5",
    },
})

// Builder pattern para logs dinâmicos
dynamicLog := opentelemetry.NewDynamicLog(opentelemetry.ERROR, "Falha no processamento", nil)
dynamicLog.WithField("error_code", "PROC_001").
    WithField("retry_count", 3).
    WithField("max_retries", 5).
    WithFields(map[string]interface{}{
        "queue_name":     "processing_queue",
        "message_id":     "msg_abc123",
        "processing_time": 5.2,
        "memory_usage":   "256MB",
    })

logger := opentelemetry.GetStructuredLogger()
logger.Error(ctx, "Falha no processamento", fmt.Errorf("timeout"), dynamicLog)
```

#### Funções de Conveniência para Logs Dinâmicos

```go
// Por nível de log
opentelemetry.LogDynamicDebug(ctx, "Debug info", fields)
opentelemetry.LogDynamicInfo(ctx, "Info message", fields)
opentelemetry.LogDynamicWarn(ctx, "Warning message", fields)
opentelemetry.LogDynamicError(ctx, "Error message", fields)
opentelemetry.LogDynamicFatal(ctx, "Fatal error", fields)
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

### DynamicLog

Para logs completamente customizáveis com campos dinâmicos:

```go
// Criação direta
dynamicLog := &opentelemetry.DynamicLog{
    BaseLog: opentelemetry.BaseLog{
        Timestamp: time.Now(),
        Level:     opentelemetry.INFO,
        Message:   "Operação customizada",
    },
    Fields: map[string]interface{}{
        "custom_field1": "valor1",
        "custom_field2": 123,
        "nested_data": map[string]interface{}{
            "sub_field": "sub_valor",
        },
    },
}

opentelemetry.Info(ctx, "Log customizado", dynamicLog)

// Usando builder pattern
dynamicLog := opentelemetry.NewDynamicLog(opentelemetry.WARN, "Alerta customizado", nil)
dynamicLog.WithField("alert_type", "performance").
    WithField("threshold", 95.5).
    WithFields(map[string]interface{}{
        "cpu_usage":    88.2,
        "memory_usage": 76.1,
        "disk_usage":   92.3,
    })

logger := opentelemetry.GetStructuredLogger()
logger.Warn(ctx, "Sistema sob carga", dynamicLog)
```

#### Casos de Uso do DynamicLog

- **Métricas customizadas**: Performance, uso de recursos, estatísticas
- **Auditoria**: Logs de compliance com campos específicos do domínio
- **Integrações**: Logs de APIs externas com metadados variáveis
- **Eventos de negócio**: Transações, operações com dados específicos
- **Debugging**: Logs temporários com informações contextuais

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