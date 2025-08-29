# CGI Initializers

Coleção de pacotes Go para inicialização e configuração de componentes essenciais em aplicações web e APIs. Este projeto fornece uma suite completa de ferramentas para acelerar o desenvolvimento de aplicações robustas e escaláveis.

## 📦 Pacotes Disponíveis

### 🔐 [Auth](./auth/README.md)
Sistema de autenticação e autorização com suporte a JWT, middleware para frameworks web e validação de tokens.

**Principais funcionalidades:**
- Geração e validação de tokens JWT
- Middleware para Gin e Echo
- Autenticação baseada em claims
- Refresh tokens
- Blacklist de tokens

### 🔒 [Crypt](./crypt/README.md)
Serviços de criptografia para proteção de dados sensíveis, incluindo hash de senhas e criptografia simétrica.

**Principais funcionalidades:**
- Hash seguro de senhas (bcrypt)
- Criptografia AES
- Geração de chaves seguras
- Middleware de criptografia
- Validação de integridade

### 📋 [Formatter](./formatter/README.md)
Padronização de respostas de erro e formatação de dados para APIs REST.

**Principais funcionalidades:**
- Estruturas padronizadas de erro
- Formatação JSON consistente
- Integração com frameworks web
- Logging estruturado
- Internacionalização de mensagens

### ☁️ [GCS](./gcs/README.md)
Inicialização simplificada do cliente Google Cloud Storage para operações com buckets e objetos.

**Principais funcionalidades:**
- Configuração automática do cliente GCS
- Autenticação via Service Account
- Gerenciamento de contexto
- Exemplos práticos de uso

### 📊 [OpenTelemetry](./opentelemetry/README.md)
Observabilidade completa com tracing, métricas e logging estruturado.

**Principais funcionalidades:**
- Configuração automática de tracing
- Métricas customizadas
- Logging estruturado
- Propagação de contexto
- Integração com Jaeger/OTLP

### 🌊 [Pacific](./pacific/README.md)
Estrutura de dados e utilitários para integração com APIs específicas do domínio.

**Principais funcionalidades:**
- Estruturas de dados padronizadas
- Validação de entrada
- Tratamento de erros
- Serialização JSON
- Cliente HTTP integrado

### 🗄️ [Postgres](./postgres/README.md)
Gerenciamento de conexões e operações com banco de dados PostgreSQL.

**Principais funcionalidades:**
- Pool de conexões otimizado
- Operações CRUD simplificadas
- Suporte a transações
- Sistema de migrações
- Monitoramento e métricas

### ✅ [Validator](./validator/README.md)
Validação robusta de dados com suporte a regras customizadas e internacionalização.

**Principais funcionalidades:**
- Validações pré-definidas
- Regras customizadas
- Mensagens internacionalizadas
- Integração com frameworks web
- Cache de validações

## 🚀 Início Rápido

### Instalação

Cada pacote pode ser instalado individualmente:

```bash
# Instalar pacote específico
go get github.com/seu-usuario/cgi/initializers/auth
go get github.com/seu-usuario/cgi/initializers/postgres
go get github.com/seu-usuario/cgi/initializers/formatter
# ... outros pacotes
```

### Exemplo de Uso Integrado

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    
    "github.com/gin-gonic/gin"
    "github.com/seu-usuario/cgi/initializers/auth"
    "github.com/seu-usuario/cgi/initializers/postgres"
    "github.com/seu-usuario/cgi/initializers/formatter"
    "github.com/seu-usuario/cgi/initializers/validator"
    "github.com/seu-usuario/cgi/initializers/opentelemetry"
)

