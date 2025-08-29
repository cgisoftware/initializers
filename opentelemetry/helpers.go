package opentelemetry

import (
	"context"
	"net/http"
	"time"
)

// HTTPMiddleware middleware para logging automático de requisições HTTP
func HTTPLoggingMiddleware(logger *StructuredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrapper para capturar o status code
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     200, // default
			}
			
			next.ServeHTTP(recorder, r)
			
			duration := time.Since(start)
			
			// Cria o log HTTP
			httpLog := &HTTPLog{
				BaseLog: BaseLog{
					Timestamp: start,
					Level:     INFO,
					Message:   "HTTP Request",
				},
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: int64(recorder.statusCode),
				Duration:   duration,
				UserAgent:  r.UserAgent(),
				RemoteAddr: r.RemoteAddr,
				Headers:    r.Header,
				QueryParams: r.URL.Query(),
			}
			
			// Log baseado no status code
			if recorder.statusCode >= 500 {
				httpLog.Level = ERROR
				httpLog.Message = "HTTP Server Error"
				logger.Error(r.Context(), "HTTP Server Error", nil, httpLog)
			} else if recorder.statusCode >= 400 {
				httpLog.Level = WARN
				httpLog.Message = "HTTP Client Error"
				logger.Warn(r.Context(), "HTTP Client Error", httpLog)
			} else {
				logger.Info(r.Context(), "HTTP Request", httpLog)
			}
		})
	}
}

// responseRecorder captura o status code da resposta
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// LogHTTPRequest função de conveniência para log manual de requisições HTTP
func LogHTTPRequest(ctx context.Context, r *http.Request, statusCode int, duration time.Duration, response string) {
	logger := GetStructuredLogger()
	
	httpLog := &HTTPLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     INFO,
			Message:   "HTTP Request",
		},
		Method:      r.Method,
		Path:        r.URL.Path,
		StatusCode:  int64(statusCode),
		Duration:    duration,
		UserAgent:   r.UserAgent(),
		RemoteAddr:  r.RemoteAddr,
		Response:    response,
		Headers:     r.Header,
		QueryParams: r.URL.Query(),
	}
	
	if statusCode >= 400 {
		httpLog.Level = ERROR
		httpLog.Message = "HTTP Error"
		logger.Error(ctx, "HTTP Error", nil, httpLog)
	} else {
		logger.Info(ctx, "HTTP Request", httpLog)
	}
}

// LogDatabaseQuery função de conveniência para log de queries de banco
func LogDatabaseQuery(ctx context.Context, query string, args []interface{}, duration time.Duration, rowsAffected int64, err error) {
	logger := GetStructuredLogger()
	
	dbLog := &DatabaseLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     INFO,
			Message:   "Database Query",
		},
		Query:        query,
		Args:         args,
		Duration:     duration,
		RowsAffected: rowsAffected,
		Database:     "postgres", // pode ser configurável
	}
	
	if err != nil {
		dbLog.Level = ERROR
		dbLog.Message = "Database Error"
		logger.Error(ctx, "Database Error", err, dbLog)
	} else {
		logger.Info(ctx, "Database Query", dbLog)
	}
}

// LogBusinessOperation função de conveniência para logs de negócio
func LogBusinessOperation(ctx context.Context, operation string, userID string, entityType string, entityID string, metadata map[string]interface{}, err error) {
	logger := GetStructuredLogger()
	
	businessLog := &BusinessLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     INFO,
			Message:   "Business Operation",
		},
		Operation:  operation,
		UserID:     userID,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   metadata,
	}
	
	if err != nil {
		businessLog.Level = ERROR
		businessLog.Message = "Business Operation Failed"
		logger.Error(ctx, "Business Operation Failed", err, businessLog)
	} else {
		logger.Info(ctx, "Business Operation", businessLog)
	}
}

// Funções de conveniência globais
func Debug(ctx context.Context, message string, data LogData) {
	GetStructuredLogger().Debug(ctx, message, data)
}

func Info(ctx context.Context, message string, data LogData) {
	GetStructuredLogger().Info(ctx, message, data)
}

func Warn(ctx context.Context, message string, data LogData) {
	GetStructuredLogger().Warn(ctx, message, data)
}

func Error(ctx context.Context, message string, err error, data LogData) {
	GetStructuredLogger().Error(ctx, message, err, data)
}

func Fatal(ctx context.Context, message string, err error, data LogData) {
	GetStructuredLogger().Fatal(ctx, message, err, data)
}

// NewDynamicLog cria um novo log dinâmico com campos customizáveis
func NewDynamicLog(level LogLevel, message string, fields map[string]interface{}) *DynamicLog {
	return &DynamicLog{
		BaseLog: BaseLog{
			Timestamp: time.Now(),
			Level:     level,
			Message:   message,
		},
		Fields: fields,
	}
}

// LogDynamic registra um log dinâmico com campos customizáveis
func LogDynamic(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) {
	dynamicLog := NewDynamicLog(level, message, fields)
	
	logger := GetStructuredLogger()
	switch level {
	case DEBUG:
		logger.Debug(ctx, message, dynamicLog)
	case INFO:
		logger.Info(ctx, message, dynamicLog)
	case WARN:
		logger.Warn(ctx, message, dynamicLog)
	case ERROR:
		logger.Error(ctx, message, nil, dynamicLog)
	case FATAL:
		logger.Fatal(ctx, message, nil, dynamicLog)
	}
}

// Funções de conveniência para logs dinâmicos por nível
func LogDynamicDebug(ctx context.Context, message string, fields map[string]interface{}) {
	LogDynamic(ctx, DEBUG, message, fields)
}

func LogDynamicInfo(ctx context.Context, message string, fields map[string]interface{}) {
	LogDynamic(ctx, INFO, message, fields)
}

func LogDynamicWarn(ctx context.Context, message string, fields map[string]interface{}) {
	LogDynamic(ctx, WARN, message, fields)
}

func LogDynamicError(ctx context.Context, message string, fields map[string]interface{}) {
	LogDynamic(ctx, ERROR, message, fields)
}

func LogDynamicFatal(ctx context.Context, message string, fields map[string]interface{}) {
	LogDynamic(ctx, FATAL, message, fields)
}

// WithField adiciona um campo ao log dinâmico (builder pattern)
func (d *DynamicLog) WithField(key string, value interface{}) *DynamicLog {
	if d.Fields == nil {
		d.Fields = make(map[string]interface{})
	}
	d.Fields[key] = value
	return d
}

// WithFields adiciona múltiplos campos ao log dinâmico
func (d *DynamicLog) WithFields(fields map[string]interface{}) *DynamicLog {
	if d.Fields == nil {
		d.Fields = make(map[string]interface{})
	}
	for key, value := range fields {
		d.Fields[key] = value
	}
	return d
}