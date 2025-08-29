package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// UserRegistration representa dados de registro de usuário
type UserRegistration struct {
	Name     string `json:"name" validate:"required,min=2,max=50" label:"Nome"`
	Email    string `json:"email" validate:"required,email" label:"E-mail"`
	Password string `json:"password" validate:"required,min=8" label:"Senha"`
	Age      int    `json:"age" validate:"required,min=18,max=120" label:"Idade"`
	Phone    string `json:"phone" validate:"required,phone" label:"Telefone"`
}

// ProductCreate representa dados de criação de produto
type ProductCreate struct {
	Name        string  `json:"name" validate:"required,min=3,max=100" label:"Nome do Produto"`
	Description string  `json:"description" validate:"required,min=10,max=500" label:"Descrição"`
	Price       float64 `json:"price" validate:"required,gt=0" label:"Preço"`
	Category    string  `json:"category" validate:"required,oneof=electronics clothing books" label:"Categoria"`
	SKU         string  `json:"sku" validate:"required,alphanum,len=8" label:"SKU"`
	Tags        []string `json:"tags" validate:"required,min=1,max=5,dive,min=2,max=20" label:"Tags"`
}

// AddressForm representa dados de endereço
type AddressForm struct {
	Street     string `json:"street" validate:"required,min=5,max=100" label:"Rua"`
	Number     string `json:"number" validate:"required,min=1,max=10" label:"Número"`
	Complement string `json:"complement" validate:"omitempty,max=50" label:"Complemento"`
	Neighborhood string `json:"neighborhood" validate:"required,min=3,max=50" label:"Bairro"`
	City       string `json:"city" validate:"required,min=2,max=50" label:"Cidade"`
	State      string `json:"state" validate:"required,len=2" label:"Estado"`
	ZipCode    string `json:"zip_code" validate:"required,zipcode" label:"CEP"`
	Country    string `json:"country" validate:"required,iso3166_1_alpha2" label:"País"`
}

// ExampleBasicInitialization demonstra inicialização básica do validator
func ExampleBasicInitialization() {
	fmt.Println("=== INICIALIZAÇÃO BÁSICA ===")

	// Inicialização simples
	Initialize()

	fmt.Println("✓ Validator inicializado com configuração padrão")
	fmt.Printf("Validator global configurado\n")

	// Teste de validação simples
	user := UserRegistration{
		Name:     "João",
		Email:    "joao@exemplo.com",
		Password: "senha123456",
		Age:      25,
		Phone:    "11999999999",
	}

	err := ValidateStruct(user)
	if err != nil {
		fmt.Printf("Erro de validação: %v\n", err)
	} else {
		fmt.Println("✓ Validação passou com sucesso")
	}
}

// ExampleInitializationWithOptions demonstra inicialização com opções
func ExampleInitializationWithOptions() {
	fmt.Println("=== INICIALIZAÇÃO COM OPÇÕES ===")

	// Dicionário personalizado
	dicionario := map[string]map[string]string{
		"en": {
			"required": "é obrigatório",
			"email":    "deve ser um e-mail válido",
			"min":      "deve ter pelo menos {0} caracteres",
			"max":      "deve ter no máximo {0} caracteres",
			"gt":       "deve ser maior que {0}",
		},
	}

	// Traduções personalizadas
	traducoes := map[string]map[string]string{
		"en": {
			"UserRegistration.Name":     "Nome do Usuário",
			"UserRegistration.Email":    "Endereço de E-mail",
			"UserRegistration.Password": "Senha de Acesso",
			"UserRegistration.Age":      "Idade",
			"UserRegistration.Phone":    "Número de Telefone",
		},
	}

	// Inicializar com opções
	Initialize(
		WithDicionario(dicionario),
		WithTraducoes(traducoes),
	)

	fmt.Println("✓ Validator inicializado com:")
	fmt.Printf("  - Dicionário personalizado (%d idiomas)\n", len(dicionario))
	fmt.Printf("  - Traduções personalizadas (%d idiomas)\n", len(traducoes))

	// Teste com dados inválidos para ver mensagens personalizadas
	user := UserRegistration{
		Name:     "A", // Muito curto
		Email:    "email-inválido",
		Password: "123", // Muito curto
		Age:      15,   // Menor que 18
		Phone:    "",   // Vazio
	}

	err := ValidateStruct(user)
	if err != nil {
		fmt.Println("\nErros de validação encontrados:")
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, err := range validationErrors {
				fmt.Printf("  - Campo: %s, Erro: %s\n", err.Field(), err.Tag())
			}
		}
	}
}

