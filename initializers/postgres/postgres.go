package postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

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
func Initialize(ctx context.Context, databaseURL string, opts ...DatabaseOption) Database {
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

	return db
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
