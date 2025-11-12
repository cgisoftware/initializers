package uow

import (
	"context"
	"errors"
	"sync"

	"github.com/cgisoftware/initializers/postgres/types"
	"github.com/jmoiron/sqlx"
)

var (
	// globalUoW is the global UnitOfWork instance
	globalUoW *UnitOfWork
	globalDB  types.Database
	uoOnce    sync.Once
)

type txKeyType string

const txKey txKeyType = "transaction"

type UnitOfWork struct {
	db types.Database
}

func SetGlobalDB(db types.Database) {
	globalDB = db
}

func New(db types.Database) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// Adiciona a transação no contexto
func (u *UnitOfWork) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, txKey, tx)

	err = fn(ctxWithTx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// Recupera uma transação ativa do contexto, se houver
func GetTx(ctx context.Context) *sqlx.Tx {
	tx, ok := ctx.Value(txKey).(*sqlx.Tx)
	if !ok {
		return nil
	}
	return tx
}

func GetUoW() *UnitOfWork {
	if globalUoW == nil {
		autoInitUoW()
	}
	return globalUoW
}

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
func WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return GetUoW().WithTransaction(ctx, fn)
}
