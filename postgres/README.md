# Pacote Postgres

O pacote `postgres` fornece uma interface robusta e eficiente para interação com bancos de dados PostgreSQL, incluindo gerenciamento de conexões, pool de conexões, transações, queries preparadas e monitoramento de performance.

## Funcionalidades

### 🔗 Gerenciamento de Conexões
- Pool de conexões configurável
- Conexões persistentes e reutilizáveis
- Health check automático
- Reconexão automática
- Timeout configurável

### 📊 Operações de Banco
- Queries síncronas e assíncronas
- Transações com rollback automático
- Prepared statements
- Named queries
- Operações em lote (batch)
- Suporte a contexto (context.Context)

### 🔍 Consultas Avançadas
- Query builder integrado
- Mapeamento automático para structs
- Suporte a arrays e JSON
- Paginação automática
- Ordenação dinâmica

### 📈 Monitoramento
- Métricas de performance
- Logging de queries
- Estatísticas do pool
- Health check endpoint
- Alertas de performance

## Interface Principal

### `Database`
```go
type Database interface {
    // Queries básicas
    Query(query string, args ...interface{}) (*sql.Rows, error)
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRow(query string, args ...interface{}) *sql.Row
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
    QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
    
    // Execução de comandos
    Exec(query string, args ...interface{}) (sql.Result, error)
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    
    // Mapeamento para structs
    Get(dest interface{}, query string, args ...interface{}) error
    GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    Select(dest interface{}, query string, args ...interface{}) error
    SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
    
    // Utilitários
    Rebind(query string) string
    Ping() error
    PingContext(ctx context.Context) error
    DriverName() string
    
    // Prepared statements
    Preparex(query string) (*sqlx.Stmt, error)
    PrepareNamed(query string) (*sqlx.NamedStmt, error)
}
```

## Configuração

### Inicialização Básica
```go
package main

import (
    "log"
    "seu-projeto/initializers/postgres"
)

func main() {
    // Configuração básica
    config := postgres.Config{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "password",
        Database: "myapp",
        SSLMode:  "disable",
    }
    
    // Inicializar conexão
    db, err := postgres.Initialize(config)
    if err != nil {
        log.Fatal("Erro ao conectar ao banco:", err)
    }
    defer db.Close()
    
    // Testar conexão
    if err := db.Ping(); err != nil {
        log.Fatal("Erro no ping:", err)
    }
    
    log.Println("Conectado ao PostgreSQL com sucesso!")
}
```

### Configuração Avançada
```go
func setupAdvancedDB() {
    config := postgres.Config{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "password",
        Database: "myapp",
        SSLMode:  "require",
        
        // Pool de conexões
        MaxOpenConns:    25,
        MaxIdleConns:    10,
        ConnMaxLifetime: time.Hour,
        ConnMaxIdleTime: time.Minute * 30,
        
        // Timeouts
        ConnectTimeout: time.Second * 10,
        QueryTimeout:   time.Second * 30,
        
        // Logging
        LogLevel: "debug",
        LogQueries: true,
        
        // Health check
        HealthCheckInterval: time.Minute * 5,
        
        // Retry
        RetryAttempts: 3,
        RetryDelay:    time.Second * 2,
    }
    
    db, err := postgres.InitializeWithOptions(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Configurar métricas
    postgres.EnableMetrics(db)
    
    // Configurar health check
    postgres.StartHealthCheck(db, config.HealthCheckInterval)
}
```

### Configuração via Variáveis de Ambiente
```go
func setupFromEnv() {
    config := postgres.ConfigFromEnv()
    
    db, err := postgres.Initialize(config)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Banco configurado via variáveis de ambiente")
}
```

## Uso Básico

### Consultas Simples
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
)

