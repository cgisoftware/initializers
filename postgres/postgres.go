package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cgisoftware/initializers/postgres/types"
	"github.com/cgisoftware/initializers/postgres/uow"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type sqlxDB struct {
	db types.Database
}

// Begin implements types.Database.
func (d sqlxDB) Begin() (*sql.Tx, error) {
	return d.db.Begin()
}

// BeginTx implements types.Database.
func (d sqlxDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, opts)
}

// BeginTxx implements types.Database.
func (d sqlxDB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {

	return d.db.BeginTxx(ctx, opts)
}

// DriverName implements types.Database.
func (d sqlxDB) DriverName() string {
	return d.db.DriverName()
}

// Exec implements types.Database.
//
// Não participa de uow.WithTransaction: sem um context.Context real não há como
// recuperar a transação ativa. Use ExecContext dentro de uma unit of work.
func (d sqlxDB) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

// ExecContext implements types.Database.
func (d sqlxDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return d.db.ExecContext(ctx, query, args...)
}

// Get implements types.Database.
//
// Não participa de uow.WithTransaction: use GetContext dentro de uma unit of work.
func (d sqlxDB) Get(dest any, query string, args ...any) error {
	return d.db.Get(dest, query, args...)
}

// GetContext implements types.Database.
func (d sqlxDB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.GetContext(ctx, dest, query, args...)
	}
	return d.db.GetContext(ctx, dest, query, args...)
}

// NamedExec implements types.Database.
//
// Não participa de uow.WithTransaction: use NamedExecContext dentro de uma unit of work.
func (d sqlxDB) NamedExec(query string, arg any) (sql.Result, error) {
	return d.db.NamedExec(query, arg)
}

// NamedExecContext implements types.Database.
func (d sqlxDB) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.NamedExecContext(ctx, query, arg)
	}
	return d.db.NamedExecContext(ctx, query, arg)
}

// NamedQuery implements types.Database.
//
// Não participa de uow.WithTransaction: use NamedQueryContext dentro de uma unit of work.
func (d sqlxDB) NamedQuery(query string, arg any) (*sqlx.Rows, error) {
	return d.db.NamedQuery(query, arg)
}

// NamedQueryContext implements types.Database.
func (d sqlxDB) NamedQueryContext(ctx context.Context, query string, arg any) (*sqlx.Rows, error) {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.NamedQuery(query, arg)
	}
	return d.db.NamedQueryContext(ctx, query, arg)
}

// Ping implements types.Database.
func (d sqlxDB) Ping() error {
	return d.db.Ping()
}

// PingContext implements types.Database.
func (d sqlxDB) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// PrepareNamed implements types.Database.
//
// Não participa de uow.WithTransaction: statements preparados fora de contexto não
// têm como amarrar a uma transação ativa.
func (d sqlxDB) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	return d.db.PrepareNamed(query)
}

// Preparex implements types.Database.
//
// Não participa de uow.WithTransaction: statements preparados fora de contexto não
// têm como amarrar a uma transação ativa.
func (d sqlxDB) Preparex(query string) (*sqlx.Stmt, error) {
	return d.db.Preparex(query)
}

// Query implements types.Database.
//
// Não participa de uow.WithTransaction: use QueryContext dentro de uma unit of work.
func (d sqlxDB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

// QueryContext implements types.Database.
func (d sqlxDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow implements types.Database.
//
// Não participa de uow.WithTransaction: use QueryRowContext dentro de uma unit of work.
func (d sqlxDB) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

// QueryRowContext implements types.Database.
func (d sqlxDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return d.db.QueryRowContext(ctx, query, args...)
}

// QueryRowx implements types.Database.
//
// Não participa de uow.WithTransaction: use QueryRowxContext dentro de uma unit of work.
func (d sqlxDB) QueryRowx(query string, args ...any) *sqlx.Row {
	return d.db.QueryRowx(query, args...)
}

// QueryRowxContext implements types.Database.
func (d sqlxDB) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.QueryRowxContext(ctx, query, args...)
	}
	return d.db.QueryRowxContext(ctx, query, args...)
}

// Rebind implements types.Database.
func (d sqlxDB) Rebind(query string) string {
	return d.db.Rebind(query)
}

