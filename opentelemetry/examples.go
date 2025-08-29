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

// ExampleDynamicLogging demonstra o uso de logs dinâmicos
func ExampleDynamicLogging(ctx context.Context) {
	// 1. Log dinâmico simples
	LogDynamicInfo(ctx, "Operação realizada com sucesso", map[string]interface{}{
		"user_id":    "12345",
		"action":     "create_user",
		"ip_address": "192.168.1.1",
		"duration":   "150ms",
	})
	
	// 2. Log dinâmico com dados complexos
	LogDynamicWarn(ctx, "Rate limit atingido", map[string]interface{}{
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
	
	// 3. Log dinâmico com builder pattern
	dynamicLog := NewDynamicLog(ERROR, "Falha no processamento", nil)
	dynamicLog.WithField("error_code", "PROC_001").
		WithField("retry_count", 3).
		WithField("max_retries", 5).
		WithFields(map[string]interface{}{
			"queue_name":     "processing_queue",
			"message_id":     "msg_abc123",
			"processing_time": 5.2,
			"memory_usage":   "256MB",
		})
	
	logger := GetStructuredLogger()
	logger.Error(ctx, "Falha no processamento", fmt.Errorf("timeout after 30s"), dynamicLog)
	
	// 4. Log dinâmico para métricas customizadas
	LogDynamicDebug(ctx, "Métricas de performance", map[string]interface{}{
		"function_name":    "calculateRevenue",
		"execution_time_ms": 45.2,
		"memory_allocated": "12MB",
		"cpu_usage":       "15%",
		"cache_hits":      85,
		"cache_misses":    15,
		"database_queries": 3,
		"external_api_calls": 2,
	})
	
	// 5. Log dinâmico para auditoria
	LogDynamicInfo(ctx, "Ação de auditoria", map[string]interface{}{
		"audit_type":    "data_access",
		"user_id":       "admin_001",
		"user_role":     "administrator",
		"resource_type": "customer_data",
		"resource_id":   "cust_456789",
		"action":        "view",
		"timestamp":     time.Now().Unix(),
		"session_id":    "sess_xyz789",
		"compliance": map[string]interface{}{
			"gdpr_consent":  true,
			"data_category": "personal",
			"retention_days": 365,
		},
	})
}

// ExampleAdvancedDynamicLogging demonstra casos de uso avançados
func ExampleAdvancedDynamicLogging(ctx context.Context) {
	// 1. Log de transação financeira
	LogDynamicInfo(ctx, "Transação processada", map[string]interface{}{
		"transaction_id":   "txn_abc123",
		"amount":          1250.50,
		"currency":        "BRL",
		"payment_method":  "credit_card",
		"merchant_id":     "merch_456",
		"customer_id":     "cust_789",
		"status":          "approved",
		"processing_time": "2.3s",
		"fees": map[string]interface{}{
			"gateway_fee": 3.75,
			"platform_fee": 12.50,
			"total_fee":   16.25,
		},
		"risk_score": 0.15,
		"location": map[string]interface{}{
			"country": "BR",
			"state":   "SP",
			"city":    "São Paulo",
		},
	})
	
	// 2. Log de evento de sistema
	LogDynamicWarn(ctx, "Sistema sob alta carga", map[string]interface{}{
		"event_type":     "system_alert",
		"severity":       "medium",
		"cpu_usage":      85.5,
		"memory_usage":   78.2,
		"disk_usage":     92.1,
		"active_connections": 1250,
		"queue_size":     450,
		"response_time_avg": "850ms",
		"error_rate":     2.3,
		"affected_services": []string{"api", "worker", "cache"},
		"auto_scaling": map[string]interface{}{
			"triggered": true,
			"target_instances": 8,
			"current_instances": 5,
		},
	})
	
	// 3. Log de integração externa
	LogDynamicError(ctx, "Falha na integração externa", map[string]interface{}{
		"integration_name": "payment_gateway",
		"endpoint":        "https://api.gateway.com/v1/charge",
		"http_method":     "POST",
		"status_code":     503,
		"response_time":   "30s",
		"retry_attempt":   3,
		"max_retries":     5,
		"error_details": map[string]interface{}{
			"error_code":    "SERVICE_UNAVAILABLE",
			"error_message": "Gateway temporarily unavailable",
			"correlation_id": "corr_xyz123",
		},
		"circuit_breaker": map[string]interface{}{
			"state":        "half_open",
			"failure_count": 15,
			"threshold":    10,
		},
	})
}