// ExampleValidationErrors demonstra tratamento de erros de validação
func ExampleValidationErrors() {
	fmt.Println("=== TRATAMENTO DE ERROS ===")

	Initialize()

	// Dados com múltiplos erros
	product := ProductCreate{
		Name:        "AB", // Muito curto
		Description: "Desc", // Muito curto
		Price:       -10, // Negativo
		Category:    "invalid", // Não está na lista
		SKU:         "123", // Muito curto e não alfanumérico
		Tags:        []string{}, // Array vazio
	}

	err := ValidateStruct(product)
	if err != nil {
		fmt.Println("Erros de validação detalhados:")
		
		// Converter para ValidationErrors para acesso detalhado
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				fmt.Printf("\nCampo: %s\n", fieldError.Field())
				fmt.Printf("  Valor: %v\n", fieldError.Value())
				fmt.Printf("  Tag: %s\n", fieldError.Tag())
				fmt.Printf("  Parâmetro: %s\n", fieldError.Param())
				fmt.Printf("  Namespace: %s\n", fieldError.Namespace())
				fmt.Printf("  Struct Namespace: %s\n", fieldError.StructNamespace())
			}
		}
	}
}

// ExampleCustomValidationTags demonstra tags de validação personalizadas
func ExampleCustomValidationTags() {
	fmt.Println("=== TAGS DE VALIDAÇÃO PERSONALIZADAS ===")

	Initialize()

	fmt.Println("Nota: Validações personalizadas devem ser registradas durante a inicialização")

	fmt.Println("✓ Validações personalizadas registradas:")
	fmt.Println("  - cpf: Valida formato de CPF")
	fmt.Println("  - strong_password: Valida senha forte")

	// Struct com validações personalizadas
	type UserWithCustomValidation struct {
		Name     string `validate:"required,min=2" label:"Nome"`
		CPF      string `validate:"required,cpf" label:"CPF"`
		Password string `validate:"required,strong_password" label:"Senha"`
	}

	// Teste com dados válidos
	validUser := UserWithCustomValidation{
		Name:     "João Silva",
		CPF:      "12345678901",
		Password: "MinhaSenh@123",
	}

	err := ValidateStruct(validUser)
	if err != nil {
		fmt.Printf("Erro inesperado: %v\n", err)
	} else {
		fmt.Println("✓ Usuário com validações personalizadas passou")
	}

	// Teste com dados inválidos
	invalidUser := UserWithCustomValidation{
		Name:     "A",
		CPF:      "123",
		Password: "senha",
	}

	err = ValidateStruct(invalidUser)
	if err != nil {
		fmt.Println("\nErros com validações personalizadas:")
		for _, err := range err.(validator.ValidationErrors) {
			fmt.Printf("  - %s: falhou na validação '%s'\n", err.Field(), err.Tag())
		}
	}
}

