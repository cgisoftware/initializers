package opentelemetry

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	t "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// LogLevel representa os níveis de log disponíveis
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String retorna a representação em string do nível de log
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogData interface genérica para diferentes tipos de log
type LogData interface {
	GetLogFields() map[string]interface{}
	GetLogType() string
}

// BaseLog contém campos comuns a todos os logs
type BaseLog struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	TraceID   string    `json:"trace_id"`
	SpanID    string    `json:"span_id"`
	Service   string    `json:"service"`
}

// HTTPLog representa logs de requisições HTTP
type HTTPLog struct {
	BaseLog
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	StatusCode  int64               `json:"status_code"`
	Duration    time.Duration       `json:"duration_ms"`
	RequestID   string              `json:"request_id"`
	UserAgent   string              `json:"user_agent"`
	RemoteAddr  string              `json:"remote_addr"`
	RequestBody string              `json:"request_body,omitempty"`
	Response    string              `json:"response,omitempty"`
	Headers     map[string][]string `json:"headers,omitempty"`
	QueryParams map[string][]string `json:"query_params,omitempty"`
}

// GetLogFields implementa LogData interface
func (h *HTTPLog) GetLogFields() map[string]interface{} {
	return map[string]interface{}{
		"type":          h.GetLogType(),
		"timestamp":     h.Timestamp,
		"level":         h.Level.String(),
		"message":       h.Message,
		"trace_id":      h.TraceID,
		"span_id":       h.SpanID,
		"service":       h.Service,
		"method":        h.Method,
		"path":          h.Path,
		"status_code":   h.StatusCode,
		"duration_ms":   h.Duration.Milliseconds(),
		"request_id":    h.RequestID,
		"user_agent":    h.UserAgent,
		"remote_addr":   h.RemoteAddr,
		"request_body":  h.RequestBody,
		"response":      h.Response,
		"headers":       h.Headers,
		"query_params":  h.QueryParams,
	}
}

// GetLogType implementa LogData interface
func (h *HTTPLog) GetLogType() string {
	return "http"
}

// DatabaseLog representa logs de operações de banco de dados
type DatabaseLog struct {
	BaseLog
	Query        string        `json:"query"`
	Args         []interface{} `json:"args,omitempty"`
	Duration     time.Duration `json:"duration_ms"`
	RowsAffected int64         `json:"rows_affected"`
	Database     string        `json:"database"`
	Operation    string        `json:"operation"`
}

// GetLogFields implementa LogData interface
func (d *DatabaseLog) GetLogFields() map[string]interface{} {
	return map[string]interface{}{
		"type":          d.GetLogType(),
		"timestamp":     d.Timestamp,
		"level":         d.Level.String(),
		"message":       d.Message,
		"trace_id":      d.TraceID,
		"span_id":       d.SpanID,
		"service":       d.Service,
		"query":         d.Query,
		"args":          d.Args,
		"duration_ms":   d.Duration.Milliseconds(),
		"rows_affected": d.RowsAffected,
		"database":      d.Database,
		"operation":     d.Operation,
	}
}

// GetLogType implementa LogData interface
func (d *DatabaseLog) GetLogType() string {
	return "database"
}

