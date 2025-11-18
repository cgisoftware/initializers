package crypt

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// CryptService encapsula operações de criptografia
type CryptService struct {
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	masterKey   []byte
	rotationKey []byte
}

// NewCryptService cria uma nova instância do serviço de criptografia
// Parâmetros: rsaPrivateKeyPath, rsaPublicKeyPath, aesMasterKeyPath, aesRotationKeyPath
// Use string vazia ("") para usar os caminhos padrão
func Initialize(rsaPrivateKeyPath, rsaPublicKeyPath, aesMasterKeyPath, aesRotationKeyPath string) (CryptService, error) {

	// Valida que todos os caminhos das chaves foram fornecidos
	if rsaPrivateKeyPath == "" {
		return CryptService{}, fmt.Errorf("caminho da chave RSA privada é obrigatório")
	}
	if rsaPublicKeyPath == "" {
		return CryptService{}, fmt.Errorf("caminho da chave RSA pública é obrigatório")
	}
	if aesMasterKeyPath == "" {
		return CryptService{}, fmt.Errorf("caminho da chave AES master é obrigatório")
	}
	if aesRotationKeyPath == "" {
		return CryptService{}, fmt.Errorf("caminho da chave AES de rotação é obrigatório")
	}

	// Carrega chaves RSA dos arquivos
	privateKey, err := LoadRSAPrivateKeyFromPath(rsaPrivateKeyPath)
	if err != nil {
		return CryptService{}, fmt.Errorf("erro ao carregar chave RSA privada: %v", err)
	}

	publicKey, err := LoadRSAPublicKeyFromPath(rsaPublicKeyPath)
	if err != nil {
		return CryptService{}, fmt.Errorf("erro ao carregar chave RSA pública: %v", err)
	}

	// Carrega chaves AES
	masterKey, err := LoadAESKeyFromPath(aesMasterKeyPath)
	if err != nil {
		return CryptService{}, fmt.Errorf("erro ao carregar chave AES master: %v", err)
	}

	rotationKey, err := LoadAESKeyFromPath(aesRotationKeyPath)
	if err != nil {
		return CryptService{}, fmt.Errorf("erro ao carregar chave AES de rotação: %v", err)
	}

	return CryptService{
		privateKey:  privateKey,
		publicKey:   publicKey,
		masterKey:   masterKey,
		rotationKey: rotationKey,
	}, nil
}

