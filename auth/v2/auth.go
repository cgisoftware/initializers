// Package auth fornece autenticação JWT e Basic Auth para APIs HTTP em Go.
//
// # Uso básico
//
//	a := auth.New("minha-chave-secreta")
//
//	// Gerar token
//	token, err := a.Sign(MyClaims{UserID: 1, Role: "admin"}, 24*time.Hour)
//
//	// Proteger rotas
//	r.Use(a.Middleware("user_id", "role"))
//
//	// Extrair do contexto no handler
//	userID, ok := auth.GetFromContext[int64](r.Context(), "user_id")
//
// # Definindo claims
//
// Implemente a interface [CustomClaims] para definir quais dados serão armazenados no token:
//
//	type UserClaims struct {
//	    UserID int64
//	    Role   string
//	}
//
//	func (c UserClaims) GetFields() map[auth.ContextValue]any {
//	    return map[auth.ContextValue]any{
//	        "user_id": c.UserID,
//	        "role":    c.Role,
//	    }
//	}
//
// # Opções de configuração
//
// Use as funções [WithCookieName], [WithBasicAuthValidator] e [WithCryptService]
// para configurar o comportamento do [Authenticator]:
//
//	a := auth.New("secret",
//	    auth.WithCookieName("SESSION"),
//	    auth.WithBasicAuthValidator(func(id, secret string) bool {
//	        return db.ValidateClient(id, secret)
//	    }),
//	)
//
// # Segurança
//
//   - Tokens são assinados com HMAC-SHA256. Algoritmos diferentes são rejeitados.
//   - Basic Auth é desabilitado por padrão; só funciona com [WithBasicAuthValidator].
//   - Tokens sem ExpiresAt são rejeitados.
//   - Cookie só é lido se [WithCookieName] for configurado.
package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ContextValue é o tipo das chaves usadas para armazenar e recuperar valores do contexto HTTP.
// Use constantes tipadas para evitar colisões:
//
//	const KeyUserID auth.ContextValue = "user_id"
type ContextValue string

// CustomClaims define os dados que serão embutidos no token JWT.
// Implemente esta interface na sua struct de claims:
//
//	type UserClaims struct{ UserID int64; Role string }
//
//	func (c UserClaims) GetFields() map[auth.ContextValue]any {
//	    return map[auth.ContextValue]any{
//	        "user_id": c.UserID,
//	        "role":    c.Role,
//	    }
//	}
type CustomClaims interface {
	GetFields() map[ContextValue]any
}

// CryptService é implementado pelo consumidor para descriptografar valores
// armazenados nos claims antes de injetá-los no contexto.
// Configurado via [WithCryptService].
//
// O [Authenticator] tenta primeiro [DecryptWithMasterKeySimple] (AES) e,
// em caso de falha, tenta [DecryptData] (híbrido). Se ambos falharem,
// o valor original é usado sem erro.
type CryptService interface {
	DecryptWithMasterKeySimple(encryptedData string) ([]byte, error)
	DecryptData(encryptedData string) ([]byte, error)
}

// Option é uma função de configuração aplicada ao [Authenticator] em [New].
type Option func(*Authenticator)

// WithCookieName configura o nome do cookie de onde o token será lido.
// O cookie tem precedência sobre o header Authorization quando presente.
//
// Se esta opção não for fornecida, apenas o header Authorization é utilizado.
func WithCookieName(name string) Option {
	return func(a *Authenticator) {
		a.cookieName = name
	}
}

// WithBasicAuthValidator configura a função de validação para autenticação Basic Auth.
// A função recebe clientID e secret e deve retornar true se as credenciais forem válidas.
//
// Se esta opção não for fornecida, requisições com Basic Auth são rejeitadas com 401.
//
//	auth.WithBasicAuthValidator(func(clientID, secret string) bool {
//	    return db.ValidateClient(clientID, secret)
//	})
func WithBasicAuthValidator(fn func(clientID, secret string) bool) Option {
	return func(a *Authenticator) {
		a.basicAuthValidator = fn
	}
}

// WithCryptService configura o serviço de descriptografia dos valores do contexto.
// Útil quando os claims no token carregam dados sensíveis criptografados.
//
// Ver [CryptService] para detalhes sobre a ordem de tentativas de descriptografia.
func WithCryptService(svc CryptService) Option {
	return func(a *Authenticator) {
		a.cryptService = svc
	}
}

// Authenticator gerencia a autenticação JWT e Basic Auth.
// Crie uma instância com [New].
type Authenticator struct {
	secretKey          []byte
	cookieName         string
	basicAuthValidator func(clientID, secret string) bool
	cryptService       CryptService
}

