package crypt

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ExampleHandler demonstra como usar o middleware de descriptografia
func ExampleHandler() {
	// Configuração do middleware
	config := DecryptionConfig{
		EncryptedFields: []string{"password", "sensitive_data", "credit_card"},
		DecryptionType:  "aes", // ou "hybrid"
		RSAPrivateKeyPath:  "/path/to/private.pem",
		RSAPublicKeyPath:   "/path/to/public.pem",
		AESMasterKeyPath:   "/path/to/master.key",
		AESRotationKeyPath: "/path/to/rotation.key",
	}

	// Cria o middleware
	middleware, err := NewDecryptionMiddlewareFromConfig(config)
	if err != nil {
		log.Fatal("Erro ao criar middleware:", err)
	}

	// Handler que recebe os dados descriptografados
	userHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user struct {
			Username     string `json:"username"`
			Password     string `json:"password"`      // Este campo será descriptografado automaticamente
			SensitiveData string `json:"sensitive_data"` // Este campo também será descriptografado
			Email        string `json:"email"`
		}

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
			return
		}

		// Neste ponto, password e sensitive_data já estão descriptografados
		fmt.Printf("Usuário recebido: %+v\n", user)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Dados recebidos e descriptografados com sucesso",
			"user":    user.Username,
		})
	})

	// Aplica o middleware
	handlerWithMiddleware := middleware.Middleware(userHandler)

	// Registra o handler
	http.Handle("/api/users", handlerWithMiddleware)

	fmt.Println("Servidor iniciado na porta 8080")
	fmt.Println("Teste com: curl -X POST http://localhost:8080/api/users -H 'Content-Type: application/json' -d '{\"username\":\"test\",\"password\":\"<dados_criptografados>\",\"email\":\"test@example.com\"}'")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ExampleWithDirectMiddleware demonstra uso direto do middleware
func ExampleWithDirectMiddleware() {
	// Inicializa o serviço de criptografia diretamente
	cryptService, err := Initialize(
		"/path/to/private.pem",
		"/path/to/public.pem",
		"/path/to/master.key",
		"/path/to/rotation.key",
	)
	if err != nil {
		log.Fatal("Erro ao inicializar serviço:", err)
	}

	// Cria o middleware diretamente
	middleware := NewDecryptionMiddleware(
		&cryptService,
		[]string{"password", "token"},
		"hybrid", // Usa criptografia híbrida
	)

	// Handler de exemplo
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Dados descriptografados recebidos!"))
	})

	// Aplica o middleware
	http.Handle("/api/secure", middleware.Middleware(handler))
}

// ExampleEncryptionForTesting demonstra como criptografar dados para teste
func ExampleEncryptionForTesting() {
	cryptService, err := Initialize(
		"/path/to/private.pem",
		"/path/to/public.pem",
		"/path/to/master.key",
		"/path/to/rotation.key",
	)
	if err != nil {
		log.Fatal("Erro ao inicializar serviço:", err)
	}

	// Criptografa uma senha para teste
	password := "minha_senha_secreta"
	encryptedPassword, err := cryptService.EncryptWithMasterKeySimple(password)
	if err != nil {
		log.Fatal("Erro ao criptografar:", err)
	}

	fmt.Printf("Senha original: %s\n", password)
	fmt.Printf("Senha criptografada: %s\n", encryptedPassword)

	// Exemplo de JSON com dados criptografados
	testJSON := fmt.Sprintf(`{
		"username": "usuario_teste",
		"password": "%s",
		"email": "teste@example.com"
	}`, encryptedPassword)

	fmt.Printf("\nJSON de teste para enviar ao endpoint:\n%s\n", testJSON)
}

// MiddlewareChain demonstra como usar o middleware em uma cadeia
func MiddlewareChain() {
	// Middleware de logging
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}

	// Middleware de autenticação (exemplo)
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "Token de autorização necessário", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	// Middleware de descriptografia
	config := DecryptionConfig{
		EncryptedFields:    []string{"password", "sensitive_info"},
		DecryptionType:     "aes",
		RSAPrivateKeyPath:  "/path/to/private.pem",
		RSAPublicKeyPath:   "/path/to/public.pem",
		AESMasterKeyPath:   "/path/to/master.key",
		AESRotationKeyPath: "/path/to/rotation.key",
	}

	decryptMiddleware, err := NewDecryptionMiddlewareFromConfig(config)
	if err != nil {
		log.Fatal("Erro ao criar middleware de descriptografia:", err)
	}

	// Handler final
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Requisição processada com sucesso!"))
	})

	// Cadeia de middlewares: logging -> auth -> decrypt -> handler
	handlerChain := loggingMiddleware(
		authMiddleware(
			decryptMiddleware.Middleware(finalHandler),
		),
	)

	http.Handle("/api/protected", handlerChain)
}