# Pacote Crypt

O pacote `crypt` fornece funcionalidades completas de criptografia, incluindo criptografia simétrica (AES), assimétrica (RSA), híbrida e gerenciamento de chaves com rotação.

## Funcionalidades

### 🔐 Criptografia AES (Simétrica)
- Criptografia AES-256-GCM
- Geração automática de chaves
- Nonces únicos para cada operação
- Autenticação integrada (AEAD)

### 🔑 Criptografia RSA (Assimétrica)
- Suporte a chaves RSA de 2048, 3072 e 4096 bits
- Criptografia e descriptografia de dados
- Carregamento de chaves de arquivos PEM
- Geração de pares de chaves

### 🔄 Criptografia Híbrida
- Combinação de RSA + AES para melhor performance
- Criptografia de chaves AES com RSA
- Criptografia de dados com AES
- Ideal para grandes volumes de dados

### 🛡️ Gerenciamento de Chaves
- Chaves mestras e de rotação
- Rotação automática de chaves
- Versionamento de chaves
- Armazenamento seguro

### 🏢 Serviços de Alto Nível
- `CryptService`: Serviço completo com carregamento de chaves
- `CryptManager`: Gerenciador para senhas e dados sensíveis
- Configuração via arquivos

## Estruturas Principais

### `EncryptedPayload`
```go
type EncryptedPayload struct {
    Data      []byte `json:"data"`
    Nonce     []byte `json:"nonce"`
    KeyID     string `json:"key_id,omitempty"`
    Algorithm string `json:"algorithm,omitempty"`
}
```

Estrutura que encapsula dados criptografados com metadados.

### `CryptService`
```go
type CryptService struct {
    publicKey  *rsa.PublicKey
    privateKey *rsa.PrivateKey
    masterKey  []byte
    rotationKey []byte
}
```

Serviço principal para operações de criptografia.

### `CryptManager`
```go
type CryptManager struct {
    masterKey []byte
}
```

Gerenciador simplificado para operações básicas.

## Configuração

### Inicialização do CryptService
```go
// Carregamento automático de chaves de arquivos
service, err := crypt.Initialize()
if err != nil {
    log.Fatal("Erro ao inicializar serviço de criptografia:", err)
}

// O serviço procura pelos arquivos:
// - private_key.pem (chave privada RSA)
// - public_key.pem (chave pública RSA)
// - master.key (chave mestra AES)
// - rotation.key (chave de rotação AES)
```

### Inicialização do CryptManager
```go
// Com chave mestra específica
manager := crypt.NewCryptManager(masterKey)

// Com chave gerada automaticamente
manager := crypt.NewCryptManager(nil) // Gera chave automaticamente
```

## Criptografia AES

### Operações Básicas
```go
// Gerar chave AES
key, err := crypt.GenerateAESKey()
if err != nil {
    log.Fatal(err)
}

// Criptografar dados
data := []byte("dados sensíveis")
encrypted, err := crypt.EncryptAES(data, key)
if err != nil {
    log.Fatal(err)
}

// Descriptografar dados
decrypted, err := crypt.DecryptAES(encrypted, key)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados originais: %s\n", string(decrypted))
```

### Usando EncryptedPayload
```go
data := []byte("informação confidencial")
key, _ := crypt.GenerateAESKey()

// Criptografar com payload estruturado
payload, err := crypt.EncryptAESWithPayload(data, key, "key-001", "AES-256-GCM")
if err != nil {
    log.Fatal(err)
}

// Serializar para JSON
jsonData, _ := json.Marshal(payload)
fmt.Printf("Payload criptografado: %s\n", string(jsonData))

// Descriptografar do payload
decrypted, err := crypt.DecryptAESFromPayload(payload, key)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados descriptografados: %s\n", string(decrypted))
```

## Criptografia RSA

### Geração de Chaves
```go
// Gerar par de chaves RSA
privateKey, publicKey, err := crypt.GenerateRSAKeyPair(2048)
if err != nil {
    log.Fatal(err)
}

// Salvar chaves em arquivos
err = crypt.SaveRSAPrivateKeyToFile(privateKey, "private_key.pem")
if err != nil {
    log.Fatal(err)
}

err = crypt.SaveRSAPublicKeyToFile(publicKey, "public_key.pem")
if err != nil {
    log.Fatal(err)
}
```

### Carregamento de Chaves
```go
// Carregar chave privada
privateKey, err := crypt.LoadRSAPrivateKeyFromFile("private_key.pem")
if err != nil {
    log.Fatal(err)
}

// Carregar chave pública
publicKey, err := crypt.LoadRSAPublicKeyFromFile("public_key.pem")
if err != nil {
    log.Fatal(err)
}
```