// Select implements types.Database.
//
// Não participa de uow.WithTransaction: use SelectContext dentro de uma unit of work.
func (d sqlxDB) Select(dest any, query string, args ...any) error {
	return d.db.Select(dest, query, args...)
}

// SelectContext implements types.Database.
func (d sqlxDB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.SelectContext(ctx, dest, query, args...)
	}
	return d.db.SelectContext(ctx, dest, query, args...)
}

// Close implements types.Database.
func (d sqlxDB) Close() error {
	return d.db.Close()
}

// Stats implements types.Database.
func (d sqlxDB) Stats() sql.DBStats {
	return d.db.Stats()
}

type DatabaseClientConfig struct {
	databaseURL     string
	context         context.Context
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	runMigrations   bool
	migrationsPath  string
}

type DatabaseOption func(d *DatabaseClientConfig)

func WithMaxOpenConns(value int) DatabaseOption {
	return func(c *DatabaseClientConfig) {
		c.maxOpenConns = value
	}
}

func WithMaxIdleConns(value int) DatabaseOption {
	return func(c *DatabaseClientConfig) {
		c.maxIdleConns = value
	}
}

func WithConnMaxLifetime(value time.Duration) DatabaseOption {
	return func(c *DatabaseClientConfig) {
		c.connMaxLifetime = value
	}
}

func WithMigrations(value bool) DatabaseOption {
	return func(c *DatabaseClientConfig) {
		c.runMigrations = value
	}
}

func WithMigrationsPath(path string) DatabaseOption {
	return func(c *DatabaseClientConfig) {
		c.migrationsPath = path
	}
}

// ShutdownFunc encerra o pool de conexões de forma graciosa: aguarda as queries
// em andamento terminarem (comportamento nativo de sql.DB.Close()), mas respeita
// o cancelamento/timeout do ctx recebido, retornando um erro envolvendo ctx.Err()
// caso o fechamento demore mais que o prazo concedido.
//
// É idempotente e segura para chamadas concorrentes: apenas a primeira chamada
// efetivamente fecha o pool; as demais recebem o mesmo resultado sem duplicar o
// Close(). Isso evita problemas quando o shutdown é acionado tanto por um defer
// quanto por um handler de sinal (SIGINT/SIGTERM), por exemplo.
type ShutdownFunc func(ctx context.Context) error

func newShutdownFunc(db types.Database) ShutdownFunc {
	var (
		once        sync.Once
		shutdownErr error
	)

	return func(ctx context.Context) error {
		once.Do(func() {
			closed := make(chan error, 1)
			go func() {
				closed <- db.Close()
			}()

			select {
			case err := <-closed:
				shutdownErr = err
			case <-ctx.Done():
				shutdownErr = fmt.Errorf("shutdown: tempo esgotado aguardando conexões em uso: %w", ctx.Err())
			}
		})
		return shutdownErr
	}
}

// Initialize retorna um pool de conexões com o banco de dados e uma ShutdownFunc
// para ser chamada no encerramento gracioso da aplicação (tipicamente a partir de
// um handler de SIGINT/SIGTERM).
func Initialize(ctx context.Context, databaseURL string, opts ...DatabaseOption) (types.Database, ShutdownFunc, error) {
	databaseOptions := &DatabaseClientConfig{
		maxOpenConns:    25,
		maxIdleConns:    10,
		connMaxLifetime: 5 * time.Minute,
		runMigrations:   true,
		migrationsPath:  "file://database/migrations",
		context:         ctx,
		databaseURL:     databaseURL,
	}
	for _, opt := range opts {
		opt(databaseOptions)
	}

	db, err := sqlx.ConnectContext(databaseOptions.context, "postgres", databaseOptions.databaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	db.DB.SetMaxOpenConns(databaseOptions.maxOpenConns)
	db.DB.SetMaxIdleConns(databaseOptions.maxIdleConns)
	db.DB.SetConnMaxLifetime(databaseOptions.connMaxLifetime)

	if databaseOptions.runMigrations {
		if err := runMigrations(databaseOptions.databaseURL, databaseOptions.migrationsPath); err != nil {
			return nil, nil, fmt.Errorf("unable to run migrations: %w", err)
		}
	}

	database := sqlxDB{db}

	uow.SetGlobalDB(database)

	return database, newShutdownFunc(database), nil
}

func runMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, databaseURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
