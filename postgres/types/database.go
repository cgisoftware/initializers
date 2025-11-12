package types

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Database interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	QueryRowx(query string, args ...any) *sqlx.Row

	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	NamedExec(query string, arg any) (sql.Result, error)
	NamedQuery(query string, arg any) (*sqlx.Rows, error)
	NamedQueryContext(ctx context.Context, query string, arg any) (*sqlx.Rows, error)

	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	Get(dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	Select(dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error

	Rebind(query string) string
	Ping() error
	PingContext(ctx context.Context) error
	DriverName() string
	Preparex(query string) (*sqlx.Stmt, error)
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
	Begin() (*sql.Tx, error)
}
