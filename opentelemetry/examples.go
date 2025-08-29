package opentelemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// ExampleUsage demonstra como usar o novo sistema de logging
func ExampleUsage() {
	ctx := context.Background()
	
	// 1. Configuração do logger
	config := &LoggerConfig{
		Level:           INFO,
		Format:          "json",
		IncludeTrace:    true,
		SensitiveFields: []string{"password", "token", "secret"},
		MaxBodySize:     1024,
		ServiceName:     "my-service",
	}
	
	// Inicializa o logger global
	InitializeStructuredLogger(config)
	
	// 2. Logs simples
	Info(ctx, "Aplicação iniciada", nil)
	Warn(ctx, "Configuração não encontrada, usando padrão", nil)
	
	// 3. Log de erro simples
	err := fmt.Errorf("erro de exemplo")
	Error(ctx, "Erro ao processar", err, nil)
	
	// 4. Log de negócio
	businessLog := &BusinessLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     INFO,
			Message:   "Usuário criado",
		},
		Operation:  "user_creation",
		UserID:     "123",
		EntityType: "user",
		EntityID:   "456",
		Metadata: map[string]interface{}{
			"email": "user@example.com",
			"role":  "admin",
		},
	}
	Info(ctx, "Usuário criado com sucesso", businessLog)
	
	// 5. Log de banco de dados
	start := time.Now()
	// ... executa query ...
	duration := time.Since(start)
	
	LogDatabaseQuery(ctx, "SELECT * FROM users WHERE id = $1", []interface{}{123}, duration, 1, nil)
	
	// 6. Log HTTP usando função helper
	req, _ := http.NewRequest("GET", "/api/users", nil)
	LogHTTPRequest(ctx, req, 200, 50*time.Millisecond, `{"users": []}`)
}

// ExampleHTTPMiddleware demonstra como usar o middleware de logging
func ExampleHTTPMiddleware() {
	// Configuração
	config := DefaultLoggerConfig()
	config.ServiceName = "api-service"
	logger := NewStructuredLogger(config)
	
	// Criação do handler
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"users": []}`))
	})
	
	// Aplicação do middleware
	handler := HTTPLoggingMiddleware(logger)(mux)
	
	// Servidor
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}
	
	// server.ListenAndServe()
	_ = server // evita warning de variável não usada
}

// ExampleDatabaseLogging demonstra logging de operações de banco
func ExampleDatabaseLogging(ctx context.Context) {
	logger := GetStructuredLogger()
	
	// Simulação de operação de banco
	start := time.Now()
	query := "INSERT INTO users (name, email) VALUES ($1, $2)"
	args := []interface{}{"João Silva", "joao@example.com"}
	
	// ... executa query ...
	duration := time.Since(start)
	rowsAffected := int64(1)
	
	// Log da operação
	dbLog := &DatabaseLog{
		BaseLog: BaseLog{
			Timestamp: start,
			Level:     INFO,
			Message:   "User inserted",
		},
		Query:        query,
		Args:         args,
		Duration:     duration,
		RowsAffected: rowsAffected,
		Database:     "postgres",
		Operation:    "INSERT",
	}
	
	logger.Info(ctx, "Usuário inserido no banco", dbLog)
}

// ExampleErrorHandling demonstra diferentes tipos de log de erro
func ExampleErrorHandling(ctx context.Context) {
	logger := GetStructuredLogger()
	
	// 1. Erro de validação
	businessLog := &BusinessLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     WARN,
			Message:   "Validation failed",
		},
		Operation: "user_validation",
		Metadata: map[string]interface{}{
			"field": "email",
			"value": "",
		},
	}
	logger.Warn(ctx, "Falha na validação", businessLog)
	
	// 2. Erro de banco de dados
	dbErr := fmt.Errorf("connection timeout")
	dbLog := &DatabaseLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     ERROR,
			Message:   "Database connection failed",
		},
		Query:    "SELECT * FROM users",
		Database: "postgres",
	}
	logger.Error(ctx, "Erro de conexão com banco", dbErr, dbLog)
	
	// 3. Erro HTTP
	req, _ := http.NewRequest("POST", "/api/users", nil)
	httpLog := &HTTPLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     ERROR,
			Message:   "HTTP request failed",
		},
		Method:     req.Method,
		Path:       req.URL.Path,
		StatusCode: 500,
		Duration:   100 * time.Millisecond,
	}
	logger.Error(ctx, "Erro na requisição HTTP", fmt.Errorf("internal server error"), httpLog)
}

// ExampleCompatibility demonstra compatibilidade com API antiga
func ExampleCompatibility(ctx context.Context) {
	// Usando a API antiga (ainda funciona)
	req, _ := http.NewRequest("GET", "/api/test", nil)
	httpLog := NewHttpLog(req, []byte(`{"result": "ok"}`), 200)
	
	ErrorLog(ctx, "Teste de compatibilidade", nil, WithHttpLog(httpLog))
	
	// Usando a nova API
	logger := GetStructuredLogger()
	logger.Info(ctx, "Nova API funcionando", httpLog)
}