type User struct {
    ID        int       `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
}

func exemploConsultas(db postgres.Database) {
    ctx := context.Background()
    
    // Query simples
    rows, err := db.QueryContext(ctx, "SELECT id, name, email FROM users WHERE active = $1", true)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, &user.Email)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Usuário: %+v\n", user)
    }
    
    // Query com mapeamento automático
    var users []User
    err = db.SelectContext(ctx, &users, "SELECT * FROM users WHERE created_at > $1", time.Now().AddDate(0, -1, 0))
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Usuários dos últimos 30 dias: %d\n", len(users))
    
    // Query single row
    var user User
    err = db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", 1)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Usuário encontrado: %+v\n", user)
}
```

### Operações CRUD
```go
func exemplosCRUD(db postgres.Database) {
    ctx := context.Background()
    
    // CREATE
    newUser := User{
        Name:  "João Silva",
        Email: "joao@email.com",
    }
    
    var userID int
    err := db.QueryRowContext(ctx, 
        "INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id",
        newUser.Name, newUser.Email).Scan(&userID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Usuário criado com ID: %d\n", userID)
    
    // READ
    var user User
    err = db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", userID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Usuário lido: %+v\n", user)
    
    // UPDATE
    result, err := db.ExecContext(ctx, 
        "UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2",
        "João Santos", userID)
    if err != nil {
        log.Fatal(err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    fmt.Printf("Linhas atualizadas: %d\n", rowsAffected)
    
    // DELETE
    result, err = db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
    if err != nil {
        log.Fatal(err)
    }
    
    rowsAffected, _ = result.RowsAffected()
    fmt.Printf("Linhas deletadas: %d\n", rowsAffected)
}
```

### Transações
```go
func exemploTransacao(db postgres.Database) {
    ctx := context.Background()
    
    // Iniciar transação
    tx, err := db.BeginTxx(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Defer rollback (será ignorado se commit for bem-sucedido)
    defer func() {
        if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
            log.Printf("Erro no rollback: %v", err)
        }
    }()
    
    // Operações dentro da transação
    var userID int
    err = tx.QueryRowContext(ctx,
        "INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) RETURNING id",
        "Maria Silva", "maria@email.com").Scan(&userID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Inserir perfil do usuário
    _, err = tx.ExecContext(ctx,
        "INSERT INTO user_profiles (user_id, bio, avatar_url) VALUES ($1, $2, $3)",
        userID, "Desenvolvedora", "https://example.com/avatar.jpg")
    if err != nil {
        log.Fatal(err)
    }
    
    // Commit da transação
    if err = tx.Commit(); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Transação concluída com sucesso. Usuário ID: %d\n", userID)
}
```

### Named Queries
```go
func exemploNamedQueries(db postgres.Database) {
    ctx := context.Background()
    
    // Preparar named query
    stmt, err := db.PrepareNamed(`
        INSERT INTO users (name, email, age, created_at, updated_at) 
        VALUES (:name, :email, :age, NOW(), NOW()) 
        RETURNING id
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()
    
    // Dados para inserção
    users := []map[string]interface{}{
        {
            "name":  "Ana Costa",
            "email": "ana@email.com",
            "age":   28,
        },
        {
            "name":  "Carlos Lima",
            "email": "carlos@email.com",
            "age":   35,
        },
    }
    
    // Inserir múltiplos usuários
    for _, userData := range users {
        var userID int
        err = stmt.QueryRowContext(ctx, userData).Scan(&userID)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Usuário %s criado com ID: %d\n", userData["name"], userID)
    }
    
    // Named query para seleção
    var users []User
    query := `SELECT * FROM users WHERE age BETWEEN :min_age AND :max_age ORDER BY created_at DESC`
    
    params := map[string]interface{}{
        "min_age": 25,
        "max_age": 40,
    }
    
    rows, err := db.NamedQueryContext(ctx, query, params)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    
    for rows.Next() {
        var user User
        err = rows.StructScan(&user)
        if err != nil {
            log.Fatal(err)
        }
        users = append(users, user)
    }
    
    fmt.Printf("Usuários encontrados: %d\n", len(users))
}
```

## Operações Avançadas

### Operações em Lote (Batch)
```go
func exemploBatch(db postgres.Database) {
    ctx := context.Background()
    
    // Preparar dados para inserção em lote
    users := []User{
        {Name: "User 1", Email: "user1@email.com"},
        {Name: "User 2", Email: "user2@email.com"},
        {Name: "User 3", Email: "user3@email.com"},
    }
    
    // Usar transação para operações em lote
    tx, err := db.BeginTxx(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer tx.Rollback()
    
    // Preparar statement
    stmt, err := tx.PreparexContext(ctx, 
        "INSERT INTO users (name, email, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())")
    if err != nil {
        log.Fatal(err)
    }
    defer stmt.Close()
    
    // Executar inserções em lote
    for _, user := range users {
        _, err = stmt.ExecContext(ctx, user.Name, user.Email)
        if err != nil {
            log.Fatal(err)
        }
    }
    
    // Commit
    if err = tx.Commit(); err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Inseridos %d usuários em lote\n", len(users))
}
```

### Paginação
```go
type PaginationResult struct {
    Data       []User `json:"data"`
    Total      int    `json:"total"`
    Page       int    `json:"page"`
    PerPage    int    `json:"per_page"`
    TotalPages int    `json:"total_pages"`
}

func exemploPaginacao(db postgres.Database, page, perPage int) (*PaginationResult, error) {
    ctx := context.Background()
    
    // Calcular offset
    offset := (page - 1) * perPage
    
    // Query para dados paginados
    var users []User
    err := db.SelectContext(ctx, &users, `
        SELECT * FROM users 
        WHERE active = true 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `, perPage, offset)
    if err != nil {
        return nil, err
    }
    
    // Query para total de registros
    var total int
    err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE active = true").Scan(&total)
    if err != nil {
        return nil, err
    }
    
    // Calcular total de páginas
    totalPages := (total + perPage - 1) / perPage
    
    return &PaginationResult{
        Data:       users,
        Total:      total,
        Page:       page,
        PerPage:    perPage,
        TotalPages: totalPages,
    }, nil
}
```

### Queries Complexas com Joins
```go
type UserWithProfile struct {
    User
    Bio       string `db:"bio"`
    AvatarURL string `db:"avatar_url"`
}

func exemploJoins(db postgres.Database) {
    ctx := context.Background()
    
    var usersWithProfiles []UserWithProfile
    
    query := `
        SELECT 
            u.id, u.name, u.email, u.created_at, u.updated_at,
            p.bio, p.avatar_url
        FROM users u
        LEFT JOIN user_profiles p ON u.id = p.user_id
        WHERE u.active = true
        ORDER BY u.created_at DESC
    `
    
    err := db.SelectContext(ctx, &usersWithProfiles, query)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, user := range usersWithProfiles {
        fmt.Printf("Usuário: %s, Bio: %s\n", user.Name, user.Bio)
    }
}
```

### Trabalho com JSON
```go
type UserMetadata struct {
    Preferences map[string]interface{} `json:"preferences"`
    Settings    map[string]string      `json:"settings"`
    Tags        []string               `json:"tags"`
}

type UserWithMetadata struct {
    User
    Metadata UserMetadata `db:"metadata"`
}

func exemploJSON(db postgres.Database) {
    ctx := context.Background()
    
    // Inserir usuário com dados JSON
    metadata := UserMetadata{
        Preferences: map[string]interface{}{
            "theme":       "dark",
            "language":    "pt-BR",
            "notifications": true,
        },
        Settings: map[string]string{
            "timezone": "America/Sao_Paulo",
            "currency": "BRL",
        },
        Tags: []string{"developer", "golang", "postgresql"},
    }
    
    metadataJSON, _ := json.Marshal(metadata)
    
    var userID int
    err := db.QueryRowContext(ctx, `
        INSERT INTO users (name, email, metadata, created_at, updated_at) 
        VALUES ($1, $2, $3, NOW(), NOW()) 
        RETURNING id
    `, "João Dev", "joao.dev@email.com", metadataJSON).Scan(&userID)
    if err != nil {
        log.Fatal(err)
    }
    
    // Consultar usuário com dados JSON
    var user UserWithMetadata
    err = db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", userID)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Usuário com metadata: %+v\n", user)
    
    // Query usando operadores JSON
    var users []User
    err = db.SelectContext(ctx, &users, `
        SELECT id, name, email, created_at, updated_at 
        FROM users 
        WHERE metadata->>'theme' = $1
    `, "dark")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Usuários com tema dark: %d\n", len(users))
}
```

## Monitoramento e Métricas

### Estatísticas do Pool
```go
func monitorarPool(db postgres.Database) {
    // Obter estatísticas do pool
    stats := db.Stats()
    
    fmt.Printf("Pool Stats:\n")
    fmt.Printf("  Conexões abertas: %d\n", stats.OpenConnections)
    fmt.Printf("  Conexões em uso: %d\n", stats.InUse)
    fmt.Printf("  Conexões idle: %d\n", stats.Idle)
    fmt.Printf("  Total de esperas: %d\n", stats.WaitCount)
    fmt.Printf("  Tempo total de espera: %v\n", stats.WaitDuration)
    fmt.Printf("  Conexões fechadas (max idle): %d\n", stats.MaxIdleClosed)
    fmt.Printf("  Conexões fechadas (max lifetime): %d\n", stats.MaxLifetimeClosed)
    
    // Alertas baseados nas métricas
    if stats.WaitCount > 100 {
        log.Printf("ALERTA: Muitas esperas no pool (%d)", stats.WaitCount)
    }
    
    if float64(stats.InUse)/float64(stats.OpenConnections) > 0.8 {
        log.Printf("ALERTA: Pool com alta utilização (%.2f%%)", 
            float64(stats.InUse)/float64(stats.OpenConnections)*100)
    }
}
```

### Health Check
```go
func healthCheck(db postgres.Database) error {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
    defer cancel()
    
    // Ping básico
    if err := db.PingContext(ctx); err != nil {
        return fmt.Errorf("ping falhou: %v", err)
    }
    
    // Query de teste
    var result int
    err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
    if err != nil {
        return fmt.Errorf("query de teste falhou: %v", err)
    }
    
    if result != 1 {
        return fmt.Errorf("resultado inesperado: %d", result)
    }
    
    return nil
}