func main() {
    ctx := context.Background()
    
    // Inicializar OpenTelemetry para observabilidade
    otelConfig := opentelemetry.Initialize(ctx, opentelemetry.Config{
        ServiceName:    "minha-api",
        ServiceVersion: "1.0.0",
        Environment:    "production",
    })
    defer otelConfig.Shutdown(ctx)
    
    // Inicializar banco de dados
    dbConfig := postgres.Config{
        Host:     os.Getenv("DB_HOST"),
        Port:     os.Getenv("DB_PORT"),
        User:     os.Getenv("DB_USER"),
        Password: os.Getenv("DB_PASSWORD"),
        Database: os.Getenv("DB_NAME"),
    }
    
    db, err := postgres.Initialize(dbConfig)
    if err != nil {
        log.Fatal("Erro ao conectar com banco:", err)
    }
    defer db.Close()
    
    // Inicializar autenticação
    authConfig := auth.Config{
        SecretKey:       os.Getenv("JWT_SECRET"),
        TokenExpiration: "24h",
        Issuer:          "minha-api",
    }
    
    authService, err := auth.Initialize(authConfig)
    if err != nil {
        log.Fatal("Erro ao inicializar auth:", err)
    }
    
    // Inicializar validador
    validatorConfig := validator.Initialize(validator.Config{
        Language: "pt-BR",
    })
    
    // Configurar Gin
    r := gin.Default()
    
    // Middleware de observabilidade
    r.Use(opentelemetry.GinMiddleware())
    
    // Middleware de autenticação
    r.Use(auth.GinMiddleware(authService))
    
    // Rotas da API
    api := r.Group("/api/v1")
    {
        api.POST("/login", loginHandler(authService))
        api.GET("/users", getUsersHandler(db))
        api.POST("/users", createUserHandler(db, validatorConfig))
    }
    
    // Iniciar servidor
    log.Println("Servidor iniciado na porta 8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func loginHandler(authService *auth.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var loginData struct {
            Email    string `json:"email" validate:"required,email"`
            Password string `json:"password" validate:"required,min=6"`
        }
        
        if err := c.ShouldBindJSON(&loginData); err != nil {
            c.JSON(400, formatter.ErrorResponse{
                Error: formatter.FormattedError{
                    Code:    "INVALID_INPUT",
                    Message: "Dados de entrada inválidos",
                    Details: err.Error(),
                },
            })
            return
        }
        
        // Lógica de autenticação...
        token, err := authService.GenerateToken("user123", map[string]interface{}{
            "email": loginData.Email,
        })
        
        if err != nil {
            c.JSON(500, formatter.ErrorResponse{
                Error: formatter.FormattedError{
                    Code:    "AUTH_ERROR",
                    Message: "Erro na autenticação",
                },
            })
            return
        }
        
        c.JSON(200, gin.H{
            "token": token,
            "user": gin.H{
                "email": loginData.Email,
            },
        })
    }
}

func getUsersHandler(db *postgres.Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Lógica para buscar usuários...
        users := []map[string]interface{}{
            {"id": 1, "name": "João", "email": "joao@email.com"},
            {"id": 2, "name": "Maria", "email": "maria@email.com"},
        }
        
        c.JSON(200, gin.H{
            "users": users,
            "total": len(users),
        })
    }
}

