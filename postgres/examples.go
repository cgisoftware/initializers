package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

// User representa um usuário para exemplos
type User struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Product representa um produto para exemplos
type Product struct {
	ID          int       `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Price       float64   `db:"price" json:"price"`
	Active      bool      `db:"active" json:"active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// ExampleBasicInitialization demonstra inicialização básica do banco
func ExampleBasicInitialization() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	// Inicialização básica
	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	// Testar conexão
	err := db.Ping()
	if err != nil {
		log.Printf("Erro ao conectar: %v", err)
		return
	}

	fmt.Println("✓ Conexão com banco estabelecida com sucesso")
	fmt.Printf("Driver: %s\n", db.DriverName())
}

// ExampleInitializationWithOptions demonstra inicialização com opções
func ExampleInitializationWithOptions() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	// Inicialização com opções customizadas
	db := Initialize(ctx, databaseURL,
		WithMaxOpenConns(25),
		WithMaxIdleConns(5),
		WithConnMaxLifetime(30*time.Minute),
		WithMigrations(true),
	)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== CONFIGURAÇÃO AVANÇADA ===")
	fmt.Println("✓ Banco inicializado com configurações customizadas:")
	fmt.Println("  - Max Open Connections: 25")
	fmt.Println("  - Max Idle Connections: 5")
	fmt.Println("  - Connection Max Lifetime: 30 minutos")
	fmt.Println("  - Migrations: Habilitadas")

	// Verificar configurações (se usando *sqlx.DB)
	if sqlxDB, ok := db.(*sqlx.DB); ok {
		stats := sqlxDB.Stats()
		fmt.Printf("\nEstatísticas atuais:\n")
		fmt.Printf("  - Conexões abertas: %d\n", stats.OpenConnections)
		fmt.Printf("  - Conexões em uso: %d\n", stats.InUse)
		fmt.Printf("  - Conexões idle: %d\n", stats.Idle)
	}
}

// ExampleBasicQueries demonstra consultas básicas
func ExampleBasicQueries() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== CONSULTAS BÁSICAS ===")

	// Query simples
	rows, err := db.Query("SELECT version()")
	if err != nil {
		log.Printf("Erro na query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			log.Printf("Erro no scan: %v", err)
			continue
		}
		fmt.Printf("Versão do PostgreSQL: %s\n", version)
	}

	// QueryRow para resultado único
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&count)
	if err != nil {
		log.Printf("Erro ao contar tabelas: %v", err)
		return
	}
	fmt.Printf("Número de tabelas públicas: %d\n", count)
}

