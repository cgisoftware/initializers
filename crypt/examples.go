package crypt

import (
	"fmt"
	"log"
)

// ExampleBasicAESEncryption demonstra criptografia AES básica
func ExampleBasicAESEncryption() {
	// Dados para criptografar
	plaintext := "Dados sensíveis que precisam ser protegidos"
	masterKeyPath := "/path/to/master.key"

	// Carregar chave AES
	masterKey, err := LoadAESKeyFromPath(masterKeyPath)
	if err != nil {
		log.Printf("Erro ao carregar chave: %v", err)
		return
	}

	// Criptografar dados
	encrypted, err := EncryptWithMasterKey(masterKey, []byte(plaintext))
	if err != nil {
		log.Printf("Erro ao criptografar: %v", err)
		return
	}

	fmt.Printf("Dados criptografados: %x\n", encrypted)

	// Descriptografar dados
	decrypted, err := DecryptWithMasterKey(masterKey, encrypted)
	if err != nil {
		log.Printf("Erro ao descriptografar: %v", err)
		return
	}

	fmt.Printf("Dados descriptografados: %s\n", string(decrypted))
}

// ExampleHybridEncryption demonstra criptografia híbrida RSA + AES
func ExampleHybridEncryption() {
	// Caminhos das chaves
	publicKeyPath := "/path/to/public.pem"
	privateKeyPath := "/path/to/private.pem"

	// Carregar chaves RSA
	publicKey, err := LoadRSAPublicKeyFromPath(publicKeyPath)
	if err != nil {
		log.Printf("Erro ao carregar chave pública: %v", err)
		return
	}

	privateKey, err := LoadRSAPrivateKeyFromPath(privateKeyPath)
	if err != nil {
		log.Printf("Erro ao carregar chave privada: %v", err)
		return
	}

	// Dados para criptografar
	data := "Informações confidenciais para criptografia híbrida"

	// Criptografar com RSA + AES
	encrypted, err := HybridEncrypt(publicKey, []byte(data))
	if err != nil {
		log.Printf("Erro na criptografia híbrida: %v", err)
		return
	}

	fmt.Printf("Dados criptografados (híbrido): %x\n", encrypted)

	// Descriptografar
	decrypted, err := HybridDecrypt(privateKey, encrypted)
	if err != nil {
		log.Printf("Erro na descriptografia híbrida: %v", err)
		return
	}

	fmt.Printf("Dados descriptografados: %s\n", string(decrypted))
}

// ExampleCryptService demonstra o uso do serviço de criptografia completo
func ExampleCryptService() {
	// Caminhos das chaves
	rsaPrivateKeyPath := "/path/to/rsa_private.pem"
	rsaPublicKeyPath := "/path/to/rsa_public.pem"
	aesMasterKeyPath := "/path/to/aes_master.key"
	aesRotationKeyPath := "/path/to/aes_rotation.key"

	// Inicializar serviço de criptografia
	cryptService, err := Initialize(rsaPrivateKeyPath, rsaPublicKeyPath, aesMasterKeyPath, aesRotationKeyPath)
	if err != nil {
		log.Printf("Erro ao inicializar serviço de criptografia: %v", err)
		return
	}

	// Exemplo 1: Criptografia híbrida
	data := "Dados para criptografia híbrida"
	encryptedData, err := cryptService.EncryptData(data)
	if err != nil {
		log.Printf("Erro ao criptografar dados: %v", err)
		return
	}

	fmt.Printf("Dados criptografados: %s\n", encryptedData)

	decryptedData, err := cryptService.DecryptData(encryptedData)
	if err != nil {
		log.Printf("Erro ao descriptografar dados: %v", err)
		return
	}

	fmt.Printf("Dados descriptografados: %s\n", string(decryptedData))

	// Exemplo 2: Criptografia com chave master
	sensitiveData := "Informação altamente sensível"
	encryptedSensitive, err := cryptService.EncryptWithMasterKeySimple(sensitiveData)
	if err != nil {
		log.Printf("Erro ao criptografar com chave master: %v", err)
		return
	}

	fmt.Printf("Dados sensíveis criptografados: %s\n", encryptedSensitive)

	decryptedSensitive, err := cryptService.DecryptWithMasterKeySimple(encryptedSensitive)
	if err != nil {
		log.Printf("Erro ao descriptografar com chave master: %v", err)
		return
	}

	fmt.Printf("Dados sensíveis descriptografados: %s\n", string(decryptedSensitive))
}