func createUserHandler(db *postgres.Database, validator *validator.ValidatorConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        var userData struct {
            Name     string `json:"name" validate:"required,min=2"`
            Email    string `json:"email" validate:"required,email"`
            Password string `json:"password" validate:"required,min=8"`
        }
        
        if err := c.ShouldBindJSON(&userData); err != nil {
            c.JSON(400, formatter.ErrorResponse{
                Error: formatter.FormattedError{
                    Code:    "INVALID_INPUT",
                    Message: "Dados de entrada inválidos",
                },
            })
            return
        }
        
        // Validar dados
        if err := validator.ValidateStruct(userData); err != nil {
            c.JSON(400, formatter.ErrorResponse{
                Error: formatter.FormattedError{
                    Code:    "VALIDATION_ERROR",
                    Message: "Erro de validação",
                    Details: err.Error(),
                },
            })
            return
        }
        
        // Lógica para criar usuário...
        c.JSON(201, gin.H{
            "message": "Usuário criado com sucesso",
            "user": gin.H{
                "name":  userData.Name,
                "email": userData.Email,
            },
        })
    }
}
```

## 🏗️ Arquitetura

### Princípios de Design

1. **Modularidade**: Cada pacote é independente e pode ser usado isoladamente
2. **Configurabilidade**: Todas as configurações são externalizáveis
3. **Observabilidade**: Logging e tracing integrados em todos os componentes
4. **Segurança**: Práticas de segurança implementadas por padrão
5. **Performance**: Otimizações para alta performance e baixa latência

### Padrões Utilizados

- **Dependency Injection**: Configuração através de structs de configuração
- **Middleware Pattern**: Integração transparente com frameworks web
- **Factory Pattern**: Funções `Initialize()` para criação de instâncias
- **Observer Pattern**: Hooks para logging e monitoramento
- **Strategy Pattern**: Diferentes implementações para diferentes cenários

## 🔧 Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` na raiz do seu projeto:

```bash
# Banco de Dados
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=senha123
DB_NAME=minha_aplicacao
DB_SSL_MODE=disable
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5

# Autenticação
JWT_SECRET=sua-chave-secreta-muito-segura
JWT_EXPIRATION=24h
JWT_ISSUER=minha-api

# Google Cloud Storage
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
GCS_BUCKET_NAME=meu-bucket

# OpenTelemetry
OTEL_SERVICE_NAME=minha-api
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=production
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317

# Criptografia
ENCRYPTION_KEY=chave-de-32-caracteres-exatamente

# Validação
VALIDATOR_LANGUAGE=pt-BR
VALIDATOR_CACHE_SIZE=1000
```

### Arquivo de Configuração YAML

```yaml
# config/app.yaml
app:
  name: "Minha API"
  version: "1.0.0"
  environment: "production"
  port: 8080

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "senha123"
  database: "minha_aplicacao"
  ssl_mode: "disable"
  max_connections: 25
  max_idle_connections: 5
  connection_timeout: "30s"

auth:
  secret_key: "sua-chave-secreta"
  token_expiration: "24h"
  refresh_expiration: "168h"
  issuer: "minha-api"
  algorithm: "HS256"

gcs:
  credentials_path: "/path/to/service-account.json"
  bucket_name: "meu-bucket"
  timeout: "30s"

opentelemetry:
  service_name: "minha-api"
  service_version: "1.0.0"
  environment: "production"
  exporter:
    type: "otlp"
    endpoint: "http://localhost:4317"
  sampling:
    ratio: 0.1

validator:
  language: "pt-BR"
  cache_size: 1000
  custom_messages:
    required: "Este campo é obrigatório"
    email: "Formato de email inválido"

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

## 🐳 Docker

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar dependências do sistema
RUN apk add --no-cache git ca-certificates tzdata

# Copiar arquivos de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

WORKDIR /root/

# Instalar certificados CA
RUN apk --no-cache add ca-certificates tzdata

# Copiar binário
COPY --from=builder /app/main .

# Copiar arquivos de configuração
COPY --from=builder /app/config ./config

# Expor porta
EXPOSE 8080

# Comando de execução
CMD ["./main"]
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=senha123
      - DB_NAME=minha_aplicacao
      - JWT_SECRET=sua-chave-secreta
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4317
    depends_on:
      - postgres
      - redis
      - jaeger
    volumes:
      - ./config:/root/config
      - ./credentials:/root/credentials

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=senha123
      - POSTGRES_DB=minha_aplicacao
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "4317:4317"
      - "4318:4318"
    environment:
      - COLLECTOR_OTLP_ENABLED=true

volumes:
  postgres_data:
  redis_data:
```

## 🧪 Testes

### Estrutura de Testes

```bash
# Executar todos os testes
go test ./...

# Executar testes com coverage
go test -cover ./...

# Executar testes de integração
go test -tags=integration ./...

# Executar benchmarks
go test -bench=. ./...
```

