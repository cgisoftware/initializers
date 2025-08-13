package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ContextValue representa uma chave usada no contexto HTTP
type ContextValue string

// CustomClaims define a interface para que o sistema forneça seus próprios dados
type CustomClaims interface {
	GetFields() map[ContextValue]any // Retorna os campos como um mapa
}

// internalClaims encapsula os dados do sistema e adiciona jwt.RegisteredClaims
type internalClaims struct {
	Data map[ContextValue]any `json:"data"`
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

// CryptService interface para descriptografia (evita dependência circular)
type CryptService interface {
	DecryptWithMasterKeySimple(encryptedData string) ([]byte, error)
	DecryptData(encryptedData string) ([]byte, error)
}

// AuthMiddleware cria um middleware que verifica a autenticação
// Se cryptService for fornecido, tentará descriptografar os valores antes de adicioná-los ao contexto
func (a *Authenticator) AuthMiddleware(values ...ContextValue) func(next http.Handler) http.Handler {
	return a.AuthMiddlewareWithCrypt(nil, values...)
}

// AuthMiddlewareWithCrypt cria um middleware que verifica a autenticação e opcionalmente descriptografa valores
// Se cryptService for fornecido, tentará descriptografar os valores antes de adicioná-los ao contexto
func (a *Authenticator) AuthMiddlewareWithCrypt(cryptService CryptService, values ...ContextValue) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			headerToken := request.Header.Get("Authorization")

			cookies := request.Cookies()
			for _, cookie := range cookies {
				if cookie.Name == "CIMSSESSIONTOKEN" {
					headerToken = cookie.Value
					break
				}
			}

			if claims, isValid := a.verifyToken(headerToken); isValid {
				ctx := request.Context()
				fields := claims.Data

				// Itera sobre os valores passados e adiciona no contexto
				for _, value := range values {
					if field, exists := fields[value]; exists {
						// Tenta descriptografar o valor se cryptService foi fornecido
						if cryptService != nil {
							if fieldStr, ok := field.(string); ok && fieldStr != "" {
								// Tenta descriptografar usando AES primeiro
								if decrypted, err := cryptService.DecryptWithMasterKeySimple(fieldStr); err == nil {
									ctx = context.WithValue(ctx, value, decrypted)
									continue
								}
								// Se falhar, tenta descriptografia híbrida
								if decrypted, err := cryptService.DecryptData(fieldStr); err == nil {
									ctx = context.WithValue(ctx, value, decrypted)
									continue
								}
							}
						}
						// Se não conseguir descriptografar ou cryptService não foi fornecido, usa o valor original
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
	tokenType, headerToken := extractToken(bearerToken)

	if headerToken == "" {
		return nil, false
	}

	if tokenType == "Basic" {
		decoded, err := base64.StdEncoding.DecodeString(headerToken)
		if err != nil {
			return nil, false
		}

		parts := strings.Split(string(decoded), ":")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, false
		}

		claims.Data["client_id"] = parts[0]
		claims.Data["secret"] = parts[1]
		return claims, true
	}

	token, err := jwt.ParseWithClaims(headerToken, claims, func(token *jwt.Token) (any, error) {
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
func extractToken(bearerToken string) (string, string) {
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[0], strArr[1]
	}
	return "", bearerToken
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
	v := ctx.Value(value)
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		if float64(int64(val)) == val {
			return int64(val)
		}
		return 0
	default:
		return 0
	}
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
func GetFloat64FromContext(ctx context.Context, value ContextValue) float64 {
	v, ok := ctx.Value(value).(float64)
	if !ok {
		return 0.0
	}
	return v
}

// GetInterfaceFromContext busca um valor genérico (any) do contexto
func GetInterfaceFromContext(ctx context.Context, value ContextValue) any {
	return ctx.Value(value)
}