// Handler HTTP para health check
func healthCheckHandler(db postgres.Database) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if err := healthCheck(db); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "error":  err.Error(),
            })
            return
        }
        
        stats := db.Stats()
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "healthy",
            "database": map[string]interface{}{
                "open_connections": stats.OpenConnections,
                "in_use":           stats.InUse,
                "idle":             stats.Idle,
            },
        })
    }
}
```

### Logging de Queries
```go
type QueryLogger struct {
    db postgres.Database
}

func NewQueryLogger(db postgres.Database) *QueryLogger {
    return &QueryLogger{db: db}
}

func (ql *QueryLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    start := time.Now()
    
    rows, err := ql.db.QueryContext(ctx, query, args...)
    
    duration := time.Since(start)
    
    // Log da query
    log.Printf("QUERY [%v] %s %v", duration, query, args)
    
    if err != nil {
        log.Printf("QUERY ERROR: %v", err)
    }
    
    // Alerta para queries lentas
    if duration > time.Second {
        log.Printf("SLOW QUERY ALERT [%v]: %s", duration, query)
    }
    
    return rows, err
}

func (ql *QueryLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    start := time.Now()
    
    result, err := ql.db.ExecContext(ctx, query, args...)
    
    duration := time.Since(start)
    
    log.Printf("EXEC [%v] %s %v", duration, query, args)
    
    if err != nil {
        log.Printf("EXEC ERROR: %v", err)
    }
    
    return result, err
}
```

## Integração com Frameworks

### Middleware Gin
```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "seu-projeto/initializers/postgres"
)

