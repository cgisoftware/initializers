# Pacote Auth

O pacote `auth` fornece funcionalidades completas de autenticação e autorização usando JWT (JSON Web Tokens) com suporte a criptografia opcional.

## Funcionalidades

### 🔐 Autenticação JWT
- Geração e validação de tokens JWT
- Claims personalizáveis
- Suporte a diferentes algoritmos de assinatura
- Extração automática de tokens de headers HTTP

### 🛡️ Middleware de Autenticação
- Middleware padrão para validação de tokens
- Middleware com criptografia para dados sensíveis
- Integração com contexto HTTP
- Tratamento de erros padronizado

### 🔑 Gerenciamento de Contexto
- Extração de dados do usuário do contexto
- Acesso a claims personalizadas
- Informações de autenticação estruturadas

## Estruturas Principais

### `Authenticator`
```go
type Authenticator struct {
    secretKey   []byte
    cryptService CryptService // Opcional
}
```

Estrutura principal que gerencia todas as operações de autenticação.

### `CustomClaims`
```go
type CustomClaims interface {
    GetFields() map[string]interface{}
}
```

Interface para implementar claims personalizadas nos tokens JWT.

### `ContextValue`
```go
type ContextValue struct {
    UserID       string
    CustomerID   string
    Email        string
    Name         string
    CustomClaims CustomClaims
}
```

Estrutura que armazena informações do usuário no contexto HTTP.

## Configuração

### Inicialização Básica
```go
// Sem criptografia
auth := auth.Initialize("minha-chave-secreta", nil)

// Com criptografia
cryptService := crypt.Initialize()
auth := auth.Initialize("minha-chave-secreta", cryptService)
```

### Geração de Tokens
```go
// Claims básicas
claims := map[string]interface{}{
    "user_id": "123",
    "email": "user@exemplo.com",
    "role": "admin",
}

token, err := auth.GetSignToken(claims, time.Hour*24) // Expira em 24h
if err != nil {
    log.Fatal(err)
}
```

### Claims Personalizadas
```go
type MeusClaims struct {
    UserID     string `json:"user_id"`
    CustomerID string `json:"customer_id"`
    Role       string `json:"role"`
}

func (c MeusClaims) GetFields() map[string]interface{} {
    return map[string]interface{}{
        "user_id":     c.UserID,
        "customer_id": c.CustomerID,
        "role":        c.Role,
    }
}

// Uso
claims := MeusClaims{
    UserID:     "123",
    CustomerID: "456",
    Role:       "admin",
}

token, err := auth.GetSignToken(claims.GetFields(), time.Hour*24)
```

## Middleware

### Middleware Básico
```go
func (r *gin.Engine) setupRoutes(auth *auth.Authenticator) {
    // Rotas públicas
    r.POST("/login", loginHandler)
    
    // Rotas protegidas
    protected := r.Group("/api")
    protected.Use(auth.AuthMiddleware())
    {
        protected.GET("/profile", profileHandler)
        protected.POST("/data", dataHandler)
    }
}
```

### Middleware com Criptografia
```go
// Para dados sensíveis que precisam ser criptografados
protected.Use(auth.AuthMiddlewareWithCrypt())
```

## Extração de Dados do Contexto

### Informações do Usuário
```go
func profileHandler(c *gin.Context) {
    // Extrair ID do usuário
    userID := auth.GetUserIDFromContext(c)
    if userID == "" {
        c.JSON(401, gin.H{"error": "Usuário não autenticado"})
        return
    }
    
    // Extrair email
    email := auth.GetEmailFromContext(c)
    
    // Extrair nome
    name := auth.GetNameFromContext(c)
    
    // Extrair customer ID
    customerID := auth.GetCustomerIDFromContext(c)
    
    // Extrair claims personalizadas
    customClaims := auth.GetCustomClaimsFromContext(c)
    
    c.JSON(200, gin.H{
        "user_id":      userID,
        "email":        email,
        "name":         name,
        "customer_id":  customerID,
        "custom_claims": customClaims,
    })
}
```

