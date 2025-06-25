package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestClaims implementa CustomClaims para testes
type TestClaims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	EncryptedKey string `json:"encrypted_key"`
}

func (c *TestClaims) GetFields() map[ContextValue]any {
	return map[ContextValue]any{
		"user_id":       c.UserID,
		"username":      c.Username,
		"encrypted_key": c.EncryptedKey,
	}
}

// MockCryptService simula o serviço de criptografia para testes
type MockCryptService struct {
	aesDecryptMap    map[string]string
	hybridDecryptMap map[string]string
}

func NewMockCryptService() *MockCryptService {
	return &MockCryptService{
		aesDecryptMap: map[string]string{
			"aes_encrypted_value": "aes_decrypted_value",
			"encrypted_password":  "decrypted_password",
		},
		hybridDecryptMap: map[string]string{
			"hybrid_encrypted_value": "hybrid_decrypted_value",
			"complex_encrypted_data": "complex_decrypted_data",
		},
	}
}

func (m *MockCryptService) DecryptWithMasterKeySimple(encryptedData string) (string, error) {
	if decrypted, exists := m.aesDecryptMap[encryptedData]; exists {
		return decrypted, nil
	}
	return "", fmt.Errorf("falha na descriptografia AES para: %s", encryptedData)
}

func (m *MockCryptService) DecryptData(encryptedData string) (string, error) {
	if decrypted, exists := m.hybridDecryptMap[encryptedData]; exists {
		return decrypted, nil
	}
	return "", fmt.Errorf("falha na descriptografia híbrida para: %s", encryptedData)
}

// TestAuthMiddleware_WithoutCrypt testa o middleware sem descriptografia (comportamento original)
func TestAuthMiddleware_WithoutCrypt(t *testing.T) {
	claims := &TestClaims{
		UserID:       123,
		Username:     "test_user",
		EncryptedKey: "aes_encrypted_value",
	}
	
	auth := Initialize("test_secret", claims)
	token := GetSignToken(claims, time.Hour, "test_secret")
	
	// Handler de teste
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		userID := GetInt64FromContext(ctx, "user_id")
		username := GetStringFromContext(ctx, "username")
		encryptedKey := GetStringFromContext(ctx, "encrypted_key")
		
		// Verifica se os valores estão corretos (sem descriptografia)
		if userID != 123 {
			t.Errorf("UserID esperado: 123, recebido: %d", userID)
		}
		if username != "test_user" {
			t.Errorf("Username esperado: test_user, recebido: %s", username)
		}
		if encryptedKey != "aes_encrypted_value" { // Deve permanecer criptografado
			t.Errorf("EncryptedKey esperado: aes_encrypted_value, recebido: %s", encryptedKey)
		}
		
		w.WriteHeader(http.StatusOK)
	})
	
	// Aplica middleware sem descriptografia
	middleware := auth.AuthMiddleware("user_id", "username", "encrypted_key")
	handler := middleware(testHandler)
	
	// Cria requisição com token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Status esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestAuthMiddlewareWithCrypt_AESDecryption testa descriptografia AES
func TestAuthMiddlewareWithCrypt_AESDecryption(t *testing.T) {
	claims := &TestClaims{
		UserID:       123,
		Username:     "test_user",
		EncryptedKey: "aes_encrypted_value",
	}
	
	auth := Initialize("test_secret", claims)
	token := GetSignToken(claims, time.Hour, "test_secret")
	cryptService := NewMockCryptService()
	
	// Handler de teste
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		userID := GetInt64FromContext(ctx, "user_id")
		username := GetStringFromContext(ctx, "username")
		decryptedKey := GetStringFromContext(ctx, "encrypted_key")
		
		// Verifica se os valores estão corretos
		if userID != 123 {
			t.Errorf("UserID esperado: 123, recebido: %d", userID)
		}
		if username != "test_user" {
			t.Errorf("Username esperado: test_user, recebido: %s", username)
		}
		if decryptedKey != "aes_decrypted_value" { // Deve estar descriptografado
			t.Errorf("DecryptedKey esperado: aes_decrypted_value, recebido: %s", decryptedKey)
		}
		
		w.WriteHeader(http.StatusOK)
	})
	
	// Aplica middleware com descriptografia
	middleware := auth.AuthMiddlewareWithCrypt(cryptService, "user_id", "username", "encrypted_key")
	handler := middleware(testHandler)
	
	// Cria requisição com token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Status esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestAuthMiddlewareWithCrypt_HybridDecryption testa descriptografia híbrida (fallback)
