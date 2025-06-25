package crypt

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
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
func (cs *CryptService) DecryptData(encryptedData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar dados: %v", err)
	}

	decrypted, err := HybridDecrypt(cs.privateKey, data)
	if err != nil {
		return "", fmt.Errorf("erro ao descriptografar dados: %v", err)
	}
	return string(decrypted), nil
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
func (cs *CryptService) DecryptWithMasterKeySimple(encryptedData string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar dados: %v", err)
	}

	decrypted, err := DecryptWithMasterKey(cs.masterKey, data)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
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
func (cm *CryptManager) DecryptPassword(encryptedPassword string) (string, error) {
	return cm.hybridService.DecryptWithMasterKeySimple(encryptedPassword)
}

// EncryptSensitiveData criptografa dados sensíveis usando criptografia híbrida
func (cm *CryptManager) EncryptSensitiveData(data string) (string, error) {
	return cm.hybridService.EncryptData(data)
}

// DecryptSensitiveData descriptografa dados sensíveis
func (cm *CryptManager) DecryptSensitiveData(encryptedData string) (string, error) {
	return cm.hybridService.DecryptData(encryptedData)
}
