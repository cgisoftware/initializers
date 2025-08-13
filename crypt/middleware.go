package crypt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DecryptionMiddleware é um middleware HTTP que descriptografa automaticamente dados criptografados
type DecryptionMiddleware struct {
	cryptService *CryptService
	// Campos que devem ser descriptografados automaticamente
	encryptedFields []string
	// Tipo de descriptografia: "hybrid" ou "aes"
	decryptionType string
}

// NewDecryptionMiddleware cria uma nova instância do middleware de descriptografia
func NewDecryptionMiddleware(cryptService *CryptService, encryptedFields []string, decryptionType string) *DecryptionMiddleware {
	return &DecryptionMiddleware{
		cryptService:    cryptService,
		encryptedFields: encryptedFields,
		decryptionType:  decryptionType,
	}
}

// Middleware retorna o handler do middleware
func (dm *DecryptionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Só processa requisições com body (POST, PUT, PATCH)
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
			next.ServeHTTP(w, r)
			return
		}

		// Só processa requisições JSON
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			next.ServeHTTP(w, r)
			return
		}

		// Lê o body da requisição
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Erro ao ler body da requisição", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// Parse do JSON
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			// Se não conseguir fazer parse como JSON, passa adiante sem modificar
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		// Descriptografa os campos especificados
		if err := dm.decryptFields(data); err != nil {
			http.Error(w, fmt.Sprintf("Erro ao descriptografar dados: %v", err), http.StatusBadRequest)
			return
		}

		// Reconstrói o body com os dados descriptografados
		newBody, err := json.Marshal(data)
		if err != nil {
			http.Error(w, "Erro ao serializar dados descriptografados", http.StatusInternalServerError)
			return
		}

		// Substitui o body da requisição
		r.Body = io.NopCloser(bytes.NewReader(newBody))
		r.ContentLength = int64(len(newBody))

		// Continua para o próximo handler
		next.ServeHTTP(w, r)
	})
}

// decryptFields descriptografa os campos especificados no mapa de dados
func (dm *DecryptionMiddleware) decryptFields(data map[string]interface{}) error {
	for _, field := range dm.encryptedFields {
		if encryptedValue, exists := data[field]; exists {
			if encryptedStr, ok := encryptedValue.(string); ok && encryptedStr != "" {
				decryptedValue, err := dm.decryptValue(encryptedStr)
				if err != nil {
					return fmt.Errorf("erro ao descriptografar campo '%s': %v", field, err)
				}
				data[field] = decryptedValue
			}
		}
	}
	return nil
}

// decryptValue descriptografa um valor usando o tipo de descriptografia configurado
func (dm *DecryptionMiddleware) decryptValue(encryptedValue string) ([]byte, error) {
	switch dm.decryptionType {
	case "hybrid":
		return dm.cryptService.DecryptData(encryptedValue)
	case "aes":
		return dm.cryptService.DecryptWithMasterKeySimple(encryptedValue)
	default:
		return []byte{}, fmt.Errorf("tipo de descriptografia não suportado: %s", dm.decryptionType)
	}
}

// DecryptionConfig configuração para o middleware de descriptografia
type DecryptionConfig struct {
	// Campos que devem ser descriptografados
	EncryptedFields []string
	// Tipo de descriptografia: "hybrid" ou "aes"
	DecryptionType string
	// Caminhos das chaves de criptografia
	RSAPrivateKeyPath  string
	RSAPublicKeyPath   string
	AESMasterKeyPath   string
	AESRotationKeyPath string
}

// NewDecryptionMiddlewareFromConfig cria um middleware a partir de uma configuração
func NewDecryptionMiddlewareFromConfig(config DecryptionConfig) (*DecryptionMiddleware, error) {
	// Inicializa o serviço de criptografia
	cryptService, err := Initialize(
		config.RSAPrivateKeyPath,
		config.RSAPublicKeyPath,
		config.AESMasterKeyPath,
		config.AESRotationKeyPath,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar serviço de criptografia: %v", err)
	}

	return NewDecryptionMiddleware(&cryptService, config.EncryptedFields, config.DecryptionType), nil
}

// MiddlewareFunc retorna uma função middleware compatível com frameworks como Gin, Echo, etc.
func (dm *DecryptionMiddleware) MiddlewareFunc() func(http.Handler) http.Handler {
	return dm.Middleware
}

// GinMiddleware retorna um middleware compatível com Gin
func (dm *DecryptionMiddleware) GinMiddleware() func(c interface{}) {
	return func(c interface{}) {
		// Esta função seria implementada especificamente para Gin
		// Por enquanto, deixamos como placeholder
		panic("GinMiddleware não implementado - use MiddlewareFunc() com adaptador")
	}
}

// EchoMiddleware retorna um middleware compatível com Echo
func (dm *DecryptionMiddleware) EchoMiddleware() func(next interface{}) interface{} {
	return func(next interface{}) interface{} {
		// Esta função seria implementada especificamente para Echo
		// Por enquanto, deixamos como placeholder
		panic("EchoMiddleware não implementado - use MiddlewareFunc() com adaptador")
	}
}
