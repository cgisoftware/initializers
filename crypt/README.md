# Pacote Crypt

O pacote `crypt` fornece funcionalidades de criptografia simétrica (AES-256-GCM),
assimétrica (RSA), híbrida (RSA + AES) e um middleware HTTP de descriptografia
automática de campos de requisição.

## Funcionalidades

### 🔐 Criptografia AES (Simétrica)
- AES-256-GCM (autenticado — detecta adulteração do ciphertext)
- Geração de chaves via `GenerateAESKey`
- Nonce único gerado para cada operação

### 🔑 Criptografia RSA (Assimétrica)
- Geração de pares de chaves RSA (mínimo 2048 bits)
- Carregamento de chaves de arquivos PEM ou de strings PEM
- Exportação de chaves em formato PEM (PKCS1 para a privada, PKIX para a pública)

### 🔄 Criptografia Híbrida
- RSA criptografa uma chave AES gerada na hora; AES-GCM criptografa os dados
- Ideal para dados grandes (evita o limite de tamanho do RSA)
- Payload serializado em JSON (`EncryptedPayload`) e depois em base64

### 🛡️ Gerenciamento de Chaves
- Chave mestra (`masterKey`) e chave de rotação (`rotationKey`) independentes
- Carregamento a partir de arquivos hexadecimais (`LoadAESKeyFromPath`)

### 🏢 Serviços de Alto Nível
- `CryptService`: serviço completo, inicializado a partir de caminhos de arquivo de chave
- `CryptManager`: wrapper simplificado para senhas e dados sensíveis
- `DecryptionMiddleware`: middleware HTTP que descriptografa campos JSON automaticamente

## Estruturas Principais

### `EncryptedPayload`
```go
type EncryptedPayload struct {
    EncryptedKey string `json:"encrypted_key"` // AES key criptografada com RSA
    Nonce        string `json:"nonce"`         // Nonce do AES-GCM
    Ciphertext   string `json:"ciphertext"`    // Dados criptografados com AES
}
```

### `CryptService`
```go
type CryptService struct {
    privateKey  *rsa.PrivateKey
    publicKey   *rsa.PublicKey
    masterKey   []byte
    rotationKey []byte
}
```

### `CryptManager`
```go
type CryptManager struct {
    hybridService CryptService
}
```

### `RSAKeyPair`
```go
type RSAKeyPair struct {
    PrivateKey string `json:"private_key"` // Chave privada em formato PEM
    PublicKey  string `json:"public_key"`  // Chave pública em formato PEM
}
```

## Criptografia AES

### Gerar e usar uma chave AES
```go
key, err := crypt.GenerateAESKey()
if err != nil {
    log.Fatal(err)
}

data := []byte("dados sensíveis")

encrypted, err := crypt.EncryptWithMasterKey(key, data)
if err != nil {
    log.Fatal(err)
}

decrypted, err := crypt.DecryptWithMasterKey(key, encrypted)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados originais: %s\n", string(decrypted))
```

`EncryptWithRotationKey`/`DecryptWithRotationKey` funcionam da mesma forma,
usando uma segunda chave dedicada à rotação periódica.

### Carregar chave AES de um arquivo
```go
// O arquivo deve conter a chave em hexadecimal (32 bytes = 64 caracteres hex)
key, err := crypt.LoadAESKeyFromPath("/path/to/master.key")
if err != nil {
    log.Fatal(err)
}
```

## Criptografia RSA

### Geração de Chaves RSA
```go
// Tamanho padrão (2048 bits)
keyPair, err := crypt.GenerateRSAKeyPairDefault()
if err != nil {
    log.Fatal(err)
}

// Tamanho customizado (mínimo 2048 bits)
keyPair4096, err := crypt.GenerateRSAKeyPair(4096)
if err != nil {
    log.Fatal(err)
}

fmt.Println(keyPair.PrivateKey) // PEM
fmt.Println(keyPair.PublicKey)  // PEM
```

