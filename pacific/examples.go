package pacific

import (
	"encoding/json"
	"fmt"
	"log"
)

// ExampleBasicPacificInput demonstra criação básica de input para Pacific
func ExampleBasicPacificInput() {
	// Dados de exemplo
	usuario := "user123"
	senha := "password456"
	programa := "PROG001"
	metodo := "consultarCliente"
	valor := `{"id": 123, "nome": "João Silva"}`

	// Criar input Pacific
	input := NewPacificInput(usuario, senha, programa, metodo, valor)

	// Converter para JSON
	jsonData, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		log.Printf("Erro ao serializar: %v", err)
		return
	}

	fmt.Println("=== PACIFIC INPUT BÁSICO ===")
	fmt.Printf("JSON gerado:\n%s\n", string(jsonData))

	// Mostrar estrutura dos parâmetros
	fmt.Println("\nParâmetros incluídos:")
	for i, param := range input.Params {
		fmt.Printf("  %d. %s (%s, %s): %s\n", 
			i+1, param.Parametro, param.DataType, param.ParamType, param.Valor)
	}
}

// ExamplePacificInputWithColab demonstra uso com senha de colaborador
func ExamplePacificInputWithColab() {
	usuario := "admin"
	senha := "adminpass"
	programa := "PROG002"
	metodo := "alterarPermissoes"
	valor := `{"usuario_id": 456, "permissoes": ["read", "write"]}`
	senhaColab := "colabpass123"

	// Criar input com senha de colaborador
	input := NewPacificInputColab(usuario, senha, programa, metodo, valor, senhaColab)

	jsonData, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		log.Printf("Erro ao serializar: %v", err)
		return
	}

	fmt.Println("=== PACIFIC INPUT COM COLABORADOR ===")
	fmt.Printf("JSON gerado:\n%s\n", string(jsonData))

	fmt.Printf("\nTotal de parâmetros: %d\n", len(input.Params))
	fmt.Println("Inclui senha de colaborador para operações administrativas")
}

// ExampleDadosStruct demonstra uso da estrutura Dados
func ExampleDadosStruct() {
	// Criar estrutura de dados
	dados := Dados{
		Usuario:  "operador01",
		Senha:    "op123456",
		Programa: "CONSULTAS",
		Metodo:   "buscarProduto",
		Valor:    `{"codigo": "PROD001", "categoria": "eletronicos"}`,
		IsGed:    true,
	}

	fmt.Println("=== ESTRUTURA DADOS ===")
	fmt.Printf("Usuário: %s\n", dados.Usuario)
	fmt.Printf("Programa: %s\n", dados.Programa)
	fmt.Printf("Método: %s\n", dados.Metodo)
	fmt.Printf("É GED: %t\n", dados.IsGed)
	fmt.Printf("Valor: %s\n", dados.Valor)

	// Converter para PacificInput
	input := NewPacificInput(dados.Usuario, dados.Senha, dados.Programa, dados.Metodo, dados.Valor)
	fmt.Println("\nConvertido para PacificInput com sucesso")
	fmt.Printf("UID: %s, PtoP: %s\n", input.UID, input.PtoP)
}

// ExampleErrorHandling demonstra tratamento de erros do Pacific
func ExampleErrorHandling() {
	fmt.Println("=== TRATAMENTO DE ERROS PACIFIC ===")

	// Exemplo 1: Resposta com erro de aplicação
	errorResponse1 := `{
		"logErroApp": [
			{"id": 1, "erro": "Usuário não encontrado"},
			{"id": 2, "erro": "Permissão negada"}
		]
	}`

	var logErro LogErroApp
	err := json.Unmarshal([]byte(errorResponse1), &logErro)
	if err != nil {
		log.Printf("Erro ao deserializar: %v", err)
		return
	}

	fmt.Println("Exemplo 1 - LogErroApp:")
	fmt.Printf("  É erro: %t\n", logErro.IsErr())
	for _, erro := range logErro.LogErroApp {
		fmt.Printf("  ID %d: %s\n", erro.ID, erro.Erro)
	}

	// Exemplo 2: Verificar se resposta contém erro
	errorResponse2 := []byte(`{"status": "error", "msg": "Falha na conexão"}`)
	isError := IsResponseErr(errorResponse2)
	fmt.Printf("\nExemplo 2 - IsResponseErr: %t\n", isError)

	// Exemplo 3: Resposta de sucesso
	successResponse := []byte(`{"data": {"id": 123, "nome": "Produto Teste"}}`)
	isSuccess := !IsResponseErr(successResponse)
	fmt.Printf("Exemplo 3 - Resposta de sucesso: %t\n", isSuccess)
}