// DatabaseMiddleware injeta a instância do banco no contexto
func DatabaseMiddleware(db postgres.Database) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Set("db", db)
        c.Next()
    }
}

// GetDB obtém a instância do banco do contexto
func GetDB(c *gin.Context) postgres.Database {
    db, exists := c.Get("db")
    if !exists {
        panic("Database não encontrado no contexto")
    }
    return db.(postgres.Database)
}

// Exemplo de handler
func GetUsersHandler(c *gin.Context) {
    db := GetDB(c)
    
    var users []User
    err := db.SelectContext(c.Request.Context(), &users, "SELECT * FROM users WHERE active = true")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, users)
}
```

### Repository Pattern
```go
package repository

import (
    "context"
    "seu-projeto/initializers/postgres"
)

type UserRepository struct {
    db postgres.Database
}

func NewUserRepository(db postgres.Database) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (name, email, created_at, updated_at) 
        VALUES ($1, $2, NOW(), NOW()) 
        RETURNING id, created_at, updated_at
    `
    
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).
        Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    var user User
    err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
    query := `
        UPDATE users 
        SET name = $1, email = $2, updated_at = NOW() 
        WHERE id = $3
        RETURNING updated_at
    `
    
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email, user.ID).
        Scan(&user.UpdatedAt)
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
    _, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
    return err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
    var users []User
    err := r.db.SelectContext(ctx, &users, `
        SELECT * FROM users 
        WHERE active = true 
        ORDER BY created_at DESC 
        LIMIT $1 OFFSET $2
    `, limit, offset)
    return users, err
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
    var count int
    err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE active = true").Scan(&count)
    return count, err
}
```