### Criptografia e Descriptografia
```go
data := []byte("dados para criptografar")

// Criptografar com chave pública
encrypted, err := crypt.EncryptRSA(data, publicKey)
if err != nil {
    log.Fatal(err)
}

// Descriptografar com chave privada
decrypted, err := crypt.DecryptRSA(encrypted, privateKey)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados descriptografados: %s\n", string(decrypted))
```

## Criptografia Híbrida

### RSA + AES
```go
// Dados grandes para criptografar
largeData := make([]byte, 1024*1024) // 1MB
for i := range largeData {
    largeData[i] = byte(i % 256)
}

// Criptografia híbrida (RSA + AES)
encryptedData, encryptedKey, err := crypt.EncryptHybrid(largeData, publicKey)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados criptografados: %d bytes\n", len(encryptedData))
fmt.Printf("Chave criptografada: %d bytes\n", len(encryptedKey))

// Descriptografia híbrida
decrypted, err := crypt.DecryptHybrid(encryptedData, encryptedKey, privateKey)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados descriptografados: %d bytes\n", len(decrypted))
```

## CryptService - Serviço Completo

### Inicialização e Uso
```go
// Inicializar serviço
service, err := crypt.Initialize()
if err != nil {
    log.Fatal(err)
}

// Criptografar com chave mestra
data := []byte("dados confidenciais")
encrypted, err := service.EncryptWithMasterKey(data)
if err != nil {
    log.Fatal(err)
}

// Descriptografar com chave mestra
decrypted, err := service.DecryptWithMasterKey(encrypted)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados: %s\n", string(decrypted))
```

### Rotação de Chaves
```go
// Criptografar com chave de rotação
encryptedWithRotation, err := service.EncryptWithRotationKey(data)
if err != nil {
    log.Fatal(err)
}

// Descriptografar com chave de rotação
decryptedFromRotation, err := service.DecryptWithRotationKey(encryptedWithRotation)
if err != nil {
    log.Fatal(err)
}

// Migrar dados da chave mestra para chave de rotação
migratedData, err := service.MigrateToRotationKey(encrypted)
if err != nil {
    log.Fatal(err)
}
```

### Criptografia Híbrida no Serviço
```go
// Criptografia híbrida usando o serviço
largeData := []byte("dados muito grandes...")

encryptedData, encryptedKey, err := service.EncryptHybridData(largeData)
if err != nil {
    log.Fatal(err)
}

// Descriptografia híbrida
decrypted, err := service.DecryptHybridData(encryptedData, encryptedKey)
if err != nil {
    log.Fatal(err)
}
```

## CryptManager - Gerenciamento Simplificado

### Operações com Senhas
```go
manager := crypt.NewCryptManager(nil) // Chave gerada automaticamente

// Criptografar senha
password := "minha-senha-secreta"
encryptedPassword, err := manager.EncryptPassword(password)
if err != nil {
    log.Fatal(err)
}

// Descriptografar senha
decryptedPassword, err := manager.DecryptPassword(encryptedPassword)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Senha original: %s\n", decryptedPassword)
```

### Operações com Dados Sensíveis
```go
// Criptografar dados sensíveis
sensitiveData := map[string]interface{}{
    "ssn": "123-45-6789",
    "credit_card": "4111-1111-1111-1111",
    "bank_account": "987654321",
}

jsonData, _ := json.Marshal(sensitiveData)
encryptedData, err := manager.EncryptSensitiveData(jsonData)
if err != nil {
    log.Fatal(err)
}

// Descriptografar dados sensíveis
decryptedData, err := manager.DecryptSensitiveData(encryptedData)
if err != nil {
    log.Fatal(err)
}

var originalData map[string]interface{}
json.Unmarshal(decryptedData, &originalData)
fmt.Printf("Dados originais: %+v\n", originalData)
```

## Exemplos Avançados

