package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ExampleClaims implementa a interface CustomClaims
type ExampleClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Email    string `json:"email"`
}

// GetFields implementa a interface CustomClaims
func (c ExampleClaims) GetFields() map[ContextValue]any {
	return map[ContextValue]any{
		"user_id":  c.UserID,
		"username": c.Username,
		"role":     c.Role,
		"email":    c.Email,
	}
}

// ExampleBasicAuthentication demonstra como usar o sistema de autenticação básico
func ExampleBasicAuthentication() {
	// Configuração
	secretKey := "minha-chave-secreta-super-segura"
	claims := ExampleClaims{
		UserID:   123,
		Username: "joao.silva",
		Role:     "admin",
		Email:    "joao@exemplo.com",
	}

	// Inicializar o autenticador
	auth := Initialize(secretKey, claims)

	// Gerar token
	token := GetSignToken(claims, 24*time.Hour, secretKey)
	fmt.Printf("Token gerado: %s\n", token)

	// Criar middleware
	middleware := auth.AuthMiddleware("user_id", "username", "role")

	// Exemplo de handler protegido
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extrair dados do contexto
		userID := GetInt64FromContext(r.Context(), "user_id")
		username := GetStringFromContext(r.Context(), "username")
		role := GetStringFromContext(r.Context(), "role")

		fmt.Fprintf(w, "Usuário autenticado: ID=%d, Username=%s, Role=%s", userID, username, role)
	})

	// Aplicar middleware
	http.Handle("/protected", middleware(protectedHandler))

	fmt.Println("Servidor iniciado em :8080")
	fmt.Println("Teste com: curl -H \"Authorization: Bearer " + token + "\" http://localhost:8080/protected")
}

// ExampleWithCryptService demonstra como usar autenticação com serviço de criptografia
func ExampleWithCryptService() {
	// Mock do CryptService para exemplo
	cryptService := &MockCryptService{}

	secretKey := "minha-chave-secreta"
	claims := ExampleClaims{
		UserID:   456,
		Username: "maria.santos",
		Role:     "user",
		Email:    "maria@exemplo.com",
	}

	auth := Initialize(secretKey, claims)

	// Middleware com criptografia
	middleware := auth.AuthMiddlewareWithCrypt(cryptService, "user_id", "email")

	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetInt64FromContext(r.Context(), "user_id")
		email := GetStringFromContext(r.Context(), "email")

		fmt.Fprintf(w, "Usuário autenticado com criptografia: ID=%d, Email=%s", userID, email)
	})

	http.Handle("/protected-crypt", middleware(protectedHandler))
}

// ExampleContextExtraction demonstra como extrair diferentes tipos de dados do contexto
func ExampleContextExtraction() {
	ctx := context.Background()

	// Simular contexto com dados
	ctx = context.WithValue(ctx, ContextValue("user_id"), int64(789))
	ctx = context.WithValue(ctx, ContextValue("username"), "carlos.oliveira")
	ctx = context.WithValue(ctx, ContextValue("is_admin"), true)
	ctx = context.WithValue(ctx, ContextValue("balance"), 1500.75)
	ctx = context.WithValue(ctx, ContextValue("metadata"), map[string]interface{}{
		"last_login": "2024-01-15T10:30:00Z",
		"ip_address": "192.168.1.100",
	})

	// Extrair dados de diferentes tipos
	userID := GetInt64FromContext(ctx, "user_id")
	username := GetStringFromContext(ctx, "username")
	isAdmin := GetBoolFromContext(ctx, "is_admin")
	balance := GetFloat64FromContext(ctx, "balance")
	metadata := GetInterfaceFromContext(ctx, "metadata")

	fmt.Printf("User ID: %d\n", userID)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Is Admin: %t\n", isAdmin)
	fmt.Printf("Balance: %.2f\n", balance)
	fmt.Printf("Metadata: %+v\n", metadata)
}