### Carregamento de Chaves
```go
// De arquivo
privateKey, err := crypt.LoadRSAPrivateKeyFromPath("private_key.pem")
publicKey, err := crypt.LoadRSAPublicKeyFromPath("public_key.pem")

// De string PEM (ex.: chave gerada em memória)
privateKey, err = crypt.LoadRSAPrivateKeyFromPEM(keyPair.PrivateKey)
publicKey, err = crypt.LoadRSAPublicKeyFromPEM(keyPair.PublicKey)
```

## Criptografia Híbrida (RSA + AES)

```go
data := []byte("dados grandes o suficiente para justificar o modo híbrido")

// Criptografar: gera uma chave AES efêmera, criptografa os dados com ela,
// e criptografa a chave AES com a chave pública RSA
encrypted, err := crypt.HybridEncrypt(publicKey, data)
if err != nil {
    log.Fatal(err)
}

decrypted, err := crypt.HybridDecrypt(privateKey, encrypted)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Dados descriptografados: %s\n", string(decrypted))
```

### Variante com dados em string + base64 (funções `*WithKeys`)
```go
encryptedBase64, err := crypt.HybridEncryptWithKeys("dados confidenciais", publicKey)
if err != nil {
    log.Fatal(err)
}

decrypted, err := crypt.HybridDecryptWithKeys(encryptedBase64, privateKey)
if err != nil {
    log.Fatal(err)
}
```

## Tokens simétricos (`GenerateToken`/`DecryptToken`)

Criptografia AES-GCM de propósito geral para gerar tokens opacos (ex.: para
armazenar um identificador criptografado em cookie ou header). O token é a
concatenação `base64(ciphertext) + "-" + base64(nonce)`.

```go
ctx := context.Background()
key, _ := crypt.GenerateAESKey()

token, err := crypt.GenerateToken(ctx, key, []byte("payload"))
if err != nil {
    log.Fatal(err)
}

payload, err := crypt.DecryptToken(ctx, key, token)
if err != nil {
    log.Fatal(err) // também falha se o token foi adulterado (falha de autenticação do GCM)
}
```

## CryptService — Serviço Completo

### Inicialização
```go
service, err := crypt.Initialize(
    "private_key.pem",  // chave privada RSA
    "public_key.pem",   // chave pública RSA
    "master.key",        // chave AES mestra (hex)
    "rotation.key",       // chave AES de rotação (hex)
)
if err != nil {
    log.Fatal(err)
}
```

Todos os quatro caminhos são obrigatórios — `Initialize` retorna erro se
qualquer um estiver vazio ou se o arquivo correspondente não puder ser lido.

### Métodos disponíveis
```go
// Híbrida (RSA + AES), usando as chaves carregadas no serviço
encrypted, err := service.EncryptData("dados confidenciais")
decrypted, err := service.DecryptData(encrypted)

// Simétrica com a chave mestra
encrypted, err = service.EncryptWithMasterKeySimple("dados sensíveis")
decrypted, err = service.DecryptWithMasterKeySimple(encrypted)

// Geração de chaves RSA a partir do serviço
keyPair, err := service.GenerateRSAKeysDefault()
keyPair4096, err := service.GenerateRSAKeys(4096)

// Híbrida com chaves fornecidas explicitamente (não as do serviço)
encrypted, err = service.HybridEncryptWithKeys("dados", outraChavePublica)
decrypted, err = service.HybridDecryptWithKeys(encrypted, outraChavePrivada)
```

## CryptManager — Gerenciamento Simplificado

```go
service, err := crypt.Initialize(privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath)
if err != nil {
    log.Fatal(err)
}

manager := crypt.NewCryptManager(service)

// Senhas (usa a chave mestra AES)
encryptedPassword, err := manager.EncryptPassword("minha-senha-secreta")
decryptedPassword, err := manager.DecryptPassword(encryptedPassword)

// Dados sensíveis (usa criptografia híbrida)
encryptedInfo, err := manager.EncryptSensitiveData("CPF: 123.456.789-00")
decryptedInfo, err := manager.DecryptSensitiveData(encryptedInfo)
```