// ExampleCryptManager demonstra o uso do gerenciador de criptografia
func ExampleCryptManager() {
	// Inicializar serviço de criptografia
	cryptService, err := Initialize(
		"/path/to/rsa_private.pem",
		"/path/to/rsa_public.pem",
		"/path/to/aes_master.key",
		"/path/to/aes_rotation.key",
	)
	if err != nil {
		log.Printf("Erro ao inicializar serviço: %v", err)
		return
	}

	// Criar gerenciador
	manager := &CryptManager{hybridService: cryptService}

	// Criptografar senha
	password := "minhaSenhaSegura123!"
	encryptedPassword, err := manager.EncryptPassword(password)
	if err != nil {
		log.Printf("Erro ao criptografar senha: %v", err)
		return
	}

	fmt.Printf("Senha criptografada: %s\n", encryptedPassword)

	// Descriptografar senha
	decryptedPassword, err := manager.DecryptPassword(encryptedPassword)
	if err != nil {
		log.Printf("Erro ao descriptografar senha: %v", err)
		return
	}

	fmt.Printf("Senha descriptografada: %s\n", string(decryptedPassword))

	// Criptografar dados sensíveis
	sensitiveInfo := "CPF: 123.456.789-00, RG: 12.345.678-9"
	encryptedInfo, err := manager.EncryptSensitiveData(sensitiveInfo)
	if err != nil {
		log.Printf("Erro ao criptografar dados sensíveis: %v", err)
		return
	}

	fmt.Printf("Dados sensíveis criptografados: %s\n", encryptedInfo)

	// Descriptografar dados sensíveis
	decryptedInfo, err := manager.DecryptSensitiveData(encryptedInfo)
	if err != nil {
		log.Printf("Erro ao descriptografar dados sensíveis: %v", err)
		return
	}

	fmt.Printf("Dados sensíveis descriptografados: %s\n", string(decryptedInfo))
}

// ExampleRotationKeyUsage demonstra o uso de chaves de rotação
func ExampleRotationKeyUsage() {
	// Carregar chave de rotação
	rotationKeyPath := "/path/to/rotation.key"
	rotationKey, err := LoadAESKeyFromPath(rotationKeyPath)
	if err != nil {
		log.Printf("Erro ao carregar chave de rotação: %v", err)
		return
	}

	// Dados para criptografar
	data := "Dados que serão rotacionados periodicamente"

	// Criptografar com chave de rotação
	encrypted, err := EncryptWithRotationKey(rotationKey, []byte(data))
	if err != nil {
		log.Printf("Erro ao criptografar com chave de rotação: %v", err)
		return
	}

	fmt.Printf("Dados criptografados com rotação: %x\n", encrypted)

	// Descriptografar
	decrypted, err := DecryptWithRotationKey(rotationKey, encrypted)
	if err != nil {
		log.Printf("Erro ao descriptografar com chave de rotação: %v", err)
		return
	}

	fmt.Printf("Dados descriptografados: %s\n", string(decrypted))
}

// ExampleKeyGeneration demonstra como gerar chaves
func ExampleKeyGeneration() {
	// Gerar chave AES
	aesKey, err := generateAESKey()
	if err != nil {
		log.Printf("Erro ao gerar chave AES: %v", err)
		return
	}

	fmt.Printf("Chave AES gerada: %x\n", aesKey)
	fmt.Printf("Tamanho da chave: %d bytes\n", len(aesKey))

	// Exemplo de como salvar a chave (não execute em produção sem proteção adequada)
	/*
	err = os.WriteFile("/path/to/new_key.key", []byte(hex.EncodeToString(aesKey)), 0600)
	if err != nil {
		log.Printf("Erro ao salvar chave: %v", err)
		return
	}
	fmt.Println("Chave salva com sucesso")
	*/
}

// ExampleRSAKeyGeneration demonstra como gerar chaves RSA
func ExampleRSAKeyGeneration() {
	fmt.Println("=== Exemplo de Geração de Chaves RSA ===")

	// Gerar par de chaves RSA com tamanho padrão (2048 bits)
	keyPair, err := GenerateRSAKeyPairDefault()
	if err != nil {
		log.Printf("Erro ao gerar chaves RSA: %v", err)
		return
	}

	fmt.Println("Chave Privada RSA:")
	fmt.Println(keyPair.PrivateKey)
	fmt.Println("\nChave Pública RSA:")
	fmt.Println(keyPair.PublicKey)

	// Gerar par de chaves RSA com tamanho personalizado (4096 bits)
	keyPair4096, err := GenerateRSAKeyPair(4096)
	if err != nil {
		log.Printf("Erro ao gerar chaves RSA 4096: %v", err)
		return
	}

	fmt.Println("\n=== Chaves RSA 4096 bits geradas com sucesso ===")
	fmt.Printf("Tamanho da chave privada: %d caracteres\n", len(keyPair4096.PrivateKey))
	fmt.Printf("Tamanho da chave pública: %d caracteres\n", len(keyPair4096.PublicKey))
}