### Sistema de Backup Criptografado
```go
package main

import (
    "encoding/json"
    "fmt"
    "time"
    "seu-projeto/initializers/crypt"
)

type BackupData struct {
    Timestamp time.Time `json:"timestamp"`
    UserData  []User    `json:"user_data"`
    Settings  map[string]interface{} `json:"settings"`
}

type User struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Password string `json:"password"` // Será criptografado
    SSN      string `json:"ssn"`      // Será criptografado
}

func createEncryptedBackup() {
    service, err := crypt.Initialize()
    if err != nil {
        panic(err)
    }
    
    // Dados do backup
    backup := BackupData{
        Timestamp: time.Now(),
        UserData: []User{
            {
                ID:       "1",
                Email:    "user1@example.com",
                Password: "senha123",
                SSN:      "123-45-6789",
            },
            {
                ID:       "2",
                Email:    "user2@example.com",
                Password: "outrasenha",
                SSN:      "987-65-4321",
            },
        },
        Settings: map[string]interface{}{
            "app_version": "1.0.0",
            "db_version":  "2.1.0",
        },
    }
    
    // Criptografar dados sensíveis
    for i := range backup.UserData {
        // Criptografar senha
        encryptedPassword, err := service.EncryptWithMasterKey([]byte(backup.UserData[i].Password))
        if err != nil {
            panic(err)
        }
        backup.UserData[i].Password = string(encryptedPassword)
        
        // Criptografar SSN
        encryptedSSN, err := service.EncryptWithMasterKey([]byte(backup.UserData[i].SSN))
        if err != nil {
            panic(err)
        }
        backup.UserData[i].SSN = string(encryptedSSN)
    }
    
    // Serializar backup
    backupJSON, err := json.Marshal(backup)
    if err != nil {
        panic(err)
    }
    
    // Criptografar backup completo com criptografia híbrida
    encryptedData, encryptedKey, err := service.EncryptHybridData(backupJSON)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Backup criptografado criado:\n")
    fmt.Printf("- Dados: %d bytes\n", len(encryptedData))
    fmt.Printf("- Chave: %d bytes\n", len(encryptedKey))
    
    // Salvar em arquivos
    // saveToFile("backup.dat", encryptedData)
    // saveToFile("backup.key", encryptedKey)
}
```

### Middleware de Criptografia para APIs
```go
package middleware

import (
    "bytes"
    "encoding/json"
    "io"
    "github.com/gin-gonic/gin"
    "seu-projeto/initializers/crypt"
)

// CryptMiddleware criptografa automaticamente campos sensíveis
func CryptMiddleware(service *crypt.CryptService, sensitiveFields []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Interceptar request body
        if c.Request.Method == "POST" || c.Request.Method == "PUT" {
            body, err := io.ReadAll(c.Request.Body)
            if err != nil {
                c.JSON(500, gin.H{"error": "Erro ao ler request"})
                c.Abort()
                return
            }
            
            // Parse JSON
            var data map[string]interface{}
            if err := json.Unmarshal(body, &data); err == nil {
                // Criptografar campos sensíveis
                for _, field := range sensitiveFields {
                    if value, exists := data[field]; exists {
                        if strValue, ok := value.(string); ok {
                            encrypted, err := service.EncryptWithMasterKey([]byte(strValue))
                            if err == nil {
                                data[field] = string(encrypted)
                            }
                        }
                    }
                }
                
                // Recriar request body
                newBody, _ := json.Marshal(data)
                c.Request.Body = io.NopCloser(bytes.NewReader(newBody))
                c.Request.ContentLength = int64(len(newBody))
            } else {
                // Restaurar body original se não for JSON
                c.Request.Body = io.NopCloser(bytes.NewReader(body))
            }
        }
        
        c.Next()
    }
}

// Uso
func setupRoutes() {
    service, _ := crypt.Initialize()
    
    r := gin.Default()
    
    // Aplicar middleware para criptografar campos sensíveis
    sensitiveFields := []string{"password", "ssn", "credit_card", "bank_account"}
    r.Use(CryptMiddleware(service, sensitiveFields))
    
    r.POST("/users", createUser)
    r.PUT("/users/:id", updateUser)
    
    r.Run(":8080")
}
```

## Segurança

### Boas Práticas

1. **Gerenciamento de Chaves**
   ```go
   // ❌ Não hardcode chaves
   key := []byte("minha-chave-123")
   
   // ✅ Use variáveis de ambiente ou arquivos seguros
   keyPath := os.Getenv("MASTER_KEY_PATH")
   key, err := crypt.LoadAESKeyFromFile(keyPath)
   ```

2. **Rotação de Chaves**
   ```go
   // Implementar rotação periódica
   func rotateKeys(service *crypt.CryptService) {
       // Gerar nova chave de rotação
       newKey, err := crypt.GenerateAESKey()
       if err != nil {
           log.Fatal(err)
       }
       
       // Salvar nova chave
       err = crypt.SaveAESKeyToFile(newKey, "new_rotation.key")
       if err != nil {
           log.Fatal(err)
       }
       
       // Migrar dados existentes
       // ... lógica de migração
   }
   ```