## Migrações

### Sistema de Migração
```go
package migrations

import (
    "fmt"
    "io/ioutil"
    "path/filepath"
    "sort"
    "strconv"
    "strings"
    "seu-projeto/initializers/postgres"
)

type Migration struct {
    Version int
    Name    string
    Up      string
    Down    string
}

type Migrator struct {
    db postgres.Database
}

func NewMigrator(db postgres.Database) *Migrator {
    return &Migrator{db: db}
}

func (m *Migrator) CreateMigrationsTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            applied_at TIMESTAMP DEFAULT NOW()
        )
    `
    
    _, err := m.db.Exec(query)
    return err
}

func (m *Migrator) LoadMigrations(dir string) ([]Migration, error) {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    
    var migrations []Migration
    
    for _, file := range files {
        if !strings.HasSuffix(file.Name(), ".sql") {
            continue
        }
        
        parts := strings.Split(file.Name(), "_")
        if len(parts) < 2 {
            continue
        }
        
        version, err := strconv.Atoi(parts[0])
        if err != nil {
            continue
        }
        
        content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
        if err != nil {
            return nil, err
        }
        
        // Separar UP e DOWN
        sections := strings.Split(string(content), "-- +migrate Down")
        up := strings.TrimSpace(sections[0])
        up = strings.TrimPrefix(up, "-- +migrate Up")
        up = strings.TrimSpace(up)
        
        var down string
        if len(sections) > 1 {
            down = strings.TrimSpace(sections[1])
        }
        
        name := strings.Join(parts[1:], "_")
        name = strings.TrimSuffix(name, ".sql")
        
        migrations = append(migrations, Migration{
            Version: version,
            Name:    name,
            Up:      up,
            Down:    down,
        })
    }
    
    // Ordenar por versão
    sort.Slice(migrations, func(i, j int) bool {
        return migrations[i].Version < migrations[j].Version
    })
    
    return migrations, nil
}

func (m *Migrator) GetAppliedMigrations() (map[int]bool, error) {
    applied := make(map[int]bool)
    
    rows, err := m.db.Query("SELECT version FROM schema_migrations")
    if err != nil {
        return applied, err
    }
    defer rows.Close()
    
    for rows.Next() {
        var version int
        if err := rows.Scan(&version); err != nil {
            return applied, err
        }
        applied[version] = true
    }
    
    return applied, nil
}

