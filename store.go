package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la base de données : %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connexion à la base de données : %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() {
	if err := s.db.Close(); err != nil {
		fmt.Printf("fermeture de la base de données : %v\n", err)
	}
}

func (s *Store) CreateSchema(ctx context.Context) error {
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("lecture du schéma SQL : %w", err)
	}

	if _, err := s.db.ExecContext(ctx, string(schema)); err != nil {
		return fmt.Errorf("création du schéma SQL : %w", err)
	}

	return nil
}

func (s *Store) InsertUser(ctx context.Context, input CreateUserInput) (User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("début de la transaction : %w", err)
	}

	defer tx.Rollback()

	var user User
	var createdAt time.Time

	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO users (pseudo, bio, ville)
		 VALUES ($1, $2, $3)
		 RETURNING id, pseudo, bio, ville, created_at`,
		input.Pseudo,
		input.Bio,
		input.Ville,
	).Scan(&user.ID, &user.Pseudo, &user.Bio, &user.Ville, &createdAt)

	if err != nil {
		var pqError *pq.Error
		if errors.As(err, &pqError) && pqError.Code == "23505" {
			return User{}, ErrPseudoAlreadyExists
		}

		return User{}, fmt.Errorf("création de l'utilisateur : %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO credit_transactions (user_id, montant, type)
		 VALUES ($1, $2, 'welcome')`,
		user.ID,
		welcomeCredits,
	)

	if err != nil {
		return User{}, fmt.Errorf("attribution des crédits : %w", err)
	}

	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("validation de la transaction : %w", err)
	}

	user.CreditBalance = welcomeCredits
	user.CreatedAt = createdAt.UTC().Format(time.RFC3339)

	return user, nil
}

func (s *Store) SelectUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT u.id, u.pseudo, u.bio, u.ville,
		        COALESCE(SUM(c.montant), 0), u.created_at
		 FROM users u
		 LEFT JOIN credit_transactions c ON c.user_id = u.id
		 GROUP BY u.id
		 ORDER BY u.id`,
	)

	if err != nil {
		return nil, fmt.Errorf("liste des utilisateurs : %w", err)
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var user User
		var createdAt time.Time

		if err := rows.Scan(
			&user.ID,
			&user.Pseudo,
			&user.Bio,
			&user.Ville,
			&user.CreditBalance,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("lecture d'un utilisateur : %w", err)
		}

		user.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des utilisateurs : %w", err)
	}

	return users, nil
}

func (s *Store) DeleteUser(ctx context.Context, id int) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("suppression de l'utilisateur : %w", err)
	}

	deletedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification de la suppression : %w", err)
	}

	if deletedRows == 0 {
		return ErrUserNotFound
	}

	return nil
}
