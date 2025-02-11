package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// ContextValue representa uma chave usada no contexto HTTP
type ContextValue string

// CustomClaims define a interface para que o sistema forneça seus próprios dados
type CustomClaims interface {
	GetFields() map[ContextValue]interface{} // Retorna os campos como um mapa
}

// internalClaims encapsula os dados do sistema e adiciona jwt.RegisteredClaims
type internalClaims struct {
	Data map[ContextValue]interface{} `json:"data"`
	jwt.RegisteredClaims
}

// Authenticator gerencia a autenticação
type Authenticator struct {
	secretKey []byte
	claims    CustomClaims
}

// NewAuthenticator cria uma instância de autenticação com claims personalizadas
func Initialize(secretKey string, claims CustomClaims) *Authenticator {
	return &Authenticator{
		secretKey: []byte(secretKey),
		claims:    claims,
	}
}

// GetSignToken cria um token com base na chave
func GetSignToken(claims CustomClaims, expireIn time.Duration, secretKey string) string {
	internal := internalClaims{
		Data: claims.GetFields(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, internal)
	tokenString, _ := token.SignedString([]byte(secretKey))

	return tokenString
}

// AuthMiddleware cria um middleware que verifica a autenticação
func (a *Authenticator) AuthMiddleware(values ...ContextValue) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			bearToken := request.Header.Get("Authorization")

			if claims, isValid := a.verifyToken(bearToken); isValid {
				ctx := request.Context()
				fields := claims.Data

				// Itera sobre os valores passados e adiciona no contexto
				for _, value := range values {
					if field, exists := fields[value]; exists {
						ctx = context.WithValue(ctx, value, field)
					}
				}

				request = request.WithContext(ctx)
				next.ServeHTTP(response, request)
				return
			}

			response.WriteHeader(http.StatusUnauthorized)
		})
	}
}

// verifyToken verifica o token JWT usando as claims encapsuladas dentro do pacote
func (a *Authenticator) verifyToken(bearerToken string) (*internalClaims, bool) {
	claims := &internalClaims{Data: a.claims.GetFields()}

	token, err := jwt.ParseWithClaims(extractToken(bearerToken), claims, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})

	if err != nil {
		return nil, false
	}

	if !token.Valid || claims.ExpiresAt == nil {
		return nil, false
	}

	return claims, true
}

// extractToken extrai o token do cabeçalho Authorization
func extractToken(bearerToken string) string {
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// GetStringFromContext busca um valor string do contexto
func GetStringFromContext(ctx context.Context, value ContextValue) string {
	v, ok := ctx.Value(value).(string)
	if !ok {
		return ""
	}
	return v
}

// GetIntFromContext busca um valor int do contexto
func GetInt64FromContext(ctx context.Context, value ContextValue) int64 {
	v, ok := ctx.Value(value).(int64)
	if !ok {
		return 0
	}
	return v
}

// GetBoolFromContext busca um valor bool do contexto
func GetBoolFromContext(ctx context.Context, value ContextValue) bool {
	v, ok := ctx.Value(value).(bool)
	if !ok {
		return false
	}
	return v
}

// GetFloatFromContext busca um valor float64 do contexto
func GetFloatFromContext(ctx context.Context, value ContextValue) float64 {
	v, ok := ctx.Value(value).(float64)
	if !ok {
		return 0.0
	}
	return v
}

// GetInterfaceFromContext busca um valor genérico (interface{}) do contexto
func GetInterfaceFromContext(ctx context.Context, value ContextValue) interface{} {
	return ctx.Value(value)
}
