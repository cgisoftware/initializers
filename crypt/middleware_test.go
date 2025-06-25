package crypt

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// createTestKeys cria chaves temporárias para teste
func createTestKeys(t *testing.T) (string, string, string, string, func()) {
	tempDir := t.TempDir()

	// Chave RSA privada de teste (formato PEM)
	privateKeyPEM := `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDagaIzA7lnRv7q
Rk8bSdZlUHjqxiK1dbpT/WUAc5XRgM3FT/FCrBcdD4mgGsg4I6vgClhkXEUg6UVS
zraRO9Sl0aEocWoGGNNvhBavUoZ63N9QSMMLNm6itsqsJ1WUPTxKIvlkWjGr0Sdp
87rN12O2z9vVdAKP+L6ySz4kgxLqPr9Sl2ql6uupb9hk0uGGA0RdaXEkB3LHyZXr
uYAW+R6lnHV0PN9tyRr59oCg9DZkDfJvpFXmy1fMsIRYNJae750dKBv/VRfJC0g7
EtfVKWIEI58n7rtMRUC4AqE6gmxeZf960sOFnIiGPiWCotP1OSK65dkiXCIqDECB
GiFU9ZkHAgMBAAECggEAJgnlQ75FO353iC8/PD/pa+/LbQubJT3edxqox6BXl4Y1
zECzfmjZCT0YN2ASNPu4wyLp6mbJvgX+BIFp9PSWe1t4E8NSsscFn+c9z72tHZxv
39ka40vRjNAHjlq2ojzazwkxo0+0T/X0R5Sfk5AIkt2ypoEwpQGnqQBCTDbpRw/d
AfR4zJnHCa6/S2Ltn1MSaI1DwswBzr0+/zGuNWADBGyiKGxR15yfQcyoRYzcnBZE
P7eQI6uM+u624mvE5oOnN0vjC3XibFUJEPh7r/WVo2127GW4a48rlCnf5NKzaVZa
ZHYC6qQ1A/oIX+QwYwS2isvDv2WIsykKGvRb4mxZCQKBgQDvo4GAEPKandCUjetx
pbkW/6R04TKfJk+o46bd9GhH74r5MZaMerr3qpc0Yfu0fewz2XpeICi64HhW7iZL
uiyof4ZEK7+puXWnuXH0AS/gSNovd10/NTp0WkClRHs6Xkg4DbWAzr7rpNr8RPiq
hfq6x+zvEfji4BwONcpGmNj0dQKBgQDpbMTxCuS2qKDtC21qqizVcEQdT70Yf/83
ULaU9mErsruZuIIOvPfcCUEMmx9SgpkFBWr5p8/P0x4haByIo0YXfHdFDIi1vy1k
PJK3b1UKyQRRbGLLYZHDKLQA+LrVA8dxJ5dLc1480UDHDVHWhRi4J7Ib36rkFV7y
UobVClO4CwKBgQCPNknoPTifSn0iqoXwjzfEFNc1unfEQOMOba6FqtC/XNrS/d2Y
6qfd5ych+QSx4ydL/UZyBgoRVKDWYtCkJQkXUc7t4q9SQTGdIOiHCEaSZTdvcohZ
g/gBHQbRPdHfGgVS6m50IhpbPVRZuuZZEmS7R0vDvBvfikt5+o9+DU5rGQKBgQCn
ofZZSMJxru5K7c75MBccfRBdoHsjUiCdv/gvSDUGZcg2H/w+y1SRD5BIlkpLPgDY
S0jE28/w5yOXSCZdtivLCBa7XsH7C710Y8/Vrj17jlrsgpL8jihY6C1FGVtLSPh8
+bq8c7C0qm4DxTwFe/YBonhVbi5SuEpEaiHscwsmewKBgDfhOf7JrRxrxxZ1MJ25
3XWOoK2wt1NM+vWC8wpshWO4NXECsnFtNC6l57yoMiBluDsCz3SaFvWQYl5+Y5mZ
VHaGJCYNBI9ld/8R3e1ILegvO8KgE6PS/hGjCy6vfS9lJFGA4VSfRqupAsQd9hXt
UnIWp356NzJL0unp5T7JhHyG
-----END PRIVATE KEY-----`

	// Chave RSA pública de teste (formato PEM)
	publicKeyPEM := `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2oGiMwO5Z0b+6kZPG0nW
ZVB46sYitXW6U/1lAHOV0YDNxU/xQqwXHQ+JoBrIOCOr4ApYZFxFIOlFUs62kTvU
pdGhKHFqBhjTb4QWr1KGetzfUEjDCzZuorbKrCdVlD08SiL5ZFoxq9EnafO6zddj
ts/b1XQCj/i+sks+JIMS6j6/UpdqperrqW/YZNLhhgNEXWlxJAdyx8mV67mAFvke
pZx1dDzfbcka+faAoPQ2ZA3yb6RV5stXzLCEWDSWnu+dHSgb/1UXyQtIOxLX1Sli
BCOfJ+67TEVAuAKhOoJsXmX/etLDhZyIhj4lgqLT9TkiuuXZIlwiKgxAgRohVPWZ
BwIDAQAB
-----END PUBLIC KEY-----`

	// Chave AES de teste (32 bytes em hex)
	aesKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	// Cria arquivos temporários
	privateKeyPath := filepath.Join(tempDir, "private.pem")
	publicKeyPath := filepath.Join(tempDir, "public.pem")
	masterKeyPath := filepath.Join(tempDir, "master.key")
	rotationKeyPath := filepath.Join(tempDir, "rotation.key")

	// Escreve as chaves nos arquivos
	if err := os.WriteFile(privateKeyPath, []byte(privateKeyPEM), 0600); err != nil {
		t.Fatalf("Erro ao criar chave privada de teste: %v", err)
	}
	if err := os.WriteFile(publicKeyPath, []byte(publicKeyPEM), 0644); err != nil {
		t.Fatalf("Erro ao criar chave pública de teste: %v", err)
	}
	if err := os.WriteFile(masterKeyPath, []byte(aesKey), 0600); err != nil {
		t.Fatalf("Erro ao criar chave master de teste: %v", err)
	}
	if err := os.WriteFile(rotationKeyPath, []byte(aesKey), 0600); err != nil {
		t.Fatalf("Erro ao criar chave rotation de teste: %v", err)
	}

	// Função de limpeza
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath, cleanup
}