// ExampleCRUDOperations demonstra operações CRUD
func ExampleCRUDOperations() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== OPERAÇÕES CRUD ===")

	// CREATE - Inserir usuário
	insertSQL := `
		INSERT INTO users (name, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id`

	now := time.Now()
	var userID int
	err := db.QueryRowContext(ctx, insertSQL, "João Silva", "joao@exemplo.com", now, now).Scan(&userID)
	if err != nil {
		log.Printf("Erro ao inserir usuário: %v", err)
		return
	}
	fmt.Printf("✓ Usuário criado com ID: %d\n", userID)

	// READ - Buscar usuário
	var user User
	selectSQL := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
	err = db.GetContext(ctx, &user, selectSQL, userID)
	if err != nil {
		log.Printf("Erro ao buscar usuário: %v", err)
		return
	}
	fmt.Printf("✓ Usuário encontrado: %+v\n", user)

	// UPDATE - Atualizar usuário
	updateSQL := "UPDATE users SET name = $1, updated_at = $2 WHERE id = $3"
	result, err := db.ExecContext(ctx, updateSQL, "João Santos", time.Now(), userID)
	if err != nil {
		log.Printf("Erro ao atualizar usuário: %v", err)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✓ Usuário atualizado (%d linhas afetadas)\n", rowsAffected)

	// DELETE - Deletar usuário
	deleteSQL := "DELETE FROM users WHERE id = $1"
	result, err = db.ExecContext(ctx, deleteSQL, userID)
	if err != nil {
		log.Printf("Erro ao deletar usuário: %v", err)
		return
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("✓ Usuário deletado (%d linhas afetadas)\n", rowsAffected)
}

// ExampleBatchOperations demonstra operações em lote
func ExampleBatchOperations() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== OPERAÇÕES EM LOTE ===")

	// Preparar dados para inserção em lote
	users := []User{
		{Name: "Alice", Email: "alice@exemplo.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Bob", Email: "bob@exemplo.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "Carol", Email: "carol@exemplo.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	// Inserção em lote usando prepared statement
	stmt, err := db.Preparex("INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, $3, $4)")
	if err != nil {
		log.Printf("Erro ao preparar statement: %v", err)
		return
	}
	defer stmt.Close()

	for i, user := range users {
		_, err := stmt.ExecContext(ctx, user.Name, user.Email, user.CreatedAt, user.UpdatedAt)
		if err != nil {
			log.Printf("Erro ao inserir usuário %d: %v", i+1, err)
			continue
		}
		fmt.Printf("✓ Usuário %d inserido: %s\n", i+1, user.Name)
	}

	// Buscar todos os usuários inseridos
	var allUsers []User
	err = db.SelectContext(ctx, &allUsers, "SELECT id, name, email, created_at, updated_at FROM users ORDER BY id DESC LIMIT 3")
	if err != nil {
		log.Printf("Erro ao buscar usuários: %v", err)
		return
	}

	fmt.Printf("\n✓ Usuários encontrados (%d):\n", len(allUsers))
	for _, user := range allUsers {
		fmt.Printf("  ID: %d, Nome: %s, Email: %s\n", user.ID, user.Name, user.Email)
	}
}

// ExampleTransactions demonstra uso de transações
func ExampleTransactions() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== TRANSAÇÕES ===")

	// Iniciar transação
	tx, err := db.(*sqlx.DB).BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("Erro ao iniciar transação: %v", err)
		return
	}

	// Função para rollback em caso de erro
	defer func() {
		if err != nil {
			tx.Rollback()
			fmt.Println("✗ Transação revertida devido a erro")
		}
	}()

	// Inserir usuário
	var userID int
	err = tx.QueryRowContext(ctx,
		"INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
		"Transação User", "tx@exemplo.com", time.Now(), time.Now()).Scan(&userID)
	if err != nil {
		log.Printf("Erro ao inserir usuário na transação: %v", err)
		return
	}
	fmt.Printf("✓ Usuário inserido na transação (ID: %d)\n", userID)

	// Inserir produto relacionado
	_, err = tx.ExecContext(ctx,
		"INSERT INTO products (name, description, price, active, created_at) VALUES ($1, $2, $3, $4, $5)",
		"Produto Transação", "Produto criado em transação", 99.99, true, time.Now())
	if err != nil {
		log.Printf("Erro ao inserir produto na transação: %v", err)
		return
	}
	fmt.Println("✓ Produto inserido na transação")

	// Simular erro condicional
	simulateError := false // Mude para true para testar rollback
	if simulateError {
		err = fmt.Errorf("erro simulado")
		return
	}

	// Commit da transação
	err = tx.Commit()
	if err != nil {
		log.Printf("Erro ao fazer commit: %v", err)
		return
	}

	fmt.Println("✓ Transação commitada com sucesso")
}