## Middleware de Descriptografia HTTP

`DecryptionMiddleware` intercepta requisições `POST`/`PUT`/`PATCH` com
`Content-Type: application/json` e descriptografa automaticamente os campos
configurados antes de repassar ao próximo handler.

```go
cryptService, err := crypt.Initialize(privateKeyPath, publicKeyPath, masterKeyPath, rotationKeyPath)
if err != nil {
    log.Fatal(err)
}

dm := crypt.NewDecryptionMiddleware(&cryptService, []string{"password", "ssn"}, "hybrid") // ou "aes"

mux := http.NewServeMux()
mux.HandleFunc("/users", createUserHandler)

handler := dm.MiddlewareFunc()(mux)
http.ListenAndServe(":8080", handler)
```

Ou a partir de uma configuração:
```go
dm, err := crypt.NewDecryptionMiddlewareFromConfig(crypt.DecryptionConfig{
    EncryptedFields:    []string{"password", "ssn"},
    DecryptionType:     "hybrid",
    RSAPrivateKeyPath:  privateKeyPath,
    RSAPublicKeyPath:   publicKeyPath,
    AESMasterKeyPath:   masterKeyPath,
    AESRotationKeyPath: rotationKeyPath,
})
```

`MiddlewareFunc()` retorna um `func(http.Handler) http.Handler` padrão da
biblioteca `net/http`, compatível com qualquer roteador que aceite esse tipo
(chi, gorilla/mux, ou como adaptador manual em Gin/Echo).

## Segurança

### Boas Práticas

1. **Gerenciamento de Chaves**
   - Nunca hardcode chaves no código-fonte.
   - Armazene os arquivos de chave (`master.key`, `rotation.key`, `*.pem`) fora
     do controle de versão, com permissões restritivas.
     ```bash
     chmod 600 private_key.pem master.key rotation.key
     chmod 644 public_key.pem
     ```
   - Implemente rotação periódica migrando dados criptografados com
     `EncryptWithRotationKey`/`DecryptWithRotationKey` conforme a política da
     aplicação.

2. **Criptografia**
   - Use AES-256-GCM (já é o padrão de todas as funções simétricas do pacote)
     — nunca reintroduza um modo não autenticado (CBC/CTR sem MAC).
   - Use RSA de 2048 bits ou mais (`GenerateRSAKeyPair` já rejeita tamanhos
     menores).
   - Para dados grandes, prefira sempre a criptografia híbrida
     (`HybridEncrypt`) em vez de RSA puro.

3. **Tratamento de Erros**
   - Nunca logue o conteúdo de chaves ou de dados descriptografados.
   - Trate falhas de descriptografia (`DecryptWithMasterKey`,
     `DecryptToken`, etc.) como possível tentativa de adulteração — o AES-GCM
     rejeita ciphertexts modificados automaticamente.

## Testes

O pacote tem testes unitários (`crypt_test.go`, `crypt_service_test.go`)
cobrindo round-trip de todas as primitivas (AES, RSA, híbrida, token) e
rejeição de ciphertexts/tokens adulterados. Rode com:

```bash
go test ./...
```

## Dependências

- `crypto/aes`, `crypto/cipher` — AES-GCM
- `crypto/rsa` — RSA e OAEP
- `crypto/rand` — geração de números aleatórios
- `crypto/x509`, `encoding/pem` — serialização de chaves

## Veja Também

- [Pacote Auth](../auth/README.md) - Para autenticação com criptografia
- [Pacote Validator](../validator/README.md) - Para validação de dados
- [Pacote Formatter](../formatter/README.md) - Para formatação de respostas

---

**Nota**: Este pacote implementa algoritmos criptográficos padrão da indústria.
Sempre mantenha as chaves seguras e implemente rotação regular em ambientes de
produção.