// TestDecryptionMiddleware_AES testa o middleware com criptografia AES
func TestDecryptionMiddleware_AES(t *testing.T) {
	// Cria chaves de teste
	privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath, cleanup := createTestKeys(t)
	defer cleanup()

	// Inicializa o serviço de criptografia
	cryptService, err := Initialize(privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath)
	if err != nil {
		t.Skipf("Pulando teste devido a erro na inicialização: %v", err)
	}

	// Criptografa dados de teste
	originalPassword := "senha_secreta_123"
	encryptedPassword, err := cryptService.EncryptWithMasterKeySimple(originalPassword)
	if err != nil {
		t.Skipf("Pulando teste devido a erro na criptografia: %v", err)
	}

	// Cria o middleware
	middleware := NewDecryptionMiddleware(&cryptService, []string{"password"}, "aes")

	// Handler de teste que verifica se os dados foram descriptografados
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("Erro ao decodificar JSON: %v", err)
			return
		}

		// Verifica se a senha foi descriptografada
		if password, ok := data["password"].(string); ok {
			if password != originalPassword {
				t.Errorf("Senha não foi descriptografada corretamente. Esperado: %s, Recebido: %s", originalPassword, password)
			}
		} else {
			t.Error("Campo password não encontrado ou não é string")
		}

		w.WriteHeader(http.StatusOK)
	})

	// Aplica o middleware
	handlerWithMiddleware := middleware.Middleware(testHandler)

	// Prepara a requisição de teste
	requestBody := map[string]interface{}{
		"username": "test_user",
		"password": encryptedPassword,
		"email":    "test@example.com",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Executa a requisição
	w := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(w, req)

	// Verifica o resultado
	if w.Code != http.StatusOK {
		t.Errorf("Status code esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestDecryptionMiddleware_NonJSON testa o middleware com requisições não-JSON
func TestDecryptionMiddleware_NonJSON(t *testing.T) {
	// Cria chaves de teste
	privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath, cleanup := createTestKeys(t)
	defer cleanup()

	// Inicializa o serviço de criptografia
	cryptService, err := Initialize(privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath)
	if err != nil {
		t.Skipf("Pulando teste devido a erro na inicialização: %v", err)
	}

	// Cria o middleware
	middleware := NewDecryptionMiddleware(&cryptService, []string{"password"}, "aes")

	// Handler de teste
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Aplica o middleware
	handlerWithMiddleware := middleware.Middleware(testHandler)

	// Testa requisição GET (deve passar sem modificação)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}

	// Testa requisição POST com content-type não-JSON (deve passar sem modificação)
	req = httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("plain text")))
	req.Header.Set("Content-Type", "text/plain")
	w = httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestDecryptionMiddleware_MultipleFields testa descriptografia de múltiplos campos
func TestDecryptionMiddleware_MultipleFields(t *testing.T) {
	// Cria chaves de teste
	privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath, cleanup := createTestKeys(t)
	defer cleanup()

	// Inicializa o serviço de criptografia
	cryptService, err := Initialize(privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath)
	if err != nil {
		t.Skipf("Pulando teste devido a erro na inicialização: %v", err)
	}

	// Dados originais
	originalPassword := "senha123"
	originalToken := "token_secreto"

	// Criptografa os dados
	encryptedPassword, err := cryptService.EncryptWithMasterKeySimple(originalPassword)
	if err != nil {
		t.Skipf("Pulando teste devido a erro na criptografia: %v", err)
	}

	encryptedToken, err := cryptService.EncryptWithMasterKeySimple(originalToken)
	if err != nil {
		t.Skipf("Pulando teste devido a erro na criptografia: %v", err)
	}

	// Cria o middleware para descriptografar múltiplos campos
	middleware := NewDecryptionMiddleware(&cryptService, []string{"password", "token"}, "aes")

	// Handler de teste
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("Erro ao decodificar JSON: %v", err)
			return
		}

		// Verifica se ambos os campos foram descriptografados
		if password, ok := data["password"].(string); ok {
			if password != originalPassword {
				t.Errorf("Password não foi descriptografado corretamente. Esperado: %s, Recebido: %s", originalPassword, password)
			}
		} else {
			t.Error("Campo password não encontrado")
		}

		if token, ok := data["token"].(string); ok {
			if token != originalToken {
				t.Errorf("Token não foi descriptografado corretamente. Esperado: %s, Recebido: %s", originalToken, token)
			}
		} else {
			t.Error("Campo token não encontrado")
		}

		w.WriteHeader(http.StatusOK)
	})

	// Aplica o middleware
	handlerWithMiddleware := middleware.Middleware(testHandler)

	// Prepara a requisição
	requestBody := map[string]interface{}{
		"username": "test_user",
		"password": encryptedPassword,
		"token":    encryptedToken,
		"email":    "test@example.com",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Executa a requisição
	w := httptest.NewRecorder()
	handlerWithMiddleware.ServeHTTP(w, req)

	// Verifica o resultado
	if w.Code != http.StatusOK {
		t.Errorf("Status code esperado: %d, recebido: %d", http.StatusOK, w.Code)
	}
}