func (m *Migrator) Up(migrations []Migration) error {
    applied, err := m.GetAppliedMigrations()
    if err != nil {
        return err
    }
    
    for _, migration := range migrations {
        if applied[migration.Version] {
            fmt.Printf("Migração %d já aplicada: %s\n", migration.Version, migration.Name)
            continue
        }
        
        fmt.Printf("Aplicando migração %d: %s\n", migration.Version, migration.Name)
        
        // Executar migração em transação
        tx, err := m.db.Begin()
        if err != nil {
            return err
        }
        
        // Executar SQL da migração
        if _, err := tx.Exec(migration.Up); err != nil {
            tx.Rollback()
            return fmt.Errorf("erro na migração %d: %v", migration.Version, err)
        }
        
        // Registrar migração aplicada
        if _, err := tx.Exec(
            "INSERT INTO schema_migrations (version, name) VALUES ($1, $2)",
            migration.Version, migration.Name); err != nil {
            tx.Rollback()
            return err
        }
        
        if err := tx.Commit(); err != nil {
            return err
        }
        
        fmt.Printf("Migração %d aplicada com sucesso\n", migration.Version)
    }
    
    return nil
}
```

### Exemplo de Arquivo de Migração
```sql
-- migrations/001_create_users_table.sql
-- +migrate Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(active);

-- +migrate Down
DROP TABLE IF EXISTS users;
```

## Testes

### Testes Unitários
```go
package postgres_test

import (
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "seu-projeto/initializers/postgres"
)

func setupTestDB(t *testing.T) postgres.Database {
    config := postgres.Config{
        Host:     "localhost",
        Port:     5432,
        User:     "postgres",
        Password: "password",
        Database: "test_db",
        SSLMode:  "disable",
    }
    
    db, err := postgres.Initialize(config)
    require.NoError(t, err)
    
    // Limpar tabelas
    _, err = db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
    require.NoError(t, err)
    
    return db
}

func TestUserCRUD(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    ctx := context.Background()
    
    // Test Create
    var userID int
    err := db.QueryRowContext(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        "Test User", "test@email.com").Scan(&userID)
    require.NoError(t, err)
    assert.Greater(t, userID, 0)
    
    // Test Read
    var user User
    err = db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", userID)
    require.NoError(t, err)
    assert.Equal(t, "Test User", user.Name)
    assert.Equal(t, "test@email.com", user.Email)
    
    // Test Update
    _, err = db.ExecContext(ctx,
        "UPDATE users SET name = $1 WHERE id = $2",
        "Updated User", userID)
    require.NoError(t, err)
    
    // Verify Update
    err = db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", userID)
    require.NoError(t, err)
    assert.Equal(t, "Updated User", user.Name)
    
    // Test Delete
    result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
    require.NoError(t, err)
    
    rowsAffected, err := result.RowsAffected()
    require.NoError(t, err)
    assert.Equal(t, int64(1), rowsAffected)
}

func TestTransaction(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    ctx := context.Background()
    
    // Test successful transaction
    tx, err := db.BeginTxx(ctx, nil)
    require.NoError(t, err)
    
    var userID int
    err = tx.QueryRowContext(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        "Transaction User", "tx@email.com").Scan(&userID)
    require.NoError(t, err)
    
    err = tx.Commit()
    require.NoError(t, err)
    
    // Verify user exists
    var count int
    err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 1, count)
    
    // Test rollback
    tx, err = db.BeginTxx(ctx, nil)
    require.NoError(t, err)
    
    _, err = tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
    require.NoError(t, err)
    
    err = tx.Rollback()
    require.NoError(t, err)
    
    // Verify user still exists
    err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = $1", userID).Scan(&count)
    require.NoError(t, err)
    assert.Equal(t, 1, count)
}

