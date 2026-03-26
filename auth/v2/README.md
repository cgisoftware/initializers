# auth/v2

Pacote de autenticação JWT e Basic Auth para APIs HTTP em Go.

## Instalação

```bash
go get github.com/cgisoftware/initializers/auth/v2
```

## Início rápido

```go
import auth "github.com/cgisoftware/initializers/auth/v2"

// 1. Defina suas claims
type UserClaims struct {
    UserID int64
    Role   string
}

func (c UserClaims) GetFields() map[auth.ContextValue]any {
    return map[auth.ContextValue]any{
        "user_id": c.UserID,
        "role":    c.Role,
    }
}

// 2. Inicialize
a := auth.New("minha-chave-secreta")

// 3. Gere um token
token, err := a.Sign(UserClaims{UserID: 1, Role: "admin"}, 24*time.Hour)

// 4. Proteja as rotas
r.Use(a.Middleware("user_id", "role"))

// 5. Acesse os dados no handler
userID, _ := auth.GetFromContext[int64](r.Context(), "user_id")
role, _   := auth.GetFromContext[string](r.Context(), "role")
```

---

## Configuração

`auth.New` aceita opções funcionais para configurar o comportamento do autenticador.

### `WithCookieName`

Lê o token de um cookie além do header `Authorization`. O cookie tem precedência quando presente.

```go
a := auth.New("secret",
    auth.WithCookieName("SESSION"),
)
```

> Se não configurado, apenas o header `Authorization` é lido.

### `WithBasicAuthValidator`

Habilita e valida autenticação Basic Auth. A função recebe `clientID` e `secret` e deve retornar `true` se as credenciais forem válidas.

```go
a := auth.New("secret",
    auth.WithBasicAuthValidator(func(clientID, secret string) bool {
        return db.ValidateClient(clientID, secret)
    }),
)
```

> Se não configurado, qualquer requisição com `Authorization: Basic ...` recebe **401**.

Após validação, `client_id` e `secret` ficam disponíveis no contexto:

```go
clientID, _ := auth.GetFromContext[string](r.Context(), "client_id")
```

### `WithCryptService`

Descriptografa automaticamente os valores dos claims antes de injetá-los no contexto. Útil quando o token carrega dados sensíveis criptografados.

```go
a := auth.New("secret",
    auth.WithCryptService(myCryptService),
)
```

Implemente a interface `CryptService`:

```go
type CryptService interface {
    DecryptWithMasterKeySimple(encryptedData string) ([]byte, error)
    DecryptData(encryptedData string) ([]byte, error)
}
```

A descriptografia tenta `DecryptWithMasterKeySimple` (AES) primeiro, depois `DecryptData` (híbrido). Se ambos falharem, o valor original é usado.

---

## API

### `New`

```go
func New(secretKey string, opts ...Option) *Authenticator
```

Cria um `Authenticator`. A `secretKey` é usada para assinar e verificar todos os tokens.

### `Sign`

```go
func (a *Authenticator) Sign(claims CustomClaims, expireIn time.Duration) (string, error)
```

Gera um token JWT assinado com HMAC-SHA256. O token expira após `expireIn`.

```go
token, err := a.Sign(UserClaims{UserID: 42, Role: "admin"}, 8*time.Hour)
```

### `Middleware`

```go
func (a *Authenticator) Middleware(values ...ContextValue) func(http.Handler) http.Handler
```

Middleware HTTP que autentica a requisição e injeta os `values` especificados no contexto. Retorna **401** se o token for inválido ou ausente.

```go
// injeta apenas os campos necessários por rota
r.Use(a.Middleware("user_id", "role"))
```

### `GetFromContext`

```go
func GetFromContext[T any](ctx context.Context, key ContextValue) (T, bool)
```

Recupera um valor tipado do contexto. Retorna o zero value e `false` se não encontrado ou tipo incorreto.

```go
userID, ok  := auth.GetFromContext[int64](ctx, "user_id")
role, ok    := auth.GetFromContext[string](ctx, "role")
isAdmin, ok := auth.GetFromContext[bool](ctx, "is_admin")
```

---

## Como o token é lido

O middleware verifica as seguintes fontes, nesta ordem:

1. Header `Authorization: Bearer <token>`
2. Header `Authorization: Basic <base64>`
3. Cookie configurado com `WithCookieName` (se presente, sobrescreve o header)

---

## Segurança

| Proteção | Comportamento |
|----------|---------------|
| Algoritmo | Apenas HMAC (HS256/HS384/HS512) é aceito. Outros algoritmos resultam em 401. |
| Expiração | Tokens sem `ExpiresAt` ou expirados são rejeitados. |
| Basic Auth | Desabilitado por padrão. Requer `WithBasicAuthValidator` para funcionar. |
| Cookie | Não lido por padrão. Requer `WithCookieName` para habilitar. |

---

## Exemplo completo com Chi

```go
package main

import (
    "fmt"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    auth "github.com/cgisoftware/initializers/auth/v2"
)

type UserClaims struct {
    UserID int64
    Role   string
}

func (c UserClaims) GetFields() map[auth.ContextValue]any {
    return map[auth.ContextValue]any{
        "user_id": c.UserID,
        "role":    c.Role,
    }
}

func main() {
    a := auth.New(
        "minha-chave-super-secreta",
        auth.WithCookieName("SESSION"),
    )

    token, _ := a.Sign(UserClaims{UserID: 1, Role: "admin"}, 24*time.Hour)
    fmt.Println("Token:", token)

    r := chi.NewRouter()

    r.Group(func(r chi.Router) {
        r.Use(a.Middleware("user_id", "role"))

        r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
            userID, _ := auth.GetFromContext[int64](r.Context(), "user_id")
            role, _   := auth.GetFromContext[string](r.Context(), "role")
            fmt.Fprintf(w, "user=%d role=%s", userID, role)
        })
    })

    http.ListenAndServe(":8080", r)
}
```
