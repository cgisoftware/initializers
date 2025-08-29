package formatter

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// ExampleBasicErrorHandling demonstra o uso básico do sistema de tratamento de erros
func ExampleBasicErrorHandling() {
	// Criar um ResponseWriter de teste
	w := httptest.NewRecorder()

	// Exemplo 1: Erro simples
	err := errors.New("algo deu errado")
	HttpErrorResponse(w, err)

	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	// Exemplo 2: Erro com mensagem customizada
	w2 := httptest.NewRecorder()
	HttpErrorResponse(w2, err, "Erro personalizado", "Detalhes adicionais")

	fmt.Printf("\nCom mensagem customizada:\n")
	fmt.Printf("Status Code: %d\n", w2.Code)
	fmt.Printf("Response Body: %s\n", w2.Body.String())
}

// ExamplePredefinedErrors demonstra o uso dos erros pré-definidos
func ExamplePredefinedErrors() {
	fmt.Println("=== ERROS PRÉ-DEFINIDOS ===")

	// Exemplo com cada tipo de erro
	errors := map[string]error{
		"Não Autorizado":     ErrAuth,
		"Não Encontrado":     ErrNotFound,
		"Duplicado":          ErrDuplicate,
		"Erro Interno":       ErrInternalServer,
		"Requisição Inválida": ErrBadRequest,
		"ID Não Encontrado":  ErrIDNotFound,
		"Token API Ausente":  ErrAPITokenKeyNotFound,
		"Código Menu Ausente": ErrCodMenuKeyNotFound,
	}

	for name, err := range errors {
		w := httptest.NewRecorder()
		HttpErrorResponse(w, err)
		fmt.Printf("%s: Status %d - %s\n", name, w.Code, w.Body.String())
	}
}

// ExampleCustomErrorAPI demonstra como criar erros customizados
func ExampleCustomErrorAPI() {
	// Criar erro customizado
	customErr := &errorAPIError{
		status: http.StatusTeapot, // 418 - I'm a teapot
		err:    errors.New("sou um bule de chá"),
	}

	w := httptest.NewRecorder()
	customErr.HttpErrorResponse(w)

	fmt.Printf("Erro customizado:\n")
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())
	fmt.Printf("Error Message: %s\n", customErr.Error())
}

// ExampleWrapError demonstra como envolver erros
func ExampleWrapError() {
	// Erro base
	baseErr := &errorAPIError{
		status: http.StatusBadRequest,
		err:    errors.New("erro original"),
	}

	// Envolver o erro com contexto adicional
	wrappedErr := WrapError(baseErr, "contexto adicional")

	w := httptest.NewRecorder()
	HttpErrorResponse(w, wrappedErr)

	fmt.Printf("Erro envolvido:\n")
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())
}

// ExampleHTTPHandlerWithErrors demonstra uso em handlers HTTP reais
func ExampleHTTPHandlerWithErrors() {
	// Handler que pode retornar diferentes tipos de erro
	userHandler := func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("id")
		
		// Validar ID
		if userID == "" {
			HttpErrorResponse(w, ErrIDNotFound)
			return
		}
		
		// Simular busca de usuário
		if userID == "999" {
			HttpErrorResponse(w, ErrNotFound)
			return
		}
		
		// Simular erro de autorização
		auth := r.Header.Get("Authorization")
		if auth == "" {
			HttpErrorResponse(w, ErrAuth)
			return
		}
		
		// Simular erro interno
		if userID == "500" {
			HttpErrorResponse(w, ErrInternalServer)
			return
		}
		
		// Sucesso
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": "%s", "name": "Usuário %s"}`, userID, userID)
	}

	// Testar diferentes cenários
	scenarios := []struct {
		name   string
		url    string
		header map[string]string
	}{
		{"ID ausente", "/user", nil},
		{"Usuário não encontrado", "/user?id=999", nil},
		{"Não autorizado", "/user?id=123", nil},
		{"Erro interno", "/user?id=500", map[string]string{"Authorization": "Bearer token"}},
		{"Sucesso", "/user?id=123", map[string]string{"Authorization": "Bearer token"}},
	}

	fmt.Println("=== TESTES DE HANDLER ===")
	for _, scenario := range scenarios {
		req := httptest.NewRequest("GET", scenario.url, nil)
		for key, value := range scenario.header {
			req.Header.Set(key, value)
		}
		
		w := httptest.NewRecorder()
		userHandler(w, req)
		
		fmt.Printf("%s: Status %d - %s\n", 
			scenario.name, w.Code, w.Body.String())
	}
}

