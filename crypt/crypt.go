package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash"
	"os"
)

const (
	// Tamanho da chave AES em bytes (256 bits)
	AESKeySize = 32
)

// Payload criptografado
type EncryptedPayload struct {
	EncryptedKey string `json:"encrypted_key"` // AES key criptografada com RSA
	Nonce        string `json:"nonce"`         // Nonce do AES-GCM
	Ciphertext   string `json:"ciphertext"`    // Dados criptografados com AES
}

// Gera uma chave AES de 256 bits
func generateAESKey() ([]byte, error) {
	key := make([]byte, AESKeySize)
	_, err := rand.Read(key)
	return key, err
}

// LoadAESKeyFromPath carrega uma chave AES de um caminho absoluto
func LoadAESKeyFromPath(filePath string) ([]byte, error) {
	// Verifica se o arquivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo de chave AES não encontrado: %s", filePath)
	}

	// Lê o conteúdo do arquivo
	hexKey, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de chave: %v", err)
	}

	// Decodifica de hexadecimal para bytes
	key, err := hex.DecodeString(string(hexKey))
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar chave hexadecimal: %v", err)
	}

	// Verifica se o tamanho da chave está correto
	if len(key) != AESKeySize {
		return nil, fmt.Errorf("tamanho de chave inválido: esperado %d bytes, obtido %d bytes", AESKeySize, len(key))
	}

	return key, nil
}

// Criptografa dados com AES-GCM
func encryptAES(aesKey, plaintext []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, nil, err
	}
	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Descriptografa com AES-GCM
func decryptAES(aesKey, nonce, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	return plaintext, err
}

// Criptografa com chave pública RSA
func encryptRSA(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(
		sha256Hash(),
		rand.Reader,
		pub,
		data,
		nil,
	)
}

// Descriptografa com chave privada RSA
func decryptRSA(priv *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(
		sha256Hash(),
		rand.Reader,
		priv,
		ciphertext,
		nil,
	)
}

// Utilitário para SHA-256 hash
func sha256Hash() hash.Hash {
	return sha256.New()
}

// Método para criptografar (Híbrido)
func HybridEncrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	aesKey, err := generateAESKey()
	if err != nil {
		return nil, err
	}

	nonce, ciphertext, err := encryptAES(aesKey, data)
	if err != nil {
		return nil, err
	}

	encKey, err := encryptRSA(pub, aesKey)
	if err != nil {
		return nil, err
	}

	payload := EncryptedPayload{
		EncryptedKey: base64.StdEncoding.EncodeToString(encKey),
		Nonce:        base64.StdEncoding.EncodeToString(nonce),
		Ciphertext:   base64.StdEncoding.EncodeToString(ciphertext),
	}
	return json.Marshal(payload)
}

// Método para descriptografar (Híbrido)
func HybridDecrypt(priv *rsa.PrivateKey, encrypted []byte) ([]byte, error) {
	var payload EncryptedPayload
	err := json.Unmarshal(encrypted, &payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar payload JSON: %v", err)
	}

	encKey, err := base64.StdEncoding.DecodeString(payload.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar chave criptografada: %v", err)
	}

	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar nonce: %v", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(payload.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar texto cifrado: %v", err)
	}

	aesKey, err := decryptRSA(priv, encKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar chave AES: %v", err)
	}

	plaintext, err := decryptAES(aesKey, nonce, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("erro ao descriptografar dados: %v", err)
	}

	return plaintext, nil
}

// Criptografia simétrica usando chave mestra
func EncryptWithMasterKey(masterKey []byte, plaintext []byte) ([]byte, error) {
	nonce, ciphertext, err := encryptAES(masterKey, plaintext)
	if err != nil {
		return nil, fmt.Errorf("erro na criptografia AES: %v", err)
	}

	// Combina nonce + ciphertext para facilitar o armazenamento
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)

	return result, nil
}

// Descriptografia simétrica usando chave mestra
func DecryptWithMasterKey(masterKey []byte, encrypted []byte) ([]byte, error) {
	// Extrai nonce e ciphertext
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("dados criptografados muito pequenos")
	}

	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]

	plaintext, err := decryptAES(masterKey, nonce, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("erro na descriptografia AES: %v", err)
	}

	return plaintext, nil
}

// Criptografia simétrica usando chave de rotação
func EncryptWithRotationKey(rotationKey []byte, plaintext []byte) ([]byte, error) {
	nonce, ciphertext, err := encryptAES(rotationKey, plaintext)
	if err != nil {
		return nil, fmt.Errorf("erro na criptografia AES: %v", err)
	}

	// Combina nonce + ciphertext
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)

	return result, nil
}

// Descriptografia simétrica usando chave de rotação
func DecryptWithRotationKey(rotationKey []byte, encrypted []byte) ([]byte, error) {
	// Extrai nonce e ciphertext
	block, err := aes.NewCipher(rotationKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar GCM: %v", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("dados criptografados muito pequenos")
	}

	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]

	plaintext, err := decryptAES(rotationKey, nonce, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("erro na descriptografia AES: %v", err)
	}

	return plaintext, nil
}

// LoadRSAPrivateKeyFromPath carrega uma chave RSA privada de um caminho absoluto
func LoadRSAPrivateKeyFromPath(filePath string) (*rsa.PrivateKey, error) {
	// Verifica se o arquivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo de chave RSA privada não encontrado: %s", filePath)
	}

	// Lê o arquivo
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de chave RSA privada: %v", err)
	}

	// Decodifica o bloco PEM
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("falha ao decodificar bloco PEM da chave RSA privada")
	}

	// Parse da chave privada
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Tenta PKCS8 se PKCS1 falhar
		privateKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("erro ao fazer parse da chave RSA privada: %v", err)
		}

		// Converte para *rsa.PrivateKey
		var ok bool
		privateKey, ok = privateKeyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("chave não é uma chave RSA privada válida")
		}
	}

	return privateKey, nil
}

// LoadRSAPublicKeyFromPath carrega uma chave RSA pública de um caminho absoluto
func LoadRSAPublicKeyFromPath(filePath string) (*rsa.PublicKey, error) {
	// Verifica se o arquivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo de chave RSA pública não encontrado: %s", filePath)
	}

	// Lê o arquivo
	keyData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de chave RSA pública: %v", err)
	}

	// Decodifica o bloco PEM
	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("falha ao decodificar bloco PEM da chave RSA pública")
	}

	// Parse da chave pública
	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse da chave RSA pública: %v", err)
	}

	// Converte para *rsa.PublicKey
	publicKey, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("chave não é uma chave RSA pública válida")
	}

	return publicKey, nil
}