3. **Validação de Integridade**
   ```go
   func validateEncryptedData(payload *crypt.EncryptedPayload) error {
       if len(payload.Data) == 0 {
           return errors.New("dados criptografados vazios")
       }
       
       if len(payload.Nonce) != 12 { // GCM nonce size
           return errors.New("nonce inválido")
       }
       
       if payload.Algorithm != "AES-256-GCM" {
           return errors.New("algoritmo não suportado")
       }
       
       return nil
   }
   ```

4. **Limpeza de Memória**
   ```go
   func secureCleanup(sensitiveData []byte) {
       // Sobrescrever dados sensíveis na memória
       for i := range sensitiveData {
           sensitiveData[i] = 0
       }
   }
   
   // Uso
   password := []byte("senha-secreta")
   defer secureCleanup(password)
   
   // ... usar password
   ```

### Configuração de Arquivos

#### Estrutura de Diretórios
```
project/
├── keys/
│   ├── private_key.pem     # Chave privada RSA
│   ├── public_key.pem      # Chave pública RSA
│   ├── master.key          # Chave mestra AES
│   └── rotation.key        # Chave de rotação AES
└── config/
    └── crypt.yaml          # Configurações
```

#### Permissões de Arquivos
```bash
# Definir permissões restritivas
chmod 600 keys/private_key.pem
chmod 600 keys/master.key
chmod 600 keys/rotation.key
chmod 644 keys/public_key.pem

# Proprietário apenas para diretório
chmod 700 keys/
```

## Performance

### Benchmarks Típicos

- **AES-256-GCM**: ~500 MB/s para criptografia/descriptografia
- **RSA-2048**: ~1000 operações/s para criptografia, ~100 operações/s para descriptografia
- **Híbrida**: Performance próxima ao AES para dados grandes

### Otimizações

1. **Pool de Chaves**
   ```go
   type KeyPool struct {
       keys [][]byte
       index int
       mutex sync.RWMutex
   }
   
   func (p *KeyPool) GetKey() []byte {
       p.mutex.RLock()
       defer p.mutex.RUnlock()
       
       key := p.keys[p.index]
       p.index = (p.index + 1) % len(p.keys)
       return key
   }
   ```

2. **Cache de Chaves RSA**
   ```go
   var (
       rsaKeyCache = make(map[string]*rsa.PrivateKey)
       cacheMutex  sync.RWMutex
   )
   
   func getCachedRSAKey(keyPath string) (*rsa.PrivateKey, error) {
       cacheMutex.RLock()
       if key, exists := rsaKeyCache[keyPath]; exists {
           cacheMutex.RUnlock()
           return key, nil
       }
       cacheMutex.RUnlock()
       
       // Carregar e cachear chave
       key, err := crypt.LoadRSAPrivateKeyFromFile(keyPath)
       if err != nil {
           return nil, err
       }
       
       cacheMutex.Lock()
       rsaKeyCache[keyPath] = key
       cacheMutex.Unlock()
       
       return key, nil
   }
   ```

## Testes

### Testes Unitários
```go
package crypt_test

import (
    "testing"
    "seu-projeto/initializers/crypt"
)

func TestAESEncryptionDecryption(t *testing.T) {
    key, err := crypt.GenerateAESKey()
    if err != nil {
        t.Fatal(err)
    }
    
    originalData := []byte("dados de teste")
    
    // Criptografar
    encrypted, err := crypt.EncryptAES(originalData, key)
    if err != nil {
        t.Fatal(err)
    }
    
    // Descriptografar
    decrypted, err := crypt.DecryptAES(encrypted, key)
    if err != nil {
        t.Fatal(err)
    }
    
    if string(decrypted) != string(originalData) {
        t.Errorf("Dados não coincidem. Original: %s, Descriptografado: %s", 
                 string(originalData), string(decrypted))
    }
}

func BenchmarkAESEncryption(b *testing.B) {
    key, _ := crypt.GenerateAESKey()
    data := make([]byte, 1024) // 1KB
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := crypt.EncryptAES(data, key)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Dependências

- `crypto/aes` - Criptografia AES
- `crypto/rsa` - Criptografia RSA
- `crypto/rand` - Geração de números aleatórios
- `crypto/cipher` - Modos de operação de cifra
- `encoding/pem` - Codificação PEM para chaves

## Veja Também

- [Pacote Auth](../auth/README.md) - Para autenticação com criptografia
- [Pacote Validator](../validator/README.md) - Para validação de dados
- [Pacote Formatter](../formatter/README.md) - Para formatação de respostas

---

**Nota**: Este pacote implementa algoritmos criptográficos padrão da indústria. Sempre mantenha as chaves seguras e implemente rotação regular em ambientes de produção.