// ExampleMiddlewareErrorHandling demonstra uso em middleware
func ExampleMiddlewareErrorHandling() {
	// Middleware de validação
	validationMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validar Content-Type para POST/PUT
			if r.Method == "POST" || r.Method == "PUT" {
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					HttpErrorResponse(w, ErrBadRequest, 
						"Content-Type deve ser application/json")
					return
				}
			}
			
			// Validar tamanho do body
			if r.ContentLength > 1024*1024 { // 1MB
				HttpErrorResponse(w, ErrBadRequest, 
					"Body muito grande (máximo 1MB)")
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}

	// Handler simples
	simpleHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	// Combinar middleware + handler
	handlerWithMiddleware := validationMiddleware(simpleHandler)

	// Testar cenários
	tests := []struct {
		name        string
		method      string
		contentType string
		contentLength int64
	}{
		{"GET válido", "GET", "", 0},
		{"POST sem Content-Type", "POST", "", 100},
		{"POST com Content-Type correto", "POST", "application/json", 100},
		{"POST com body muito grande", "POST", "application/json", 2*1024*1024},
	}

	fmt.Println("\n=== TESTES DE MIDDLEWARE ===")
	for _, test := range tests {
		req := httptest.NewRequest(test.method, "/", nil)
		if test.contentType != "" {
			req.Header.Set("Content-Type", test.contentType)
		}
		req.ContentLength = test.contentLength
		
		w := httptest.NewRecorder()
		handlerWithMiddleware.ServeHTTP(w, req)
		
		fmt.Printf("%s: Status %d - %s\n", 
			test.name, w.Code, w.Body.String())
	}
}

// ExampleErrorChaining demonstra encadeamento de erros
func ExampleErrorChaining() {
	// Simular uma cadeia de erros
	processOrder := func(orderID string) error {
		if orderID == "" {
			return ErrIDNotFound
		}
		
		// Simular erro de validação
		if orderID == "invalid" {
			return WrapError(ErrBadRequest, "ID do pedido inválido")
		}
		
		// Simular erro de autorização
		if orderID == "unauthorized" {
			return WrapError(ErrAuth, "usuário não tem permissão para este pedido")
		}
		
		// Simular erro interno
		if orderID == "error" {
			return WrapError(ErrInternalServer, "falha ao processar pedido no banco de dados")
		}
		
		return nil
	}

	// Handler que usa a função
	orderHandler := func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("id")
		
		err := processOrder(orderID)
		if err != nil {
			HttpErrorResponse(w, err)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"order_id": "%s", "status": "processed"}`, orderID)
	}

	// Testar diferentes cenários
	orderTests := []string{"", "invalid", "unauthorized", "error", "123"}
	
	fmt.Println("\n=== TESTES DE ENCADEAMENTO ===")
	for _, orderID := range orderTests {
		url := "/order"
		if orderID != "" {
			url += "?id=" + orderID
		}
		
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		orderHandler(w, req)
		
		fmt.Printf("Order ID '%s': Status %d - %s\n", 
			orderID, w.Code, w.Body.String())
	}
}

// ExampleJSONErrorResponse demonstra respostas JSON estruturadas
func ExampleJSONErrorResponse() {
	// Exemplo de como a resposta JSON é estruturada
	w := httptest.NewRecorder()
	HttpErrorResponse(w, ErrBadRequest, "Campo 'email' é obrigatório")

	fmt.Println("=== ESTRUTURA DA RESPOSTA JSON ===")
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	fmt.Printf("Response Body: %s\n", w.Body.String())

	// Demonstrar que a resposta segue o padrão HttpResponse
	fmt.Println("\nA resposta sempre segue a estrutura:")
	fmt.Println(`{"message": "mensagem de erro"}`)
}

// ExampleNilErrorHandling demonstra comportamento com erro nil
func ExampleNilErrorHandling() {
	w := httptest.NewRecorder()
	
	// Chamar com erro nil - não deve fazer nada
	HttpErrorResponse(w, nil)
	
	fmt.Println("=== TESTE COM ERRO NIL ===")
	fmt.Printf("Status Code: %d (deve ser 200 - padrão)\n", w.Code)
	fmt.Printf("Response Body: '%s' (deve estar vazio)\n", w.Body.String())
	fmt.Printf("Headers: %+v\n", w.Header())
}

// ExampleBestPractices demonstra melhores práticas
func ExampleBestPractices() {
	fmt.Println("=== MELHORES PRÁTICAS ===")
	fmt.Println("")
	fmt.Println("1. SEMPRE use HttpErrorResponse para erros HTTP")
	fmt.Println("2. Use erros pré-definidos quando possível")
	fmt.Println("3. Adicione contexto com WrapError quando necessário")
	fmt.Println("4. Mantenha mensagens de erro consistentes")
	fmt.Println("5. Não exponha detalhes internos em mensagens de erro")
	fmt.Println("6. Use códigos de status HTTP apropriados")
	fmt.Println("7. Sempre defina Content-Type como application/json")
	fmt.Println("8. Trate erros nil adequadamente")
	fmt.Println("9. Use middleware para validações comuns")
	fmt.Println("10. Mantenha logs detalhados para debugging interno")
}