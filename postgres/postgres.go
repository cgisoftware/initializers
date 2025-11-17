package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
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
func (d sqlxDB) Exec(query string, args ...any) (sql.Result, error) {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.Exec(query, args...)
	}
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
func (d sqlxDB) Get(dest any, query string, args ...any) error {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.Get(dest, query, args...)
	}
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
func (d sqlxDB) NamedExec(query string, arg any) (sql.Result, error) {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.NamedExec(query, arg)
	}
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
func (d sqlxDB) NamedQuery(query string, arg any) (*sqlx.Rows, error) {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.NamedQuery(query, arg)
	}
	return d.db.NamedQuery(query, arg)
}

// NamedQueryContext implements types.Database.
func (d sqlxDB) NamedQueryContext(ctx context.Context, query string, arg any) (*sqlx.Rows, error) {
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
func (d sqlxDB) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.PrepareNamed(query)
	}
	return d.db.PrepareNamed(query)
}

// Preparex implements types.Database.
func (d sqlxDB) Preparex(query string) (*sqlx.Stmt, error) {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.Preparex(query)
	}
	return d.db.Preparex(query)
}

// Query implements types.Database.
func (d sqlxDB) Query(query string, args ...any) (*sql.Rows, error) {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.Query(query, args...)
	}
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
func (d sqlxDB) QueryRow(query string, args ...any) *sql.Row {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.QueryRow(query, args...)
	}
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
func (d sqlxDB) QueryRowx(query string, args ...any) *sqlx.Row {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.QueryRowx(query, args...)
	}
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
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.Rebind(query)
	}
	return d.db.Rebind(query)
}

// Select implements types.Database.
func (d sqlxDB) Select(dest any, query string, args ...any) error {
	if tx := uow.GetTx(context.Background()); tx != nil {
		return tx.Select(dest, query, args...)
	}
	return d.db.Select(dest, query, args...)
}

// SelectContext implements types.Database.
func (d sqlxDB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	if tx := uow.GetTx(ctx); tx != nil {
		return tx.SelectContext(ctx, dest, query, args...)
	}
	return d.db.SelectContext(ctx, dest, query, args...)
}

type DatabaseClientConfig struct {
	databaseURL     string
	context         context.Context
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
	runMigrations   bool
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

// Initialize retorna um pool de conexões com o banco de dados
func Initialize(ctx context.Context, databaseURL string, opts ...DatabaseOption) types.Database {
	databaseOptions := &DatabaseClientConfig{
		maxOpenConns:    25,
		maxIdleConns:    10,
		connMaxLifetime: 4,
		runMigrations:   true,
		context:         ctx,
		databaseURL:     databaseURL,
	}
	for _, opt := range opts {
		opt(databaseOptions)
	}

	db, err := sqlx.ConnectContext(databaseOptions.context, "postgres", databaseOptions.databaseURL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	db.DB.SetMaxOpenConns(databaseOptions.maxOpenConns)
	db.DB.SetMaxIdleConns(databaseOptions.maxIdleConns)
	db.DB.SetConnMaxLifetime(databaseOptions.connMaxLifetime)

	if databaseOptions.runMigrations {
		runMigrations(databaseOptions.databaseURL)
	}

	database := sqlxDB{db}

	uow.SetGlobalDB(database)

	return database
}

func runMigrations(databaseURL string) {
	m, err := migrate.New("file://database/migrations", databaseURL)
	if err != nil {
		log.Println(err)
	}

	if err := m.Up(); err != nil {
		log.Println(err)
	}
}