func BenchmarkInsert(b *testing.B) {
    db := setupTestDB(b)
    defer db.Close()
    
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := db.ExecContext(ctx,
            "INSERT INTO users (name, email) VALUES ($1, $2)",
            fmt.Sprintf("User %d", i), fmt.Sprintf("user%d@email.com", i))
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkSelect(b *testing.B) {
    db := setupTestDB(b)
    defer db.Close()
    
    ctx := context.Background()
    
    // Inserir dados de teste
    for i := 0; i < 1000; i++ {
        _, err := db.ExecContext(ctx,
            "INSERT INTO users (name, email) VALUES ($1, $2)",
            fmt.Sprintf("User %d", i), fmt.Sprintf("user%d@email.com", i))
        if err != nil {
            b.Fatal(err)
        }
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var users []User
        err := db.SelectContext(ctx, &users, "SELECT * FROM users LIMIT 10")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Configuração de Produção

### Variáveis de Ambiente
```bash
# Configurações de conexão
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="password"
export DB_NAME="myapp"
export DB_SSLMODE="require"

# Pool de conexões
export DB_MAX_OPEN_CONNS="25"
export DB_MAX_IDLE_CONNS="10"
export DB_CONN_MAX_LIFETIME="1h"
export DB_CONN_MAX_IDLE_TIME="30m"

# Timeouts
export DB_CONNECT_TIMEOUT="10s"
export DB_QUERY_TIMEOUT="30s"

# Logging
export DB_LOG_LEVEL="info"
export DB_LOG_QUERIES="false"

# Health check
export DB_HEALTH_CHECK_INTERVAL="5m"

# Retry
export DB_RETRY_ATTEMPTS="3"
export DB_RETRY_DELAY="2s"
```

### Docker Compose
```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 30s
      timeout: 10s
      retries: 3

  app:
    build: .
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: password
      DB_NAME: myapp
      DB_SSLMODE: disable
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "8080:8080"

volumes:
  postgres_data:
```

## Melhores Práticas

### 1. Gerenciamento de Conexões
```go
// ✅ Configurar pool adequadamente
config := postgres.Config{
    MaxOpenConns:    25,  // Não muito alto
    MaxIdleConns:    10,  // Menor que MaxOpenConns
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: time.Minute * 30,
}

// ✅ Sempre usar contexto com timeout
ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
defer cancel()

// ❌ Evitar conexões sem timeout
// db.Query("SELECT * FROM large_table") // Pode travar indefinidamente
```

### 2. Tratamento de Erros
```go
// ✅ Verificar tipos específicos de erro
if err != nil {
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound
    }
    return nil, fmt.Errorf("erro na consulta: %w", err)
}

// ✅ Log de erros com contexto
log.Printf("Erro na query [%s] com args %v: %v", query, args, err)
```

### 3. Transações
```go
// ✅ Sempre usar defer para rollback
tx, err := db.BeginTxx(ctx, nil)
if err != nil {
    return err
}
defer func() {
    if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
        log.Printf("Erro no rollback: %v", err)
    }
}()

// ... operações ...

return tx.Commit()
```

### 4. Performance
```go
// ✅ Usar prepared statements para queries repetidas
stmt, err := db.Preparex("SELECT * FROM users WHERE status = $1")
if err != nil {
    return err
}
defer stmt.Close()

// ✅ Usar LIMIT em queries que podem retornar muitos dados
query := "SELECT * FROM users ORDER BY created_at DESC LIMIT $1"

// ✅ Criar índices apropriados
// CREATE INDEX idx_users_status ON users(status);
// CREATE INDEX idx_users_created_at ON users(created_at);
```

### 5. Segurança
```go
// ✅ Sempre usar placeholders para evitar SQL injection
query := "SELECT * FROM users WHERE email = $1"
rows, err := db.Query(query, email)

// ❌ Nunca concatenar strings diretamente
// query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email) // PERIGOSO!

// ✅ Validar entrada
if !isValidEmail(email) {
    return ErrInvalidEmail
}
```

## Dependências

- `database/sql` - Interface padrão do Go para SQL
- `github.com/jmoiron/sqlx` - Extensões para database/sql
- `github.com/lib/pq` - Driver PostgreSQL
- `context` - Controle de contexto e timeout
- `time` - Manipulação de tempo

## Veja Também

- [Pacote Formatter](../formatter/README.md) - Para formatação de erros
- [Pacote Validator](../validator/README.md) - Para validação de dados
- [Pacote OpenTelemetry](../opentelemetry/README.md) - Para observabilidade
- [Pacote Auth](../auth/README.md) - Para autenticação

---

**Nota**: Este pacote requer PostgreSQL 12+ para funcionalidades completas. Certifique-se de configurar adequadamente o pool de conexões para sua carga de trabalho específica.