### Contexto Completo
```go
func dataHandler(c *gin.Context) {
    contextValue := auth.GetContextValueFromContext(c)
    if contextValue == nil {
        c.JSON(401, gin.H{"error": "Contexto inválido"})
        return
    }
    
    // Usar dados do contexto
    fmt.Printf("Usuário: %s (%s)\n", contextValue.Name, contextValue.Email)
    fmt.Printf("Customer: %s\n", contextValue.CustomerID)
    
    if contextValue.CustomClaims != nil {
        fields := contextValue.CustomClaims.GetFields()
        fmt.Printf("Claims personalizadas: %+v\n", fields)
    }
}
```

## Exemplos de Uso

### Sistema de Login Completo
```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "seu-projeto/initializers/auth"
)

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type UserClaims struct {
    UserID     string `json:"user_id"`
    CustomerID string `json:"customer_id"`
    Role       string `json:"role"`
    Permissions []string `json:"permissions"`
}

func (c UserClaims) GetFields() map[string]interface{} {
    return map[string]interface{}{
        "user_id":     c.UserID,
        "customer_id": c.CustomerID,
        "role":        c.Role,
        "permissions": c.Permissions,
    }
}

func main() {
    // Inicializar autenticador
    authenticator := auth.Initialize("minha-chave-super-secreta", nil)
    
    r := gin.Default()
    
    // Login endpoint
    r.POST("/login", func(c *gin.Context) {
        var req LoginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        // Validar credenciais (implementar sua lógica)
        user, err := validateCredentials(req.Email, req.Password)
        if err != nil {
            c.JSON(401, gin.H{"error": "Credenciais inválidas"})
            return
        }
        
        // Criar claims
        claims := UserClaims{
            UserID:     user.ID,
            CustomerID: user.CustomerID,
            Role:       user.Role,
            Permissions: user.Permissions,
        }
        
        // Gerar token
        token, err := authenticator.GetSignToken(claims.GetFields(), time.Hour*24)
        if err != nil {
            c.JSON(500, gin.H{"error": "Erro ao gerar token"})
            return
        }
        
        c.JSON(200, gin.H{
            "token": token,
            "user": gin.H{
                "id":    user.ID,
                "email": user.Email,
                "name":  user.Name,
                "role":  user.Role,
            },
        })
    })
    
    // Rotas protegidas
    api := r.Group("/api")
    api.Use(authenticator.AuthMiddleware())
    {
        api.GET("/profile", getProfile)
        api.PUT("/profile", updateProfile)
        api.GET("/dashboard", getDashboard)
    }
    
    r.Run(":8080")
}
```

### Middleware Personalizado
```go
// Middleware que verifica permissões específicas
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        contextValue := auth.GetContextValueFromContext(c)
        if contextValue == nil {
            c.JSON(401, gin.H{"error": "Não autenticado"})
            c.Abort()
            return
        }
        
        if contextValue.CustomClaims != nil {
            fields := contextValue.CustomClaims.GetFields()
            if permissions, ok := fields["permissions"].([]string); ok {
                for _, p := range permissions {
                    if p == permission {
                        c.Next()
                        return
                    }
                }
            }
        }
        
        c.JSON(403, gin.H{"error": "Permissão insuficiente"})
        c.Abort()
    }
}

// Uso
api.POST("/admin/users", RequirePermission("create_users"), createUserHandler)
api.DELETE("/admin/users/:id", RequirePermission("delete_users"), deleteUserHandler)
```

## Segurança

### Boas Práticas

1. **Chave Secreta Forte**
   ```go
   // ❌ Não faça isso
   auth := auth.Initialize("123456", nil)
   
   // ✅ Use uma chave forte
   secretKey := os.Getenv("JWT_SECRET_KEY") // Pelo menos 32 caracteres
   auth := auth.Initialize(secretKey, nil)
   ```

2. **Tempo de Expiração Adequado**
   ```go
   // Para APIs web
   token, _ := auth.GetSignToken(claims, time.Hour*2) // 2 horas
   
   // Para mobile apps
   token, _ := auth.GetSignToken(claims, time.Hour*24*7) // 1 semana
   
   // Para refresh tokens
   refreshToken, _ := auth.GetSignToken(claims, time.Hour*24*30) // 30 dias
   ```