// EncryptData criptografa dados usando criptografia híbrida
func (cs *CryptService) EncryptData(data string) (string, error) {
	encrypted, err := HybridEncrypt(cs.publicKey, []byte(data))
	if err != nil {
		return "", fmt.Errorf("erro ao criptografar dados: %v", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptData descriptografa dados usando criptografia híbrida
func (cs *CryptService) DecryptData(encryptedData string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return []byte{}, fmt.Errorf("erro ao decodificar dados: %v", err)
	}

	decrypted, err := HybridDecrypt(cs.privateKey, data)
	if err != nil {
		return []byte{}, fmt.Errorf("erro ao descriptografar dados: %v", err)
	}
	return decrypted, nil
}

// EncryptWithMasterKeySimple criptografa usando a chave mestra AES (mais simples)
func (cs *CryptService) EncryptWithMasterKeySimple(data string) (string, error) {
	encrypted, err := EncryptWithMasterKey(cs.masterKey, []byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// DecryptWithMasterKeySimple descriptografa usando a chave mestra AES
func (cs *CryptService) DecryptWithMasterKeySimple(encryptedData string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return []byte{}, fmt.Errorf("erro ao decodificar dados: %v", err)
	}

	decrypted, err := DecryptWithMasterKey(cs.masterKey, data)
	if err != nil {
		return []byte{}, err
	}
	return decrypted, nil
}

// CryptManager gerencia diferentes tipos de criptografia
type CryptManager struct {
	hybridService CryptService
}

// EncryptPassword criptografa uma senha usando AES (recomendado para senhas)
func (cm *CryptManager) EncryptPassword(password string) (string, error) {
	return cm.hybridService.EncryptWithMasterKeySimple(password)
}

// DecryptPassword descriptografa uma senha
func (cm *CryptManager) DecryptPassword(encryptedPassword string) ([]byte, error) {
	return cm.hybridService.DecryptWithMasterKeySimple(encryptedPassword)
}

// EncryptSensitiveData criptografa dados sensíveis usando criptografia híbrida
func (cm *CryptManager) EncryptSensitiveData(data string) (string, error) {
	return cm.hybridService.EncryptData(data)
}

// DecryptSensitiveData descriptografa dados sensíveis
func (cm *CryptManager) DecryptSensitiveData(encryptedData string) ([]byte, error) {
	return cm.hybridService.DecryptData(encryptedData)
}

// GenerateRSAKeys gera um novo par de chaves RSA
// keySize: tamanho da chave em bits (recomendado: 2048 ou 4096)
func (cs *CryptService) GenerateRSAKeys(keySize int) (*RSAKeyPair, error) {
	return GenerateRSAKeyPair(keySize)
}

// GenerateRSAKeysDefault gera um par de chaves RSA com tamanho padrão de 2048 bits
func (cs *CryptService) GenerateRSAKeysDefault() (*RSAKeyPair, error) {
	return GenerateRSAKeyPairDefault()
}

// HybridEncryptWithKeys criptografa dados usando criptografia híbrida com chaves fornecidas
func (cs *CryptService) HybridEncryptWithKeys(data string, publicKey *rsa.PublicKey) (string, error) {
	encrypted, err := HybridEncrypt(publicKey, []byte(data))
	if err != nil {
		return "", fmt.Errorf("erro ao criptografar dados com chaves fornecidas: %v", err)
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// HybridDecryptWithKeys descriptografa dados usando criptografia híbrida com chaves fornecidas
func (cs *CryptService) HybridDecryptWithKeys(encryptedData string, privateKey *rsa.PrivateKey) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return []byte{}, fmt.Errorf("erro ao decodificar dados: %v", err)
	}

	decrypted, err := HybridDecrypt(privateKey, data)
	if err != nil {
		return []byte{}, fmt.Errorf("erro ao descriptografar dados com chaves fornecidas: %v", err)
	}
	return decrypted, nil
}

func GenerateToken(ctx context.Context, key []byte, data []byte) (string, error) {
	// Gerar IV aleatório
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("erro ao gerar IV aleatório: %w", err)
	}

	// Criar cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("erro ao criar cipher: %w", err)
	}

	// CBC mode
	mode := cipher.NewCBCEncrypter(block, iv)

	// PKCS7 padding
	padding := aes.BlockSize - len(data)%aes.BlockSize
	padText := append(data, bytesRepeat(byte(padding), padding)...)

	cipherText := make([]byte, len(padText))
	mode.CryptBlocks(cipherText, padText)

	return base64.StdEncoding.EncodeToString(cipherText) + "-" + base64.StdEncoding.EncodeToString(iv), nil
}

func DecryptToken(ctx context.Context, key []byte, token string) ([]byte, error) {
	parts := strings.Split(token, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("token inválido, formato esperado 'cipher-iv'")
	}

	cipherText, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar cipherText: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar IV: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher: %w", err)
	}

	if len(cipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("tamanho do cipherText inválido")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	mode.CryptBlocks(plainText, cipherText)

	if len(plainText) == 0 {
		return nil, fmt.Errorf("plaintext vazio após descriptografia")
	}
	padding := int(plainText[len(plainText)-1])
	if padding <= 0 || padding > aes.BlockSize || padding > len(plainText) {
		return nil, fmt.Errorf("padding PKCS7 inválido")
	}
	for i := range padding {
		if plainText[len(plainText)-1-i] != byte(padding) {
			return nil, fmt.Errorf("padding PKCS7 inconsistente")
		}
	}

	return plainText[:len(plainText)-padding], nil
}

// Função para gerar padding
func bytesRepeat(b byte, count int) []byte {
	result := make([]byte, count)
	for i := range result {
		result[i] = b
	}
	return result
}
