package uow

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/cgisoftware/initializers/postgres/types"
)

var (
	// globalUoW is the global UnitOfWork instance
	globalUoW *UnitOfWork
	globalDB  types.Database
	uoOnce    sync.Once
)

// UnitOfWork represents a unit of work pattern for database operations
type UnitOfWork struct {
	db           types.Database
	tx           *sql.Tx
	repositories map[string]any
	mu           sync.Mutex
}

// SetGlobalDB sets the global database instance
func SetGlobalDB(db types.Database) {
	globalDB = db
}

// New creates a new UnitOfWork instance (alias for NewUnitOfWork for backward compatibility)
func New(db types.Database) *UnitOfWork {
	return NewUnitOfWork(db)
}

// NewUnitOfWork creates a new UnitOfWork instance
func NewUnitOfWork(db types.Database) *UnitOfWork {
	return &UnitOfWork{
		db:           db,
		repositories: make(map[string]any),
	}
}

// Begin starts a new transaction
func (uow *UnitOfWork) Begin(ctx context.Context) error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.tx != nil {
		return errors.New("transaction already in progress")
	}

	tx, err := uow.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	uow.tx = tx
	return nil
}

// Commit commits the current transaction
func (uow *UnitOfWork) Commit() error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.tx == nil {
		return errors.New("no transaction in progress")
	}

	err := uow.tx.Commit()
	if err != nil {
		return err
	}

	uow.tx = nil
	uow.clearRepositories()
	return nil
}

// Rollback rolls back the current transaction
func (uow *UnitOfWork) Rollback() error {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.tx == nil {
		return errors.New("no transaction in progress")
	}

	err := uow.tx.Rollback()
	uow.tx = nil
	uow.clearRepositories()
	return err
}

// GetRepository returns a repository of the specified type
func (uow *UnitOfWork) GetRepository(name string) (any, bool) {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	repo, exists := uow.repositories[name]
	return repo, exists
}

// RegisterRepository registers a repository with the unit of work
func (uow *UnitOfWork) RegisterRepository(name string, repo any) {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	uow.repositories[name] = repo
}

// Exec executes a query without returning any rows
func (uow *UnitOfWork) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if uow.tx != nil {
		return uow.tx.ExecContext(ctx, query, args...)
	}
	return uow.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (uow *UnitOfWork) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if uow.tx != nil {
		return uow.tx.QueryContext(ctx, query, args...)
	}
	return uow.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (uow *UnitOfWork) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	if uow.tx != nil {
		return uow.tx.QueryRowContext(ctx, query, args...)
	}
	return uow.db.QueryRowContext(ctx, query, args...)
}

// clearRepositories clears all registered repositories
func (uow *UnitOfWork) clearRepositories() {
	uow.repositories = make(map[string]any)
}

// GetDB returns the underlying *sql.DB from the global database instance
func GetDB() (*sql.DB, bool) {
	switch db := globalDB.(type) {
	case *sqlx.DB:
		return db.DB, true
	default:
		return nil, false
	}
}

// autoInitUoW initializes the global UnitOfWork instance
// This is called automatically when needed
func autoInitUoW() {
	uoOnce.Do(func() {
		if globalUoW == nil {
			if globalDB == nil {
				panic("database not initialized. Call postgres.Initialize first")
			}
			globalUoW = New(globalDB)
		}
	})
}

// GetUoW returns the global UnitOfWork instance
func GetUoW() *UnitOfWork {
	if globalUoW == nil {
		autoInitUoW()
	}
	return globalUoW
}

// WithTransaction executes the given function within a transaction using the global UoW
func WithTransaction(ctx context.Context, fn func() error) error {
	return GetUoW().WithTransaction(ctx, fn)
}

// GetRepository returns a repository from the global UoW
func GetRepository(name string) (any, bool) {
	return GetUoW().GetRepository(name)
}

// RegisterRepository registers a repository with the global UoW
func RegisterRepository(name string, repo any) {
	GetUoW().RegisterRepository(name, repo)
}

// WithTransaction executes the given function within a transaction
func (uow *UnitOfWork) WithTransaction(ctx context.Context, fn func() error) error {
	// Begin transaction
	err := uow.Begin(ctx)
	if err != nil {
		return err
	}

	// Execute the function
	err = fn()
	if err != nil {
		// Rollback if there was an error
		if rbErr := uow.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	// Commit if no errors
	return uow.Commit()
}

// DB returns the underlying *sql.DB or *sql.Tx if in transaction
func (uow *UnitOfWork) DB() any {
	uow.mu.Lock()
	defer uow.mu.Unlock()

	if uow.tx != nil {
		return uow.tx
	}
	return uow.db
}
