package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleClaims implementa CustomClaims para demonstração
type ExampleClaims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	EncryptedKey string `json:"encrypted_key"` // Campo que pode estar criptografado
}

// GetFields implementa a interface CustomClaims
func (c *ExampleClaims) GetFields() map[ContextValue]any {
	return map[ContextValue]any{
		"user_id":       c.UserID,
		"username":      c.Username,
		"email":         c.Email,
		"encrypted_key": c.EncryptedKey,
	}
}

// ExampleCryptService implementa a interface CryptService para demonstração
type ExampleCryptService struct {
	// Em um caso real, você usaria o CryptService do pacote crypt
}

func (e *ExampleCryptService) DecryptWithMasterKeySimple(encryptedData string) (string, error) {
	// Simulação de descriptografia AES
	if encryptedData == "encrypted_secret_key" {
		return "decrypted_secret_key", nil
	}
	return "", fmt.Errorf("falha na descriptografia AES")
}

func (e *ExampleCryptService) DecryptData(encryptedData string) (string, error) {
	// Simulação de descriptografia híbrida
	if encryptedData == "hybrid_encrypted_data" {
		return "hybrid_decrypted_data", nil
	}
	return "", fmt.Errorf("falha na descriptografia híbrida")
}

// ExampleUsage demonstra como usar o middleware com e sem descriptografia
func ExampleUsage() {
	// Inicializa o autenticador
	claims := &ExampleClaims{}
	auth := Initialize("minha_chave_secreta", claims)

	// Cria um serviço de criptografia (em um caso real, use o do pacote crypt)
	cryptService := &ExampleCryptService{}

	// Handler protegido que usa valores descriptografados
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		// Obtém valores do contexto
		userID := GetInt64FromContext(ctx, "user_id")
		username := GetStringFromContext(ctx, "username")
		email := GetStringFromContext(ctx, "email")
		decryptedKey := GetStringFromContext(ctx, "encrypted_key") // Este valor foi descriptografado automaticamente
		
		response := map[string]interface{}{
			"user_id":       userID,
			"username":      username,
			"email":         email,
			"decrypted_key": decryptedKey,
			"message":       "Dados acessados com sucesso!",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Exemplo 1: Middleware SEM descriptografia (comportamento original)
	middlewareWithoutCrypt := auth.AuthMiddleware("user_id", "username", "email", "encrypted_key")
	http.Handle("/api/without-decrypt", middlewareWithoutCrypt(protectedHandler))

	// Exemplo 2: Middleware COM descriptografia
	middlewareWithCrypt := auth.AuthMiddlewareWithCrypt(
		cryptService,
		"user_id", "username", "email", "encrypted_key", // encrypted_key será descriptografado automaticamente
	)
	http.Handle("/api/with-decrypt", middlewareWithCrypt(protectedHandler))

	// Exemplo 3: Middleware com descriptografia apenas para campos específicos
	middlewareSelectiveDecrypt := auth.AuthMiddlewareWithCrypt(
		cryptService,
		"encrypted_key", // Apenas este campo será descriptografado
	)
	http.Handle("/api/selective-decrypt", middlewareSelectiveDecrypt(protectedHandler))

	// Handler para gerar token de teste
	tokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testClaims := &ExampleClaims{
			UserID:       123,
			Username:     "test_user",
			Email:        "test@example.com",
			EncryptedKey: "encrypted_secret_key", // Este valor será descriptografado pelo middleware
		}
		
		token := GetSignToken(testClaims, time.Hour, "minha_chave_secreta")
		
		response := map[string]string{
			"token": token,
			"usage": "Use este token no header Authorization: Bearer <token>",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	http.Handle("/api/token", tokenHandler)

	log.Println("Servidor iniciado na porta 8080")
	log.Println("Endpoints disponíveis:")
	log.Println("  GET /api/token - Gera um token de teste")
	log.Println("  GET /api/without-decrypt - Endpoint sem descriptografia")
	log.Println("  GET /api/with-decrypt - Endpoint com descriptografia automática")
	log.Println("  GET /api/selective-decrypt - Endpoint com descriptografia seletiva")
	
	// Inicia o servidor
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ExampleIntegrationWithCryptPackage demonstra como integrar com o pacote crypt real
func ExampleIntegrationWithCryptPackage() {
	/*
	// Exemplo de integração com o pacote crypt real:
	
	import "github.com/cgisoftware/initializers/crypt"
	
	// Inicializa o serviço de criptografia
	cryptService, err := crypt.Initialize(
		"/path/to/private.pem",
		"/path/to/public.pem",
		"/path/to/master.key",
		"/path/to/rotation.key",
	)
	if err != nil {
		log.Fatal("Erro ao inicializar serviço de criptografia:", err)
	}
	
	// Inicializa o autenticador
	claims := &ExampleClaims{}
	auth := Initialize("minha_chave_secreta", claims)
	
	// Cria middleware com descriptografia
	middleware := auth.AuthMiddlewareWithCrypt(
		&cryptService, // Passa o serviço de criptografia real
		"user_id", "username", "encrypted_key",
	)
	
	// Aplica o middleware aos handlers
	http.Handle("/api/protected", middleware(protectedHandler))
	*/
	
	fmt.Println("Veja o código comentado acima para exemplo de integração com o pacote crypt real")
}