### Exemplo de Teste de Integração

```go
// tests/integration_test.go
//go:build integration

package tests

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/seu-usuario/cgi/initializers/postgres"
    "github.com/seu-usuario/cgi/initializers/auth"
)

func TestFullIntegration(t *testing.T) {
    ctx := context.Background()
    
    // Setup banco de dados de teste
    dbConfig := postgres.Config{
        Host:     "localhost",
        Port:     "5432",
        User:     "postgres",
        Password: "senha123",
        Database: "test_db",
    }
    
    db, err := postgres.Initialize(dbConfig)
    assert.NoError(t, err)
    defer db.Close()
    
    // Setup autenticação
    authConfig := auth.Config{
        SecretKey:       "test-secret-key",
        TokenExpiration: "1h",
        Issuer:          "test-api",
    }
    
    authService, err := auth.Initialize(authConfig)
    assert.NoError(t, err)
    
    // Teste de geração de token
    token, err := authService.GenerateToken("user123", map[string]interface{}{
        "email": "test@example.com",
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    
    // Teste de validação de token
    claims, err := authService.ValidateToken(token)
    assert.NoError(t, err)
    assert.Equal(t, "user123", claims.Subject)
}
```

## 📊 Monitoramento

### Métricas Disponíveis

- **Database**: Conexões ativas, tempo de resposta, queries por segundo
- **Auth**: Tokens gerados, validações, falhas de autenticação
- **HTTP**: Requests por segundo, latência, códigos de status
- **GCS**: Uploads, downloads, erros de operação
- **Validator**: Validações executadas, cache hits/misses

### Dashboards Grafana

Exemplos de queries Prometheus:

```promql
# Taxa de requests HTTP
rate(http_requests_total[5m])

# Latência P95 de requests
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Conexões ativas do banco
postgres_connections_active

# Taxa de erro de autenticação
rate(auth_validation_errors_total[5m])
```

## 🔒 Segurança

### Checklist de Segurança

- [ ] Credenciais em variáveis de ambiente
- [ ] Tokens JWT com expiração adequada
- [ ] Senhas hasheadas com bcrypt
- [ ] Conexões de banco com SSL
- [ ] Rate limiting implementado
- [ ] Logs sem informações sensíveis
- [ ] Validação de entrada rigorosa
- [ ] CORS configurado adequadamente

### Auditoria de Segurança

```bash
# Verificar dependências vulneráveis
go list -json -m all | nancy sleuth

# Análise estática de segurança
gosec ./...

# Verificar licenças
go-licenses check ./...
```

## 🚀 Deploy

### Kubernetes

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minha-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: minha-api
  template:
    metadata:
      labels:
        app: minha-api
    spec:
      containers:
      - name: api
        image: minha-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: host
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-secret
              key: jwt-secret
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 📚 Documentação

Cada pacote possui documentação detalhada:

- [Auth](./auth/README.md) - Sistema de autenticação
- [Crypt](./crypt/README.md) - Serviços de criptografia
- [Formatter](./formatter/README.md) - Formatação de respostas
- [GCS](./gcs/README.md) - Google Cloud Storage
- [OpenTelemetry](./opentelemetry/README.md) - Observabilidade
- [Pacific](./pacific/README.md) - Estruturas de dados
- [Postgres](./postgres/README.md) - Banco de dados
- [Validator](./validator/README.md) - Validação de dados

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

### Padrões de Código

- Use `gofmt` para formatação
- Execute `golint` para verificar estilo
- Mantenha cobertura de testes > 80%
- Documente funções públicas
- Siga os padrões de nomenclatura Go

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

- **Issues**: [GitHub Issues](https://github.com/seu-usuario/cgi/issues)
- **Discussões**: [GitHub Discussions](https://github.com/seu-usuario/cgi/discussions)
- **Email**: suporte@exemplo.com

---

**Desenvolvido com ❤️ pela equipe CGI**