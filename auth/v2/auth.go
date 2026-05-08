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
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type signingMethod int

const (
	signingMethodHMAC signingMethod = iota
	signingMethodRSA
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
// Crie uma instância com [New], [NewRSA] ou [NewRSAVerifier].
type Authenticator struct {
	secretKey          []byte
	privateKey         *rsa.PrivateKey
	publicKey          *rsa.PublicKey
	sigMethod          signingMethod
	cookieName         string
	basicAuthValidator func(clientID, secret string) bool
	cryptService       CryptService
}

// internalClaims encapsula os dados do sistema e adiciona jwt.RegisteredClaims
type internalClaims struct {
	Data map[ContextValue]any `json:"data"`
	jwt.RegisteredClaims
}

// New cria um [Authenticator] HMAC-SHA256 com a chave secreta e as opções fornecidas.
//
//	a := auth.New("minha-chave-secreta",
//	    auth.WithCookieName("SESSION"),
//	    auth.WithBasicAuthValidator(validateFn),
//	)
func New(secretKey string, opts ...Option) *Authenticator {
	a := &Authenticator{
		secretKey: []byte(secretKey),
		sigMethod: signingMethodHMAC,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// NewRSA cria um [Authenticator] RS256 para o emissor do token (assina + valida).
// privateKeyPEM e publicKeyPEM devem ser chaves PEM (PKCS8 ou PKCS1/PKIX).
//
//	a, err := auth.NewRSA(os.Getenv("RSA_PRIVATE_KEY"), os.Getenv("RSA_PUBLIC_KEY"))
func NewRSA(privateKeyPEM, publicKeyPEM string, opts ...Option) (*Authenticator, error) {
	priv, err := parseRSAPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("private key: %w", err)
	}
	pub, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("public key: %w", err)
	}
	a := &Authenticator{privateKey: priv, publicKey: pub, sigMethod: signingMethodRSA}
	for _, opt := range opts {
		opt(a)
	}
	return a, nil
}

// NewRSAVerifier cria um [Authenticator] RS256 para backends que só validam tokens.
// Não requer a private key — apenas a public key é necessária.
//
//	a, err := auth.NewRSAVerifier(os.Getenv("RSA_PUBLIC_KEY"))
func NewRSAVerifier(publicKeyPEM string, opts ...Option) (*Authenticator, error) {
	pub, err := parseRSAPublicKey(publicKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("public key: %w", err)
	}
	a := &Authenticator{publicKey: pub, sigMethod: signingMethodRSA}
	for _, opt := range opts {
		opt(a)
	}
	return a, nil
}

// Sign gera e assina um token JWT com os claims fornecidos e o tempo de expiração.
// Usa HMAC-SHA256 se criado com [New], ou RS256 se criado com [NewRSA].
//
//	token, err := a.Sign(UserClaims{UserID: 1, Role: "admin"}, 24*time.Hour)
func (a *Authenticator) Sign(claims CustomClaims, expireIn time.Duration) (string, error) {
	internal := internalClaims{
		Data: claims.GetFields(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireIn)),
		},
	}

	switch a.sigMethod {
	case signingMethodRSA:
		if a.privateKey == nil {
			return "", fmt.Errorf("private key não configurada: use NewRSA para assinar tokens")
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, internal)
		return token.SignedString(a.privateKey)
	default:
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, internal)
		return token.SignedString(a.secretKey)
	}
}

// JWKSHandler retorna um [http.Handler] que expõe a public key RSA no formato JWKS.
// Registre em /.well-known/jwks.json para que backends busquem a chave automaticamente.
// Retorna 404 se o Authenticator não foi criado com [NewRSA] ou [NewRSAVerifier].
//
//	r.Get("/.well-known/jwks.json", a.JWKSHandler().ServeHTTP)
func (a *Authenticator) JWKSHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.publicKey == nil || a.sigMethod != signingMethodRSA {
			http.Error(w, "not configured for RSA", http.StatusNotFound)
			return
		}
		n := base64.RawURLEncoding.EncodeToString(a.publicKey.N.Bytes())
		e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(a.publicKey.E)).Bytes())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"keys":[{"kty":"RSA","alg":"RS256","use":"sig","n":%q,"e":%q}]}`, n, e)
	})
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
		switch a.sigMethod {
		case signingMethodRSA:
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("algoritmo de assinatura inesperado: %v", token.Header["alg"])
			}
			return a.publicKey, nil
		default:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritmo de assinatura inesperado: %v", token.Header["alg"])
			}
			return a.secretKey, nil
		}
	})

	if err != nil || !token.Valid || claims.ExpiresAt == nil {
		return nil, false
	}

	return claims, true
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("falha ao decodificar bloco PEM")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
		return nil, fmt.Errorf("não é uma chave RSA")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("falha ao decodificar bloco PEM")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("não é uma chave RSA pública")
	}
	return rsaKey, nil
}

// extractToken separa o tipo do token do valor
func extractToken(bearerToken string) (string, string) {
	parts := strings.Split(bearerToken, " ")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", bearerToken
}