// ExampleLogErr001 demonstra uso da estrutura LogErr001
func ExampleLogErr001() {
	// Simular resposta de erro 001
	errorJSON := `{"status": "error", "msg": "Parâmetros inválidos"}`

	var logErr LogErr001
	err := json.Unmarshal([]byte(errorJSON), &logErr)
	if err != nil {
		log.Printf("Erro ao deserializar LogErr001: %v", err)
		return
	}

	fmt.Println("=== LOG ERR001 ===")
	fmt.Printf("Status: %s\n", logErr.Status)
	fmt.Printf("Mensagem: %s\n", logErr.Msg)

	// Verificar se é erro usando função auxiliar
	isErr001 := isLogErr001([]byte(errorJSON))
	fmt.Printf("É erro 001: %t\n", isErr001)
}

// ExampleCompleteWorkflow demonstra um fluxo completo
func ExampleCompleteWorkflow() {
	fmt.Println("=== FLUXO COMPLETO PACIFIC ===")

	// 1. Preparar dados
	dados := Dados{
		Usuario:  "sistema",
		Senha:    "sys123",
		Programa: "VENDAS",
		Metodo:   "processarPedido",
		Valor:    `{"pedido_id": 789, "cliente_id": 456, "total": 1500.00}`,
		IsGed:    false,
	}

	fmt.Printf("1. Dados preparados: %+v\n", dados)

	// 2. Criar input Pacific
	input := NewPacificInput(dados.Usuario, dados.Senha, dados.Programa, dados.Metodo, dados.Valor)
	fmt.Printf("2. Input Pacific criado com %d parâmetros\n", len(input.Params))

	// 3. Serializar para envio
	jsonData, err := json.Marshal(input)
	if err != nil {
		log.Printf("Erro na serialização: %v", err)
		return
	}
	fmt.Printf("3. JSON serializado (%d bytes)\n", len(jsonData))

	// 4. Simular resposta (em produção, viria da API)
	responseJSON := `{
		"resultado": {
			"pedido_id": 789,
			"status": "processado",
			"numero_nf": "NF-2024-001"
		}
	}`

	// 5. Verificar se há erros na resposta
	if IsResponseErr([]byte(responseJSON)) {
		fmt.Println("4. ERRO: Resposta contém erro")
		return
	}

	fmt.Println("4. Resposta processada com sucesso")
	fmt.Printf("5. Resultado: %s\n", responseJSON)
}

// ExampleParameterTypes demonstra diferentes tipos de parâmetros
func ExampleParameterTypes() {
	fmt.Println("=== TIPOS DE PARÂMETROS PACIFIC ===")

	// Criar input customizado com diferentes tipos
	input := PacificInput{
		UID:  "dev_user",
		PWD:  "dev_pass",
		PtoP: "DEV_PROG",
		Params: []Param{
			{
				Parametro: "pcTexto",
				DataType:  "char",
				ParamType: "input",
				Valor:     "Texto de exemplo",
			},
			{
				Parametro: "pcNumero",
				DataType:  "integer",
				ParamType: "input",
				Valor:     "12345",
			},
			{
				Parametro: "pcDecimal",
				DataType:  "decimal",
				ParamType: "input",
				Valor:     "999.99",
			},
			{
				Parametro: "pcData",
				DataType:  "date",
				ParamType: "input",
				Valor:     "2024-01-15",
			},
			{
				Parametro: "pcJSON",
				DataType:  "longchar",
				ParamType: "input",
				Valor:     `{"complexo": true, "array": [1,2,3]}`,
			},
			{
				Parametro: "pcRetorno",
				DataType:  "longchar",
				ParamType: "output",
				Valor:     "",
			},
		},
	}

	fmt.Println("Parâmetros configurados:")
	for i, param := range input.Params {
		fmt.Printf("  %d. %s:\n", i+1, param.Parametro)
		fmt.Printf("     Tipo: %s (%s)\n", param.DataType, param.ParamType)
		fmt.Printf("     Valor: %s\n", param.Valor)
	}

	jsonData, _ := json.MarshalIndent(input, "", "  ")
	fmt.Printf("\nJSON completo:\n%s\n", string(jsonData))
}