func TestAuthMiddlewareWithCrypt_HybridDecryption(t *testing.T) {
	claims := &TestClaims{
		UserID:       123,
		Username:     "test_user",
		EncryptedKey: "hybrid_encrypted_value", // Este valor só pode ser descriptografado com híbrida
	}
	
	auth := Initialize("test_secret", claims)
	token := GetSignToken(claims, time.Hour, "test_secret")
	cryptService := NewMockCryptService()
	
	// Handler de teste
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		decryptedKey := GetStringFromContext(ctx, "encrypted_key")
		
		if decryptedKey != "hybrid_decrypted_value" { // Deve estar descriptografado via híbrida
			t.Errorf("DecryptedKey esperado: hybrid_decrypted_value, recebido: %s", decryptedKey)
		}
		
		w.WriteHeader(http.StatusOK)
	})
	
	// Aplica middleware com descriptografia
	middleware := auth.AuthMiddlewareWithCrypt(cryptService, "encrypted_key")
	handler := middleware(testHandler)
	
	// Cria requisição com token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Status esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestAuthMiddlewareWithCrypt_FailedDecryption testa quando a descriptografia falha
func TestAuthMiddlewareWithCrypt_FailedDecryption(t *testing.T) {
	claims := &TestClaims{
		UserID:       123,
		Username:     "test_user",
		EncryptedKey: "unknown_encrypted_value", // Este valor não pode ser descriptografado
	}
	
	auth := Initialize("test_secret", claims)
	token := GetSignToken(claims, time.Hour, "test_secret")
	cryptService := NewMockCryptService()
	
	// Handler de teste
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		encryptedKey := GetStringFromContext(ctx, "encrypted_key")
		
		// Deve manter o valor original quando a descriptografia falha
		if encryptedKey != "unknown_encrypted_value" {
			t.Errorf("EncryptedKey esperado: unknown_encrypted_value, recebido: %s", encryptedKey)
		}
		
		w.WriteHeader(http.StatusOK)
	})
	
	// Aplica middleware com descriptografia
	middleware := auth.AuthMiddlewareWithCrypt(cryptService, "encrypted_key")
	handler := middleware(testHandler)
	
	// Cria requisição com token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Status esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestAuthMiddlewareWithCrypt_NonStringValue testa com valores não-string
func TestAuthMiddlewareWithCrypt_NonStringValue(t *testing.T) {
	claims := &TestClaims{
		UserID:   123, // Valor int64, não deve tentar descriptografar
		Username: "test_user",
	}
	
	auth := Initialize("test_secret", claims)
	token := GetSignToken(claims, time.Hour, "test_secret")
	cryptService := NewMockCryptService()
	
	// Handler de teste
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := GetInt64FromContext(ctx, "user_id")
		
		// Valor int64 deve permanecer inalterado
		if userID != 123 {
			t.Errorf("UserID esperado: 123, recebido: %d", userID)
		}
		
		w.WriteHeader(http.StatusOK)
	})
	
	// Aplica middleware com descriptografia
	middleware := auth.AuthMiddlewareWithCrypt(cryptService, "user_id")
	handler := middleware(testHandler)
	
	// Cria requisição com token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Status esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestAuthMiddleware_Unauthorized testa requisição não autorizada
func TestAuthMiddleware_Unauthorized(t *testing.T) {
	claims := &TestClaims{}
	auth := Initialize("test_secret", claims)
	cryptService := NewMockCryptService()
	
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler não deveria ser chamado para requisição não autorizada")
	})
	
	// Aplica middleware
	middleware := auth.AuthMiddlewareWithCrypt(cryptService, "user_id")
	handler := middleware(testHandler)
	
	// Cria requisição sem token
	req := httptest.NewRequest("GET", "/test", nil)
	
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status esperado: %d, recebido: %d", http.StatusUnauthorized, w.Code)
	}
}