// ExampleWebIntegration demonstra integração com frameworks web
func ExampleWebIntegration() {
	fmt.Println("=== INTEGRAÇÃO COM FRAMEWORKS WEB ===")

	Initialize()

	// Função para validar e formatar erros
	validateAndFormatErrors := func(user UserRegistration) map[string]interface{} {
		err := ValidateStruct(user)
		if err != nil {
			// Converter erros de validação para formato amigável
			validationErrors := make([]Field, 0)
			
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				for _, err := range validationErrs {
					validationErrors = append(validationErrors, Field{
						Field:   err.Field(),
						Message: fmt.Sprintf("Campo %s falhou na validação %s", err.Field(), err.Tag()),
					})
				}
			}

			return map[string]interface{}{
				"success": false,
				"message": "Dados de entrada inválidos",
				"errors":  validationErrors,
			}
		}

		return map[string]interface{}{
			"success": true,
			"message": "Usuário validado com sucesso",
			"user": map[string]interface{}{
				"name":  user.Name,
				"email": user.Email,
				"age":   user.Age,
			},
		}
	}

	fmt.Println("✓ Função de validação criada")
	fmt.Println("✓ Inclui:")
	fmt.Println("  - Validação de struct")
	fmt.Println("  - Formatação de erros")
	fmt.Println("  - Resposta estruturada")

	// Teste com dados válidos
	validUser := UserRegistration{
		Name:     "João Silva",
		Email:    "joao@exemplo.com",
		Password: "senha123456",
		Age:      25,
		Phone:    "11999999999",
	}

	result := validateAndFormatErrors(validUser)
	fmt.Printf("\nResultado da validação: %+v\n", result)
}

// ExampleValidationHelper demonstra criação de helpers de validação
func ExampleValidationHelper() {
	fmt.Println("=== HELPERS DE VALIDAÇÃO ===")

	Initialize()

	// Helper genérico de validação
	validationHelper := func(model interface{}) map[string]interface{} {
		err := ValidateStruct(model)
		if err != nil {
			validationErrors := make([]Field, 0)
			
			if validationErrs, ok := err.(validator.ValidationErrors); ok {
				for _, err := range validationErrs {
					validationErrors = append(validationErrors, Field{
						Field:   err.Field(),
						Message: fmt.Sprintf("%s é inválido", err.Field()),
						Errs:    err.Tag(),
					})
				}
			}

			return map[string]interface{}{
				"valid":  false,
				"errors": validationErrors,
			}
		}

		return map[string]interface{}{
			"valid":  true,
			"errors": nil,
		}
	}

	fmt.Println("✓ Helper de validação criado")
	fmt.Println("✓ Funcionalidades:")
	fmt.Println("  - Validação automática")
	fmt.Println("  - Formatação de erros")
	fmt.Println("  - Resposta estruturada")

	// Teste do helper
	user := UserRegistration{
		Name: "A", // Inválido
	}

	result := validationHelper(user)
	fmt.Printf("\nResultado da validação: %+v\n", result)
}