// internalClaims encapsula os dados do sistema e adiciona jwt.RegisteredClaims
type internalClaims struct {
	Data map[ContextValue]any `json:"data"`
	jwt.RegisteredClaims
}

// New cria um [Authenticator] com a chave secreta e as opções fornecidas.
//
//	a := auth.New("minha-chave-secreta",
//	    auth.WithCookieName("SESSION"),
//	    auth.WithBasicAuthValidator(validateFn),
//	)
func New(secretKey string, opts ...Option) *Authenticator {
	a := &Authenticator{
		secretKey: []byte(secretKey),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Sign gera e assina um token JWT com os claims fornecidos e o tempo de expiração.
// O token é assinado com HMAC-SHA256 usando a chave configurada em [New].
//
//	token, err := a.Sign(UserClaims{UserID: 1, Role: "admin"}, 24*time.Hour)
func (a *Authenticator) Sign(claims CustomClaims, expireIn time.Duration) (string, error) {
	internal := internalClaims{
		Data: claims.GetFields(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, internal)
	tokenString, err := token.SignedString(a.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Middleware retorna um middleware HTTP que autentica a requisição e injeta
// os values especificados no contexto para uso nos handlers.
//
// O token é lido do header Authorization (Bearer ou Basic) ou do cookie
// configurado em [WithCookieName]. Requisições sem token válido recebem 401.
//
// Os values injetados no contexto podem ser recuperados com [GetFromContext]:
//
//	r.Use(a.Middleware("user_id", "role"))
//
//	// no handler:
//	userID, _ := auth.GetFromContext[int64](r.Context(), "user_id")
func (a *Authenticator) Middleware(values ...ContextValue) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			headerToken := request.Header.Get("Authorization")

			if a.cookieName != "" {
				for _, cookie := range request.Cookies() {
					if cookie.Name == a.cookieName {
						headerToken = cookie.Value
						break
					}
				}
			}

			claims, isValid := a.verifyToken(headerToken)
			if !isValid {
				response.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := request.Context()
			for _, value := range values {
				field, exists := claims.Data[value]
				if !exists {
					continue
				}

				if a.cryptService != nil {
					if fieldStr, ok := field.(string); ok && fieldStr != "" {
						if decrypted, err := a.cryptService.DecryptWithMasterKeySimple(fieldStr); err == nil {
							ctx = context.WithValue(ctx, value, decrypted)
							continue
						}
						if decrypted, err := a.cryptService.DecryptData(fieldStr); err == nil {
							ctx = context.WithValue(ctx, value, decrypted)
							continue
						}
					}
				}

				ctx = context.WithValue(ctx, value, field)
			}

			next.ServeHTTP(response, request.WithContext(ctx))
		})
	}
}

// GetFromContext recupera um valor tipado do contexto injetado pelo [Authenticator.Middleware].
// Retorna o valor e true se encontrado e do tipo correto, ou o zero value e false caso contrário.
//
//	userID, ok := auth.GetFromContext[int64](ctx, "user_id")
//	role, ok   := auth.GetFromContext[string](ctx, "role")
//	isAdmin, ok := auth.GetFromContext[bool](ctx, "is_admin")
func GetFromContext[T any](ctx context.Context, key ContextValue) (T, bool) {
	v, ok := ctx.Value(key).(T)
	return v, ok
}

// verifyToken verifica e retorna as claims do token
func (a *Authenticator) verifyToken(bearerToken string) (*internalClaims, bool) {
	tokenType, headerToken := extractToken(bearerToken)

	if headerToken == "" {
		return nil, false
	}

	if tokenType == "Basic" {
		return a.verifyBasicToken(headerToken)
	}

	return a.verifyJWTToken(headerToken)
}

func (a *Authenticator) verifyBasicToken(encoded string) (*internalClaims, bool) {
	if a.basicAuthValidator == nil {
		return nil, false
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, false
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, false
	}

	if !a.basicAuthValidator(parts[0], parts[1]) {
		return nil, false
	}

	claims := &internalClaims{Data: map[ContextValue]any{
		"client_id": parts[0],
		"secret":    parts[1],
	}}
	return claims, true
}

func (a *Authenticator) verifyJWTToken(tokenString string) (*internalClaims, bool) {
	claims := &internalClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo de assinatura inesperado: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil || !token.Valid || claims.ExpiresAt == nil {
		return nil, false
	}

	return claims, true
}

// extractToken separa o tipo do token do valor
func extractToken(bearerToken string) (string, string) {
	parts := strings.Split(bearerToken, " ")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", bearerToken
}