// ExampleNamedQueries demonstra uso de named queries
func ExampleNamedQueries() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== NAMED QUERIES ===")

	// Preparar named statement
	stmt, err := db.PrepareNamed(`
		INSERT INTO users (name, email, created_at, updated_at) 
		VALUES (:name, :email, :created_at, :updated_at)
	`)
	if err != nil {
		log.Printf("Erro ao preparar named statement: %v", err)
		return
	}
	defer stmt.Close()

	// Usar struct para parâmetros
	user := User{
		Name:      "Named Query User",
		Email:     "named@exemplo.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = stmt.ExecContext(ctx, user)
	if err != nil {
		log.Printf("Erro ao executar named query: %v", err)
		return
	}
	fmt.Println("✓ Usuário inserido usando named query")

	// Buscar usando named parameters com map
	params := map[string]interface{}{
		"email": "named@exemplo.com",
	}

	var foundUser User
	query := "SELECT id, name, email, created_at, updated_at FROM users WHERE email = :email"
	nrows, err := db.(*sqlx.DB).NamedQuery(query, params)
	if err != nil {
		log.Printf("Erro na named query de busca: %v", err)
		return
	}
	defer nrows.Close()

	if nrows.Next() {
		err = nrows.StructScan(&foundUser)
		if err != nil {
			log.Printf("Erro no scan: %v", err)
			return
		}
		fmt.Printf("✓ Usuário encontrado: %+v\n", foundUser)
	}
}

// ExampleConnectionPooling demonstra monitoramento do pool de conexões
func ExampleConnectionPooling() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL,
		WithMaxOpenConns(10),
		WithMaxIdleConns(3),
		WithConnMaxLifetime(5*time.Minute),
	)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== MONITORAMENTO DO POOL ===")

	// Função para mostrar estatísticas
	showStats := func(label string) {
		if sqlxDB, ok := db.(*sqlx.DB); ok {
			stats := sqlxDB.Stats()
			fmt.Printf("%s:\n", label)
			fmt.Printf("  Max Open: %d\n", stats.MaxOpenConnections)
			fmt.Printf("  Open: %d\n", stats.OpenConnections)
			fmt.Printf("  In Use: %d\n", stats.InUse)
			fmt.Printf("  Idle: %d\n", stats.Idle)
			fmt.Printf("  Wait Count: %d\n", stats.WaitCount)
			fmt.Printf("  Wait Duration: %v\n", stats.WaitDuration)
			fmt.Println()
		}
	}

	showStats("Estado inicial")

	// Simular múltiplas conexões
	for i := 0; i < 5; i++ {
		go func(id int) {
			var count int
			err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables").Scan(&count)
			if err != nil {
				log.Printf("Erro na goroutine %d: %v", id, err)
			}
			fmt.Printf("Goroutine %d: %d tabelas\n", id, count)
		}(i)
	}

	// Aguardar um pouco e mostrar estatísticas
	time.Sleep(100 * time.Millisecond)
	showStats("Após consultas concorrentes")
}

// ExampleErrorHandling demonstra tratamento de erros
func ExampleErrorHandling() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== TRATAMENTO DE ERROS ===")

	// Erro de sintaxe SQL
	_, err := db.Query("SELECT * FORM invalid_table") // FORM em vez de FROM
	if err != nil {
		fmt.Printf("✓ Erro de sintaxe capturado: %v\n", err)
	}

	// Erro de tabela inexistente
	_, err = db.Query("SELECT * FROM tabela_inexistente")
	if err != nil {
		fmt.Printf("✓ Erro de tabela inexistente: %v\n", err)
	}

	// Erro de constraint (assumindo que existe uma constraint de email único)
	now := time.Now()
	_, err = db.Exec("INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		"User 1", "duplicate@exemplo.com", now, now)
	if err != nil {
		fmt.Printf("Primeiro insert OK ou erro: %v\n", err)
	}

	_, err = db.Exec("INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, $3, $4)",
		"User 2", "duplicate@exemplo.com", now, now)
	if err != nil {
		fmt.Printf("✓ Erro de constraint capturado: %v\n", err)
	}

	// Erro de timeout (simulado)
	ctxTimeout, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	_, err = db.QueryContext(ctxTimeout, "SELECT pg_sleep(1)")
	if err != nil {
		fmt.Printf("✓ Erro de timeout capturado: %v\n", err)
	}
}