3. **Validação de Claims**
   ```go
   func validateUserClaims(claims map[string]interface{}) error {
       userID, ok := claims["user_id"].(string)
       if !ok || userID == "" {
           return errors.New("user_id inválido")
       }
       
       // Verificar se usuário ainda existe e está ativo
       if !isUserActive(userID) {
           return errors.New("usuário inativo")
       }
       
       return nil
   }
   ```

4. **Headers de Segurança**
   ```go
   func securityHeaders() gin.HandlerFunc {
       return func(c *gin.Context) {
           c.Header("X-Content-Type-Options", "nosniff")
           c.Header("X-Frame-Options", "DENY")
           c.Header("X-XSS-Protection", "1; mode=block")
           c.Next()
       }
   }
   
   r.Use(securityHeaders())
   ```

### Integração com Criptografia

Quando usado com o pacote `crypt`, dados sensíveis podem ser criptografados:

```go
cryptService := crypt.Initialize()
auth := auth.Initialize(secretKey, cryptService)

// Claims sensíveis serão automaticamente criptografadas
sensitiveClaims := map[string]interface{}{
    "user_id": "123",
    "ssn": "123-45-6789", // Será criptografado
    "credit_card": "4111-1111-1111-1111", // Será criptografado
}

token, err := auth.GetSignToken(sensitiveClaims, time.Hour*24)
```

## Tratamento de Erros

### Erros Comuns

1. **Token Inválido**
   ```go
   // O middleware automaticamente retorna 401 para tokens inválidos
   // Você pode personalizar a resposta:
   
   func customAuthMiddleware(auth *auth.Authenticator) gin.HandlerFunc {
       return func(c *gin.Context) {
           token := auth.extractToken(c.Request)
           if token == "" {
               c.JSON(401, gin.H{
                   "error": "Token não fornecido",
                   "code": "MISSING_TOKEN",
               })
               c.Abort()
               return
           }
           
           // Continuar com validação...
       }
   }
   ```

2. **Token Expirado**
   ```go
   // Implementar refresh token
   func refreshTokenHandler(c *gin.Context) {
       refreshToken := c.GetHeader("X-Refresh-Token")
       
       // Validar refresh token
       claims, err := auth.verifyToken(refreshToken)
       if err != nil {
           c.JSON(401, gin.H{"error": "Refresh token inválido"})
           return
       }
       
       // Gerar novo access token
       newToken, err := auth.GetSignToken(claims, time.Hour*2)
       if err != nil {
           c.JSON(500, gin.H{"error": "Erro ao gerar novo token"})
           return
       }
       
       c.JSON(200, gin.H{"token": newToken})
   }
   ```

## Testes

### Testando Autenticação

```go
package auth_test

import (
    "testing"
    "time"
    "seu-projeto/initializers/auth"
)

func TestTokenGeneration(t *testing.T) {
    authenticator := auth.Initialize("test-secret-key", nil)
    
    claims := map[string]interface{}{
        "user_id": "123",
        "email": "test@example.com",
    }
    
    token, err := authenticator.GetSignToken(claims, time.Hour)
    if err != nil {
        t.Fatalf("Erro ao gerar token: %v", err)
    }
    
    if token == "" {
        t.Fatal("Token não deve estar vazio")
    }
    
    // Verificar se o token é válido
    parsedClaims, err := authenticator.verifyToken(token)
    if err != nil {
        t.Fatalf("Erro ao verificar token: %v", err)
    }
    
    if parsedClaims["user_id"] != "123" {
        t.Errorf("user_id esperado: 123, obtido: %v", parsedClaims["user_id"])
    }
}
```

## Dependências

- `github.com/golang-jwt/jwt/v5` - Para operações JWT
- `github.com/gin-gonic/gin` - Para middleware HTTP (opcional)
- Pacote `crypt` interno - Para criptografia (opcional)

## Veja Também

- [Pacote Crypt](../crypt/README.md) - Para criptografia de dados sensíveis
- [Pacote Validator](../validator/README.md) - Para validação de dados de entrada
- [Pacote Formatter](../formatter/README.md) - Para formatação de respostas HTTP

---

**Nota**: Este pacote foi projetado para ser flexível e seguro. Sempre revise as configurações de segurança antes de usar em produção.