// ExampleRSAKeyGenerationWithService demonstra geração de chaves usando o CryptService
func ExampleRSAKeyGenerationWithService() {
	fmt.Println("=== Exemplo de Geração de Chaves RSA via CryptService ===")

	// Nota: Este exemplo assume que você já tem um CryptService inicializado
	// Para fins de demonstração, vamos mostrar como seria usado

	// Simular um serviço (normalmente você teria um serviço já inicializado)
	// cryptService := &CryptService{} // Normalmente inicializado com Initialize()

	// Gerar chaves usando o serviço
	// keyPair, err := cryptService.GenerateRSAKeysDefault()
	// if err != nil {
	//     log.Printf("Erro ao gerar chaves via serviço: %v", err)
	//     return
	// }

	fmt.Println("Para usar com CryptService:")
	fmt.Println("keyPair, err := cryptService.GenerateRSAKeysDefault()")
	fmt.Println("// ou")
	fmt.Println("keyPair, err := cryptService.GenerateRSAKeys(4096)")

	// Demonstração direta sem serviço
	keyPair, err := GenerateRSAKeyPairDefault()
	if err != nil {
		log.Printf("Erro ao gerar chaves: %v", err)
		return
	}

	fmt.Println("\nChaves geradas com sucesso!")
	fmt.Printf("Chave privada começa com: %.50s...\n", keyPair.PrivateKey)
	fmt.Printf("Chave pública começa com: %.50s...\n", keyPair.PublicKey)
}

// ExampleHybridEncryptionWithKeys demonstra criptografia híbrida usando chaves fornecidas
func ExampleHybridEncryptionWithKeys() {
	fmt.Println("=== Exemplo de Criptografia Híbrida com Chaves Fornecidas ===")

	// Gerar um par de chaves para o exemplo
	keyPair, err := GenerateRSAKeyPairDefault()
	if err != nil {
		log.Printf("Erro ao gerar chaves: %v", err)
		return
	}

	// Converter as chaves PEM de volta para objetos RSA
	publicKey, err := LoadRSAPublicKeyFromPEM(keyPair.PublicKey)
	if err != nil {
		log.Printf("Erro ao carregar chave pública: %v", err)
		return
	}

	privateKey, err := LoadRSAPrivateKeyFromPEM(keyPair.PrivateKey)
	if err != nil {
		log.Printf("Erro ao carregar chave privada: %v", err)
		return
	}

	// Dados para criptografar
	data := "Dados confidenciais para criptografia híbrida com chaves fornecidas"

	// Criptografar usando função global
	encrypted, err := HybridEncryptWithKeys(data, publicKey)
	if err != nil {
		log.Printf("Erro ao criptografar: %v", err)
		return
	}

	fmt.Printf("Dados originais: %s\n", data)
	fmt.Printf("Dados criptografados: %.100s...\n", encrypted)

	// Descriptografar usando função global
	decrypted, err := HybridDecryptWithKeys(encrypted, privateKey)
	if err != nil {
		log.Printf("Erro ao descriptografar: %v", err)
		return
	}

	fmt.Printf("Dados descriptografados: %s\n", string(decrypted))

	// Verificar se os dados são iguais
	if string(decrypted) == data {
		fmt.Println("✅ Criptografia e descriptografia funcionaram corretamente!")
	} else {
		fmt.Println("❌ Erro: dados descriptografados não coincidem com os originais")
	}
}