// BusinessLog representa logs de operações de negócio
type BusinessLog struct {
	BaseLog
	Operation  string                 `json:"operation"`
	UserID     string                 `json:"user_id,omitempty"`
	EntityType string                 `json:"entity_type,omitempty"`
	EntityID   string                 `json:"entity_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DynamicLog permite logs completamente customizáveis com campos dinâmicos
type DynamicLog struct {
	BaseLog
	Fields map[string]interface{} `json:"fields"`
}

// GetLogFields implementa LogData interface
func (b *BusinessLog) GetLogFields() map[string]interface{} {
	fields := map[string]interface{}{
		"type":        b.GetLogType(),
		"timestamp":   b.Timestamp,
		"level":       b.Level.String(),
		"message":     b.Message,
		"trace_id":    b.TraceID,
		"span_id":     b.SpanID,
		"service":     b.Service,
		"operation":   b.Operation,
		"user_id":     b.UserID,
		"entity_type": b.EntityType,
		"entity_id":   b.EntityID,
	}
	
	// Adiciona metadata se existir
	for k, v := range b.Metadata {
		fields[k] = v
	}
	
	return fields
}

// GetLogType implementa LogData interface
func (b *BusinessLog) GetLogType() string {
	return "business"
}

// GetLogFields implementa LogData interface para DynamicLog
func (d *DynamicLog) GetLogFields() map[string]interface{} {
	fields := map[string]interface{}{
		"type":       d.GetLogType(),
		"timestamp":  d.Timestamp,
		"level":      d.Level.String(),
		"message":    d.Message,
		"trace_id":   d.TraceID,
		"span_id":    d.SpanID,
		"service":    d.Service,
	}
	
	// Adiciona todos os campos dinâmicos
	for key, value := range d.Fields {
		fields[key] = value
	}
	
	return fields
}

// GetLogType implementa LogData interface para DynamicLog
func (d *DynamicLog) GetLogType() string {
	return "dynamic"
}

// LoggerConfig configuração do logger estruturado
type LoggerConfig struct {
	Level           LogLevel `json:"level"`
	Format          string   `json:"format"` // "json" ou "text"
	IncludeTrace    bool     `json:"include_trace"`
	SensitiveFields []string `json:"sensitive_fields"`
	MaxBodySize     int      `json:"max_body_size"`
	ServiceName     string   `json:"service_name"`
}

// DefaultLoggerConfig retorna uma configuração padrão
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:           INFO,
		Format:          "json",
		IncludeTrace:    true,
		SensitiveFields: []string{"password", "token", "secret", "authorization", "cookie"},
		MaxBodySize:     1024,
		ServiceName:     "unknown",
	}
}

// StructuredLogger logger estruturado com suporte a diferentes tipos de log
type StructuredLogger struct {
	config *LoggerConfig
}

// NewStructuredLogger cria uma nova instância do logger estruturado
func NewStructuredLogger(config *LoggerConfig) *StructuredLogger {
	if config == nil {
		config = DefaultLoggerConfig()
	}
	return &StructuredLogger{
		config: config,
	}
}

// shouldLog verifica se deve fazer log baseado no nível
func (sl *StructuredLogger) shouldLog(level LogLevel) bool {
	return level >= sl.config.Level
}

// maskSensitiveData mascara dados sensíveis
func (sl *StructuredLogger) maskSensitiveData(fields map[string]interface{}) map[string]interface{} {
	masked := make(map[string]interface{})
	for k, v := range fields {
		if sl.isSensitiveField(k) {
			masked[k] = "***MASKED***"
		} else {
			masked[k] = v
		}
	}
	return masked
}

// isSensitiveField verifica se um campo é sensível
func (sl *StructuredLogger) isSensitiveField(field string) bool {
	lowerField := strings.ToLower(field)
	for _, sensitive := range sl.config.SensitiveFields {
		if strings.Contains(lowerField, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}

// getTraceInfo extrai informações de trace do contexto
func (sl *StructuredLogger) getTraceInfo(ctx context.Context) (string, string, string) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return "", "", ""
	}
	
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()
	contextName := ""
	
	if readOnlySpan, ok := span.(t.ReadOnlySpan); ok {
		contextName = readOnlySpan.Name()
	}
	
	return traceID, spanID, contextName
}

// log função interna para fazer log
func (sl *StructuredLogger) log(ctx context.Context, level LogLevel, message string, err error, data LogData) {
	if !sl.shouldLog(level) {
		return
	}
	
	traceID, spanID, contextName := sl.getTraceInfo(ctx)
	
	// Campos base
	fields := map[string]interface{}{
		"timestamp":    time.Now(),
		"level":        level.String(),
		"message":      message,
		"service":      sl.config.ServiceName,
		"context_name": contextName,
	}
	
	// Adiciona trace info se habilitado
	if sl.config.IncludeTrace {
		fields["trace_id"] = traceID
		fields["span_id"] = spanID
	}
	
	// Adiciona erro se existir
	if err != nil {
		fields["error"] = err.Error()
	}
	
	// Adiciona dados específicos do log
	if data != nil {
		dataFields := data.GetLogFields()
		for k, v := range dataFields {
			fields[k] = v
		}
	}
	
	// Mascara dados sensíveis
	fields = sl.maskSensitiveData(fields)
	
	// Faz o log usando o logger do OpenTelemetry
	switch level {
	case DEBUG:
		openTelemetryConfig.logger.DebugContext(ctx, message, convertToSlogArgs(fields)...)
	case INFO:
		openTelemetryConfig.logger.InfoContext(ctx, message, convertToSlogArgs(fields)...)
	case WARN:
		openTelemetryConfig.logger.WarnContext(ctx, message, convertToSlogArgs(fields)...)
	case ERROR, FATAL:
		openTelemetryConfig.logger.ErrorContext(ctx, message, convertToSlogArgs(fields)...)
	}
}

// convertToSlogArgs converte map para argumentos do slog
func convertToSlogArgs(fields map[string]interface{}) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}

// Debug faz log de nível DEBUG
func (sl *StructuredLogger) Debug(ctx context.Context, message string, data LogData) {
	sl.log(ctx, DEBUG, message, nil, data)
}

// Info faz log de nível INFO
func (sl *StructuredLogger) Info(ctx context.Context, message string, data LogData) {
	sl.log(ctx, INFO, message, nil, data)
}

// Warn faz log de nível WARN
func (sl *StructuredLogger) Warn(ctx context.Context, message string, data LogData) {
	sl.log(ctx, WARN, message, nil, data)
}

// Error faz log de nível ERROR
func (sl *StructuredLogger) Error(ctx context.Context, message string, err error, data LogData) {
	sl.log(ctx, ERROR, message, err, data)
}

// Fatal faz log de nível FATAL
func (sl *StructuredLogger) Fatal(ctx context.Context, message string, err error, data LogData) {
	sl.log(ctx, FATAL, message, err, data)
}

// Instância global do logger estruturado
var globalStructuredLogger *StructuredLogger

// InitializeStructuredLogger inicializa o logger estruturado global
func InitializeStructuredLogger(config *LoggerConfig) {
	globalStructuredLogger = NewStructuredLogger(config)
}

// GetStructuredLogger retorna a instância global do logger estruturado
func GetStructuredLogger() *StructuredLogger {
	if globalStructuredLogger == nil {
		globalStructuredLogger = NewStructuredLogger(nil)
	}
	return globalStructuredLogger
}

// Funções de conveniência para compatibilidade com código existente
type postgresLog = DatabaseLog
type httpLog = HTTPLog

// NewHttpLog cria um novo log HTTP com a nova estrutura
func NewHttpLog(
	request *http.Request,
	response []byte,
	statusCode int64,
) *HTTPLog {
	bodyBytes, _ := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return &HTTPLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     INFO,
		},
		Method:      request.Method,
		Path:        request.URL.Path,
		StatusCode:  statusCode,
		Response:    string(response),
		Headers:     request.Header,
		QueryParams: request.URL.Query(),
		RequestBody: string(bodyBytes),
		RemoteAddr:  request.RemoteAddr,
		UserAgent:   request.UserAgent(),
	}
}

// NewPostgresLog cria um novo log de banco de dados
func NewPostgresLog(query string, queryArgs ...any) *DatabaseLog {
	return &DatabaseLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     INFO,
		},
		Query: query,
		Args:  queryArgs,
	}
}

// LegacyLoggerConfig para compatibilidade com código existente
type LegacyLoggerConfig struct {
	context     context.Context
	err         error
	httpLog     *HTTPLog
	postgresLog *DatabaseLog
}

type LegacyLoggerConfigOption func(d *LegacyLoggerConfig)

func WithHttpLog(value *HTTPLog) LegacyLoggerConfigOption {
	return func(c *LegacyLoggerConfig) {
		c.httpLog = value
	}
}

func WithPostgresLog(value *DatabaseLog) LegacyLoggerConfigOption {
	return func(c *LegacyLoggerConfig) {
		c.postgresLog = value
	}
}

// ErrorLog mantém compatibilidade com a API existente
func ErrorLog(
	ctx context.Context,
	message string,
	err error,
	opts ...LegacyLoggerConfigOption,
) {
	loggerConfig := &LegacyLoggerConfig{
		context: ctx,
		err:     err,
	}
	for _, opt := range opts {
		opt(loggerConfig)
	}

	// Usa o logger estruturado
	logger := GetStructuredLogger()
	
	if loggerConfig.httpLog != nil {
		logger.Error(ctx, message, err, loggerConfig.httpLog)
	} else if loggerConfig.postgresLog != nil {
		logger.Error(ctx, message, err, loggerConfig.postgresLog)
	} else {
		// Log simples sem dados específicos
		logger.Error(ctx, message, err, nil)
	}
}