// ExampleComplexValidation demonstra validação de estruturas complexas
func ExampleComplexValidation() {
	fmt.Println("=== VALIDAÇÃO COMPLEXA ===")

	Initialize()

	// Estrutura complexa com validação aninhada
	type Order struct {
		ID       string          `json:"id" validate:"required,uuid4" label:"ID do Pedido"`
		User     UserRegistration `json:"user" validate:"required" label:"Usuário"`
		Address  AddressForm     `json:"address" validate:"required" label:"Endereço"`
		Products []ProductCreate `json:"products" validate:"required,min=1,max=10,dive" label:"Produtos"`
		Total    float64         `json:"total" validate:"required,gt=0" label:"Total"`
		Notes    string          `json:"notes" validate:"omitempty,max=500" label:"Observações"`
	}

	// Dados de teste válidos
	validOrder := Order{
		ID: "550e8400-e29b-41d4-a716-446655440000",
		User: UserRegistration{
			Name:     "João Silva",
			Email:    "joao@exemplo.com",
			Password: "senha123456",
			Age:      30,
			Phone:    "11999999999",
		},
		Address: AddressForm{
			Street:       "Rua das Flores, 123",
			Number:       "123",
			Complement:   "Apto 45",
			Neighborhood: "Centro",
			City:         "São Paulo",
			State:        "SP",
			ZipCode:      "01234567",
			Country:      "BR",
		},
		Products: []ProductCreate{
			{
				Name:        "Smartphone",
				Description: "Smartphone Android com 128GB",
				Price:       899.99,
				Category:    "electronics",
				SKU:         "SMRT0001",
				Tags:        []string{"smartphone", "android", "128gb"},
			},
		},
		Total: 899.99,
		Notes: "Entrega rápida solicitada",
	}

	// Validar estrutura complexa
	err := ValidateStruct(validOrder)
	if err != nil {
		fmt.Printf("Erro inesperado na validação: %v\n", err)
	} else {
		fmt.Println("✓ Pedido complexo validado com sucesso")
		fmt.Printf("  - ID: %s\n", validOrder.ID)
		fmt.Printf("  - Usuário: %s\n", validOrder.User.Name)
		fmt.Printf("  - Cidade: %s\n", validOrder.Address.City)
		fmt.Printf("  - Produtos: %d\n", len(validOrder.Products))
		fmt.Printf("  - Total: R$ %.2f\n", validOrder.Total)
	}

	// Teste com dados inválidos
	invalidOrder := Order{
		ID: "invalid-uuid",
		User: UserRegistration{
			Name:  "A", // Muito curto
			Email: "email-inválido",
			Age:   15, // Menor que 18
		},
		Products: []ProductCreate{}, // Array vazio
		Total:    -100, // Negativo
	}

	err = ValidateStruct(invalidOrder)
	if err != nil {
		fmt.Println("\nErros encontrados na validação complexa:")
		for i, err := range err.(validator.ValidationErrors) {
			fmt.Printf("  %d. %s: %s (valor: %v)\n", i+1, err.Namespace(), err.Tag(), err.Value())
		}
	}
}

// ExampleBestPractices demonstra melhores práticas
func ExampleBestPractices() {
	fmt.Println("=== MELHORES PRÁTICAS VALIDATOR ===")
	fmt.Println("")
	fmt.Println("1. ESTRUTURAÇÃO DE TAGS:")
	fmt.Println("   - Use tags descritivas e específicas")
	fmt.Println("   - Combine múltiplas validações com vírgula")
	fmt.Println("   - Use 'omitempty' para campos opcionais")
	fmt.Println("   - Defina labels para mensagens amigáveis")
	fmt.Println("")
	fmt.Println("2. VALIDAÇÕES PERSONALIZADAS:")
	fmt.Println("   - Registre validações específicas do domínio")
	fmt.Println("   - Use nomes descritivos para tags customizadas")
	fmt.Println("   - Implemente validações reutilizáveis")
	fmt.Println("   - Documente validações personalizadas")
	fmt.Println("")
	fmt.Println("3. TRATAMENTO DE ERROS:")
	fmt.Println("   - Sempre verifique o tipo de erro")
	fmt.Println("   - Formate erros para o usuário final")
	fmt.Println("   - Use estruturas padronizadas para respostas")
	fmt.Println("   - Inclua contexto suficiente nos erros")
	fmt.Println("")
	fmt.Println("4. PERFORMANCE:")
	fmt.Println("   - Reutilize instâncias do validator")
	fmt.Println("   - Cache validações compiladas")
	fmt.Println("   - Evite validações desnecessárias")
	fmt.Println("   - Use validação condicional quando apropriado")
	fmt.Println("")
	fmt.Println("5. INTEGRAÇÃO:")
	fmt.Println("   - Integre com frameworks web (Gin, Echo, etc.)")
	fmt.Println("   - Use middlewares para validação automática")
	fmt.Println("   - Padronize respostas de erro")
	fmt.Println("   - Implemente validação em camadas")
	fmt.Println("")
	fmt.Println("6. INTERNACIONALIZAÇÃO:")
	fmt.Println("   - Configure traduções para diferentes idiomas")
	fmt.Println("   - Use dicionários personalizados")
	fmt.Println("   - Mantenha mensagens consistentes")
	fmt.Println("   - Considere contexto cultural nas validações")
}