// ExampleBestPractices demonstra melhores práticas
func ExampleBestPractices() {
	fmt.Println("=== MELHORES PRÁTICAS PACIFIC ===")
	fmt.Println("")
	fmt.Println("1. SEGURANÇA:")
	fmt.Println("   - Nunca hardcode credenciais no código")
	fmt.Println("   - Use variáveis de ambiente ou cofres de segredos")
	fmt.Println("   - Implemente rotação de senhas")
	fmt.Println("   - Use HTTPS sempre")
	fmt.Println("")
	fmt.Println("2. TRATAMENTO DE ERROS:")
	fmt.Println("   - Sempre verifique IsResponseErr() antes de processar")
	fmt.Println("   - Implemente retry logic para falhas temporárias")
	fmt.Println("   - Registre erros para auditoria")
	fmt.Println("   - Não exponha detalhes internos ao usuário final")
	fmt.Println("")
	fmt.Println("3. PERFORMANCE:")
	fmt.Println("   - Reutilize conexões quando possível")
	fmt.Println("   - Configure timeouts adequados")
	fmt.Println("   - Use pool de conexões para alta concorrência")
	fmt.Println("   - Monitore latência e throughput")
	fmt.Println("")
	fmt.Println("4. DADOS:")
	fmt.Println("   - Valide parâmetros antes de enviar")
	fmt.Println("   - Use tipos de dados apropriados")
	fmt.Println("   - Escape caracteres especiais em JSON")
	fmt.Println("   - Limite tamanho de payloads")
	fmt.Println("")
	fmt.Println("5. MONITORAMENTO:")
	fmt.Println("   - Registre todas as chamadas para auditoria")
	fmt.Println("   - Monitore taxa de erro e latência")
	fmt.Println("   - Configure alertas para falhas")
	fmt.Println("   - Implemente health checks")
}

// ExampleJSONHandling demonstra manipulação de JSON
func ExampleJSONHandling() {
	fmt.Println("=== MANIPULAÇÃO DE JSON ===")

	// Exemplo de dados complexos
	complexData := map[string]interface{}{
		"cliente": map[string]interface{}{
			"id":   123,
			"nome": "Empresa ABC Ltda",
			"endereco": map[string]string{
				"rua":    "Rua das Flores, 123",
				"cidade": "São Paulo",
				"cep":    "01234-567",
			},
		},
		"produtos": []map[string]interface{}{
			{"id": 1, "nome": "Produto A", "preco": 99.90},
			{"id": 2, "nome": "Produto B", "preco": 149.90},
		},
		"total": 249.80,
	}

	// Serializar dados complexos
	jsonBytes, err := json.Marshal(complexData)
	if err != nil {
		log.Printf("Erro ao serializar dados complexos: %v", err)
		return
	}

	jsonString := string(jsonBytes)
	fmt.Printf("Dados complexos serializados:\n%s\n\n", jsonString)

	// Criar input Pacific com dados complexos
	input := NewPacificInput("user", "pass", "VENDAS", "criarPedido", jsonString)

	// Mostrar como fica no parâmetro
	for _, param := range input.Params {
		if param.Parametro == "pcParametros" {
			fmt.Printf("Parâmetro pcParametros:\n%s\n", param.Valor)
			break
		}
	}

	// Exemplo de deserialização de resposta
	responseJSON := `{
		"pedido": {
			"id": 789,
			"status": "criado",
			"data_criacao": "2024-01-15T10:30:00Z"
		}
	}`

	var response map[string]interface{}
	err = json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		log.Printf("Erro ao deserializar resposta: %v", err)
		return
	}

	fmt.Printf("\nResposta deserializada: %+v\n", response)
}

// ExampleValidation demonstra validação de dados
func ExampleValidation() {
	fmt.Println("=== VALIDAÇÃO DE DADOS ===")

	// Função para validar dados antes de criar input
	validateData := func(dados Dados) []string {
		var errors []string

		if dados.Usuario == "" {
			errors = append(errors, "Usuário é obrigatório")
		}
		if dados.Senha == "" {
			errors = append(errors, "Senha é obrigatória")
		}
		if dados.Programa == "" {
			errors = append(errors, "Programa é obrigatório")
		}
		if dados.Metodo == "" {
			errors = append(errors, "Método é obrigatório")
		}
		if dados.Valor == "" {
			errors = append(errors, "Valor é obrigatório")
		}

		// Validar se Valor é JSON válido
		if dados.Valor != "" {
			var temp interface{}
			if err := json.Unmarshal([]byte(dados.Valor), &temp); err != nil {
				errors = append(errors, "Valor deve ser JSON válido")
			}
		}

		return errors
	}

	// Teste com dados válidos
	dadosValidos := Dados{
		Usuario:  "user123",
		Senha:    "pass456",
		Programa: "TEST",
		Metodo:   "validar",
		Valor:    `{"teste": true}`,
	}

	errors := validateData(dadosValidos)
	if len(errors) == 0 {
		fmt.Println("✓ Dados válidos - criando input Pacific")
		input := NewPacificInput(dadosValidos.Usuario, dadosValidos.Senha, 
			dadosValidos.Programa, dadosValidos.Metodo, dadosValidos.Valor)
		fmt.Printf("  Input criado com %d parâmetros\n", len(input.Params))
	} else {
		fmt.Println("✗ Dados inválidos:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	// Teste com dados inválidos
	dadosInvalidos := Dados{
		Usuario: "user123",
		// Senha omitida
		Programa: "TEST",
		// Método omitido
		Valor: `{json inválido}`,
	}

	errors = validateData(dadosInvalidos)
	if len(errors) > 0 {
		fmt.Println("\n✗ Dados inválidos encontrados:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	}
}