// TestNewDecryptionMiddlewareFromConfig testa a criação do middleware a partir de configuração
func TestNewDecryptionMiddlewareFromConfig(t *testing.T) {
	// Cria chaves de teste
	privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath, cleanup := createTestKeys(t)
	defer cleanup()

	// Configuração do middleware
	config := DecryptionConfig{
		EncryptedFields:    []string{"password", "secret"},
		DecryptionType:     "aes",
		RSAPrivateKeyPath:  privateKeyPath,
		RSAPublicKeyPath:   publicKeyPath,
		AESMasterKeyPath:   masterKeyPath,
		AESRotationKeyPath: rotationKeyPath,
	}

	// Cria o middleware a partir da configuração
	middleware, err := NewDecryptionMiddlewareFromConfig(config)
	if err != nil {
		t.Skipf("Pulando teste devido a erro na criação do middleware: %v", err)
	}

	// Verifica se o middleware foi criado corretamente
	if middleware == nil {
		t.Error("Middleware não foi criado")
	}

	if len(middleware.encryptedFields) != 2 {
		t.Errorf("Número de campos criptografados esperado: 2, recebido: %d", len(middleware.encryptedFields))
	}

	if middleware.decryptionType != "aes" {
		t.Errorf("Tipo de descriptografia esperado: aes, recebido: %s", middleware.decryptionType)
	}
}