// ExamplePerformanceOptimization demonstra otimizações de performance
func ExamplePerformanceOptimization() {
	fmt.Println("=== OTIMIZAÇÃO DE PERFORMANCE ===")

	// Inicialização única do validator global
	Initialize()
	fmt.Println("✓ Validator global inicializado")

	// Função de validação otimizada
	validateStruct := func(s interface{}) error {
		return ValidateStruct(s)
	}

	// Teste de performance com múltiplas validações
	users := []UserRegistration{
		{Name: "User1", Email: "user1@test.com", Password: "password123", Age: 25, Phone: "11999999999"},
		{Name: "User2", Email: "user2@test.com", Password: "password123", Age: 30, Phone: "11888888888"},
		{Name: "User3", Email: "user3@test.com", Password: "password123", Age: 35, Phone: "11777777777"},
	}

	validCount := 0
	for i, user := range users {
		if err := validateStruct(user); err == nil {
			validCount++
			fmt.Printf("✓ Usuário %d validado\n", i+1)
		} else {
			fmt.Printf("✗ Usuário %d inválido: %v\n", i+1, err)
		}
	}

	fmt.Printf("\nResultado: %d/%d usuários válidos\n", validCount, len(users))
	fmt.Println("\nDicas de performance:")
	fmt.Println("  - Reutilize a mesma instância do validator")
	fmt.Println("  - Evite criar novos validators desnecessariamente")
	fmt.Println("  - Use validação em lote quando possível")
	fmt.Println("  - Cache resultados de validações custosas")
}

// Função auxiliar para verificar se string é numérica
func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

// ExampleErrorFormatting demonstra formatação avançada de erros
func ExampleErrorFormatting() {
	fmt.Println("=== FORMATAÇÃO AVANÇADA DE ERROS ===")

	Initialize()

	// Função para formatar erros de validação
	formatValidationErrors := func(err error) map[string]interface{} {
		result := map[string]interface{}{
			"success": false,
			"message": "Erro de validação",
			"errors":  make(map[string]string),
			"fields":  make([]string, 0),
		}

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMap := make(map[string]string)
			fields := make([]string, 0)

			for _, fieldError := range validationErrors {
				fieldName := fieldError.Field()
				fields = append(fields, fieldName)

				// Mensagem personalizada baseada na tag
				var message string
				switch fieldError.Tag() {
				case "required":
					message = fmt.Sprintf("%s é obrigatório", fieldName)
				case "email":
					message = fmt.Sprintf("%s deve ser um e-mail válido", fieldName)
				case "min":
					message = fmt.Sprintf("%s deve ter pelo menos %s caracteres", fieldName, fieldError.Param())
				case "max":
					message = fmt.Sprintf("%s deve ter no máximo %s caracteres", fieldName, fieldError.Param())
				case "gt":
					message = fmt.Sprintf("%s deve ser maior que %s", fieldName, fieldError.Param())
				default:
					message = fmt.Sprintf("%s é inválido", fieldName)
				}

				errorMap[fieldName] = message
			}

			result["errors"] = errorMap
			result["fields"] = fields
			result["count"] = len(fields)
		}

		return result
	}

	// Teste com dados inválidos
	user := UserRegistration{
		Name:     "A",
		Email:    "email-inválido",
		Password: "123",
		Age:      15,
		Phone:    "",
	}

	err := ValidateStruct(user)
	if err != nil {
		formattedError := formatValidationErrors(err)
		fmt.Println("Erro formatado:")
		fmt.Printf("  Success: %v\n", formattedError["success"])
		fmt.Printf("  Message: %s\n", formattedError["message"])
		fmt.Printf("  Count: %v\n", formattedError["count"])
		fmt.Println("  Errors:")
		
		if errors, ok := formattedError["errors"].(map[string]string); ok {
			for field, message := range errors {
				fmt.Printf("    %s: %s\n", field, message)
			}
		}
	}
}