// ExampleHybridEncryptionWithCryptService demonstra uso via CryptService
func ExampleHybridEncryptionWithCryptService() {
	fmt.Println("=== Exemplo de Criptografia Híbrida via CryptService ===")

	// Para este exemplo, vamos simular o uso com chaves geradas
	keyPair, err := GenerateRSAKeyPairDefault()
	if err != nil {
		log.Printf("Erro ao gerar chaves: %v", err)
		return
	}

	// Converter chaves PEM para objetos RSA
	publicKey, err := LoadRSAPublicKeyFromPEM(keyPair.PublicKey)
	if err != nil {
		log.Printf("Erro ao carregar chave pública: %v", err)
		return
	}

	privateKey, err := LoadRSAPrivateKeyFromPEM(keyPair.PrivateKey)
	if err != nil {
		log.Printf("Erro ao carregar chave privada: %v", err)
		return
	}

	// Criar uma instância do CryptService (simulado)
	cryptService := &CryptService{}

	// Dados para teste
	data := "Teste de criptografia híbrida via CryptService"

	// Usar métodos do CryptService
	encrypted, err := cryptService.HybridEncryptWithKeys(data, publicKey)
	if err != nil {
		log.Printf("Erro ao criptografar via CryptService: %v", err)
		return
	}

	fmt.Printf("Dados criptografados via CryptService: %.100s...\n", encrypted)

	// Descriptografar
	decrypted, err := cryptService.HybridDecryptWithKeys(encrypted, privateKey)
	if err != nil {
		log.Printf("Erro ao descriptografar via CryptService: %v", err)
		return
	}

	fmt.Printf("Dados descriptografados: %s\n", string(decrypted))
	fmt.Println("✅ Exemplo concluído com sucesso!")
}

// ExampleEncryptedPayload demonstra o uso da estrutura EncryptedPayload
func ExampleEncryptedPayload() {
	// Exemplo de payload criptografado
	payload := EncryptedPayload{
		EncryptedKey: "base64-encoded-encrypted-aes-key",
		Nonce:        "base64-encoded-nonce",
		Ciphertext:   "base64-encoded-encrypted-data",
	}

	fmt.Printf("Payload criptografado:\n")
	fmt.Printf("  Chave criptografada: %s\n", payload.EncryptedKey)
	fmt.Printf("  Nonce: %s\n", payload.Nonce)
	fmt.Printf("  Texto cifrado: %s\n", payload.Ciphertext)

	// Em um cenário real, você converteria isso para JSON
	/*
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Erro ao serializar payload: %v", err)
		return
	}
	fmt.Printf("Payload JSON: %s\n", string(jsonData))
	*/
}

// ExampleSecurityBestPractices demonstra boas práticas de segurança
func ExampleSecurityBestPractices() {
	fmt.Println("=== BOAS PRÁTICAS DE SEGURANÇA ===")
	fmt.Println("")
	fmt.Println("1. GERENCIAMENTO DE CHAVES:")
	fmt.Println("   - Armazene chaves em local seguro (HSM, Key Vault, etc.)")
	fmt.Println("   - Use permissões restritivas (600) para arquivos de chave")
	fmt.Println("   - Implemente rotação regular de chaves")
	fmt.Println("   - Nunca hardcode chaves no código")
	fmt.Println("")
	fmt.Println("2. CRIPTOGRAFIA:")
	fmt.Println("   - Use AES-256 para criptografia simétrica")
	fmt.Println("   - Use RSA-2048 ou superior para criptografia assimétrica")
	fmt.Println("   - Sempre use modos autenticados (GCM)")
	fmt.Println("   - Gere nonces/IVs únicos para cada operação")
	fmt.Println("")
	fmt.Println("3. TRATAMENTO DE ERROS:")
	fmt.Println("   - Não exponha detalhes de criptografia em logs")
	fmt.Println("   - Limpe dados sensíveis da memória após uso")
	fmt.Println("   - Implemente timeouts para operações")
	fmt.Println("")
	fmt.Println("4. AUDITORIA:")
	fmt.Println("   - Registre operações de criptografia/descriptografia")
	fmt.Println("   - Monitore tentativas de acesso a chaves")
	fmt.Println("   - Implemente alertas para falhas de segurança")
}

// ExampleErrorHandling demonstra tratamento de erros
func ExampleErrorHandling() {
	// Exemplo de tratamento de erro ao carregar chave inexistente
	_, err := LoadAESKeyFromPath("/path/inexistente/chave.key")
	if err != nil {
		fmt.Printf("Erro esperado ao carregar chave inexistente: %v\n", err)
	}

	// Exemplo de tratamento de erro com chave RSA inválida
	_, err = LoadRSAPrivateKeyFromPath("/path/inexistente/private.pem")
	if err != nil {
		fmt.Printf("Erro esperado ao carregar chave RSA inexistente: %v\n", err)
	}

	// Exemplo de inicialização com parâmetros inválidos
	_, err = Initialize("", "", "", "")
	if err != nil {
		fmt.Printf("Erro esperado com parâmetros vazios: %v\n", err)
	}

	fmt.Println("\nSempre trate erros adequadamente em produção!")
}