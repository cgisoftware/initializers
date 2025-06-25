# Pacote Crypt - Middleware de Descriptografia HTTP

Este pacote fornece funcionalidades de criptografia/descriptografia e um middleware HTTP para descriptografar automaticamente dados em requisições.

## Funcionalidades

- **Criptografia Híbrida**: Combina RSA e AES para segurança e performance
- **Middleware HTTP**: Descriptografa automaticamente campos específicos em requisições JSON
- **Suporte a múltiplos tipos**: AES e criptografia híbrida
- **Configuração flexível**: Permite especificar quais campos descriptografar

## Instalação

```bash
go get github.com/cgisoftware/initializers/crypt
```

## Configuração Inicial

Antes de usar o pacote, você precisa inicializar o serviço de criptografia com os caminhos para suas chaves:

```go
package main

import (
    "log"
    "github.com/cgisoftware/initializers/crypt"
)

func main() {
    // Inicializa o serviço de criptografia
    cryptService, err := crypt.Initialize(
        "/path/to/private.pem",     // Chave RSA privada
        "/path/to/public.pem",      // Chave RSA pública
        "/path/to/master.key",      // Chave AES master
        "/path/to/rotation.key",    // Chave AES de rotação
    )
    if err != nil {
        log.Fatal("Erro ao inicializar serviço de criptografia:", err)
    }
    
    // Agora você pode usar o cryptService...
}
```

## Uso do Middleware

### Exemplo Básico

```go
package main

import (
    "encoding/json"
    "net/http"
    "log"
    "github.com/cgisoftware/initializers/crypt"
)

func main() {
    // Inicializa o serviço de criptografia
    cryptService, err := crypt.Initialize(
        "/path/to/private.pem",
        "/path/to/public.pem",
        "/path/to/master.key",
        "/path/to/rotation.key",
    )
    if err != nil {
        log.Fatal(err)
    }

    // Cria o middleware para descriptografar campos 'password' e 'token'
    decryptMiddleware := crypt.NewDecryptionMiddleware(
        &cryptService,
        []string{"password", "token"}, // Campos a serem descriptografados
        "aes",                          // Tipo de descriptografia
    )

    // Handler que recebe dados descriptografados
    loginHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var loginData struct {
            Username string `json:"username"`
            Password string `json:"password"` // Este campo será descriptografado automaticamente
        }
        
        if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
            http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
            return
        }
        
        // Agora loginData.Password contém a senha descriptografada
        log.Printf("Login attempt for user: %s", loginData.Username)
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Login processado"))
    })

    // Aplica o middleware ao handler
    http.Handle("/login", decryptMiddleware.Middleware(loginHandler))
    
    log.Println("Servidor iniciado na porta 8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Configuração via Struct

```go
config := crypt.DecryptionConfig{
    EncryptedFields:    []string{"password", "secret", "token"},
    DecryptionType:     "aes",
    RSAPrivateKeyPath:  "/path/to/private.pem",
    RSAPublicKeyPath:   "/path/to/public.pem",
    AESMasterKeyPath:   "/path/to/master.key",
    AESRotationKeyPath: "/path/to/rotation.key",
}

middleware, err := crypt.NewDecryptionMiddlewareFromConfig(config)
if err != nil {
    log.Fatal(err)
}

// Use o middleware...
http.Handle("/api/secure", middleware.Middleware(yourHandler))
```

### Encadeamento de Middlewares

```go
// Múltiplos middlewares podem ser encadeados
authMiddleware := NewAuthMiddleware()
decryptMiddleware := crypt.NewDecryptionMiddleware(&cryptService, []string{"password"}, "aes")
loggingMiddleware := NewLoggingMiddleware()

// Aplica os middlewares em ordem
handler := loggingMiddleware.Middleware(
    authMiddleware.Middleware(
        decryptMiddleware.Middleware(yourFinalHandler),
    ),
)

http.Handle("/api/endpoint", handler)
```

## Tipos de Descriptografia

### AES ("aes")
Usa apenas criptografia AES com a chave master. Mais rápido para dados pequenos.

### Híbrida ("hybrid")
Usa criptografia híbrida RSA+AES. Recomendado para dados maiores ou quando maior segurança é necessária.

## Comportamento do Middleware

- **Métodos suportados**: POST, PUT, PATCH
- **Content-Type**: Apenas `application/json`
- **Campos não encontrados**: Ignorados silenciosamente
- **Erros de descriptografia**: Retorna HTTP 500
- **Requisições não-JSON**: Passam sem modificação

## Exemplo de Requisição

```bash
# Antes da descriptografia (dados enviados pelo cliente)
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "user123",
    "password": "encrypted_password_here",
    "email": "user@example.com"
  }'

# Após o middleware, o handler recebe:
# {
#   "username": "user123",
#   "password": "decrypted_password_here",
#   "email": "user@example.com"
# }
```

## Geração de Chaves

### Chaves RSA

```bash
# Gera chave privada RSA (2048 bits)
openssl genrsa -out private.pem 2048

# Extrai chave pública
openssl rsa -in private.pem -pubout -out public.pem
```

### Chaves AES

```bash
# Gera chave AES de 256 bits (64 caracteres hex)
openssl rand -hex 32 > master.key
openssl rand -hex 32 > rotation.key
```

## Testes

```bash
# Executa todos os testes
go test -v ./...

# Executa testes com cobertura
go test -v -cover ./...
```

## Segurança

- **Proteção de chaves**: Mantenha os arquivos de chave com permissões restritivas (600)
- **HTTPS**: Use sempre HTTPS em produção
- **Rotação de chaves**: Implemente rotação regular das chaves AES
- **Logs**: Evite logar dados descriptografados

## Estrutura do Projeto

```
crypt/
├── crypt.go              # Funções de criptografia core
├── crypt_service.go      # Serviço de criptografia
├── middleware.go         # Middleware HTTP
├── middleware_example.go # Exemplos de uso
├── middleware_test.go    # Testes do middleware
├── go.mod               # Dependências
└── README.md            # Esta documentação
```

## Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.