// EcommerceClaims implementa CustomClaims para um sistema de e-commerce
type EcommerceClaims struct {
	CustomerID   int64    `json:"customer_id"`
	StoreID      int64    `json:"store_id"`
	Permissions  []string `json:"permissions"`
	Subscription string   `json:"subscription"`
}

// GetFields implementa a interface CustomClaims
func (c EcommerceClaims) GetFields() map[ContextValue]any {
	return map[ContextValue]any{
		"customer_id":  c.CustomerID,
		"store_id":     c.StoreID,
		"permissions": c.Permissions,
		"subscription": c.Subscription,
	}
}

// ExampleCustomClaims demonstra como criar claims personalizadas
func ExampleCustomClaims() {
	claims := EcommerceClaims{
		CustomerID:   999,
		StoreID:      123,
		Permissions:  []string{"read", "write", "delete"},
		Subscription: "premium",
	}

	// Gerar token
	token := GetSignToken(claims, 2*time.Hour, "ecommerce-secret")

	fmt.Printf("Token E-commerce: %s\n", token)
}

// ExampleMiddlewareChain demonstra como usar múltiplos middlewares
func ExampleMiddlewareChain() {
	secretKey := "chain-secret"
	claims := ExampleClaims{
		UserID:   111,
		Username: "admin",
		Role:     "super_admin",
		Email:    "admin@exemplo.com",
	}

	auth := Initialize(secretKey, claims)

	// Middleware de autenticação
	authMiddleware := auth.AuthMiddleware("user_id", "role")

	// Middleware de logging personalizado
	loggingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetInt64FromContext(r.Context(), "user_id")
			role := GetStringFromContext(r.Context(), "role")
			log.Printf("Acesso: UserID=%d, Role=%s, Path=%s", userID, role, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}

	// Handler final
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := GetStringFromContext(r.Context(), "role")
		if role != "super_admin" {
			http.Error(w, "Acesso negado", http.StatusForbidden)
			return
		}
		fmt.Fprintf(w, "Acesso liberado para super admin")
	})

	// Cadeia de middlewares
	http.Handle("/admin", authMiddleware(loggingMiddleware(finalHandler)))
}

// MockCryptService implementa CryptService para exemplos
type MockCryptService struct{}

func (m *MockCryptService) DecryptWithMasterKeySimple(encryptedData string) ([]byte, error) {
	// Implementação mock - em produção, use criptografia real
	return []byte("decrypted-" + encryptedData), nil
}

func (m *MockCryptService) DecryptData(encryptedData string) ([]byte, error) {
	// Implementação mock - em produção, use criptografia real
	return []byte("hybrid-decrypted-" + encryptedData), nil
}

// ExampleRunServer demonstra como executar um servidor completo com autenticação
func ExampleRunServer() {
	// Esta função demonstra um exemplo completo
	// Descomente para executar
	/*
	secretKey := "exemplo-servidor-secret"
	claims := ExampleClaims{
		UserID:   1,
		Username: "demo",
		Role:     "user",
		Email:    "demo@exemplo.com",
	}

	auth := Initialize(secretKey, claims)
	token := GetSignToken(claims, 24*time.Hour, secretKey)

	fmt.Printf("Token para testes: %s\n", token)

	// Rota pública
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Servidor de exemplo rodando!\nToken: %s", token)
	})

	// Rota protegida
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetInt64FromContext(r.Context(), "user_id")
		username := GetStringFromContext(r.Context(), "username")
		fmt.Fprintf(w, "Olá %s (ID: %d)!", username, userID)
	})

	http.Handle("/protected", auth.AuthMiddleware("user_id", "username")(protectedHandler))

	fmt.Println("Servidor rodando em :8080")
	fmt.Printf("Teste: curl -H \"Authorization: Bearer %s\" http://localhost:8080/protected\n", token)
	log.Fatal(http.ListenAndServe(":8080", nil))
	*/

	fmt.Println("Exemplo de servidor configurado (descomente para executar)")
}