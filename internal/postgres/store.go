package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

const uniqueViolationSQLState = "23505"

//go:embed schema.sql
var schema string

type Store struct {
	db *sql.DB
}

func New(ctx context.Context, databaseURL string) (Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return Store{}, fmt.Errorf("ouverture de la base de données : %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return Store{}, fmt.Errorf("connexion à la base de données : %w", err)
	}

	return Store{db: db}, nil
}

func (store Store) Close() error {
	return store.db.Close()
}

func (store Store) CreateSchema(ctx context.Context) error {
	if _, err := store.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("création du schéma SQL : %w", err)
	}
	return nil
}

type sqlStateError interface {
	SQLState() string
}

func hasSQLState(err error, state string) bool {
	var stateError sqlStateError
	return errors.As(err, &stateError) && stateError.SQLState() == state
}