// ExampleBestPractices demonstra melhores práticas
func ExampleBestPractices() {
	fmt.Println("=== MELHORES PRÁTICAS POSTGRES ===")
	fmt.Println("")
	fmt.Println("1. CONFIGURAÇÃO DE CONEXÃO:")
	fmt.Println("   - Configure MaxOpenConns baseado na capacidade do servidor")
	fmt.Println("   - Use MaxIdleConns para manter conexões reutilizáveis")
	fmt.Println("   - Defina ConnMaxLifetime para evitar conexões obsoletas")
	fmt.Println("   - Use SSL em produção (sslmode=require)")
	fmt.Println("")
	fmt.Println("2. CONSULTAS:")
	fmt.Println("   - Sempre use parâmetros ($1, $2) para evitar SQL injection")
	fmt.Println("   - Use prepared statements para consultas repetitivas")
	fmt.Println("   - Prefira Get/Select do sqlx para mapeamento automático")
	fmt.Println("   - Use contexto com timeout para consultas longas")
	fmt.Println("")
	fmt.Println("3. TRANSAÇÕES:")
	fmt.Println("   - Use transações para operações que devem ser atômicas")
	fmt.Println("   - Sempre faça defer rollback para cleanup")
	fmt.Println("   - Mantenha transações curtas para evitar locks")
	fmt.Println("   - Use isolation levels apropriados")
	fmt.Println("")
	fmt.Println("4. TRATAMENTO DE ERROS:")
	fmt.Println("   - Sempre verifique erros de conexão")
	fmt.Println("   - Trate diferentes tipos de erro apropriadamente")
	fmt.Println("   - Use retry logic para erros temporários")
	fmt.Println("   - Registre erros para debugging")
	fmt.Println("")
	fmt.Println("5. PERFORMANCE:")
	fmt.Println("   - Use índices apropriados")
	fmt.Println("   - Monitore slow queries")
	fmt.Println("   - Use EXPLAIN ANALYZE para otimização")
	fmt.Println("   - Considere connection pooling externo (pgbouncer)")
	fmt.Println("")
	fmt.Println("6. MIGRATIONS:")
	fmt.Println("   - Use migrations para mudanças de schema")
	fmt.Println("   - Versione suas migrations")
	fmt.Println("   - Teste migrations em ambiente similar à produção")
	fmt.Println("   - Mantenha migrations idempotentes")
}

// ExampleHealthCheck demonstra implementação de health check
func ExampleHealthCheck() {
	ctx := context.Background()
	databaseURL := "postgres://user:password@localhost:5432/dbname?sslmode=disable"

	db := Initialize(ctx, databaseURL)
	defer db.(*sqlx.DB).Close()

	fmt.Println("=== HEALTH CHECK ===")

	// Health check simples
	healthCheck := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return db.PingContext(ctx)
	}

	// Executar health check
	if err := healthCheck(); err != nil {
		fmt.Printf("✗ Health check falhou: %v\n", err)
	} else {
		fmt.Println("✓ Health check passou - banco está saudável")
	}

	// Health check mais detalhado
	detailedHealthCheck := func() map[string]interface{} {
		result := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
		}

		// Testar conexão
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		err := db.PingContext(ctx)
		latency := time.Since(start)

		if err != nil {
			result["status"] = "unhealthy"
			result["error"] = err.Error()
		} else {
			result["latency_ms"] = latency.Milliseconds()
		}

		// Adicionar estatísticas do pool
		if sqlxDB, ok := db.(*sqlx.DB); ok {
			stats := sqlxDB.Stats()
			result["pool_stats"] = map[string]interface{}{
				"open_connections": stats.OpenConnections,
				"in_use":           stats.InUse,
				"idle":             stats.Idle,
				"wait_count":       stats.WaitCount,
			}
		}

		return result
	}

	// Executar health check detalhado
	detailedResult := detailedHealthCheck()
	fmt.Printf("\nHealth check detalhado: %+v\n", detailedResult)
}
