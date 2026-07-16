package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func (store Store) CreateUser(ctx context.Context, input domain.CreateUserInput) (domain.User, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, fmt.Errorf("début de la transaction : %w", err)
	}
	defer tx.Rollback()

	var user domain.User
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
		if hasSQLState(err, uniqueViolationSQLState) {
			return domain.User{}, domain.ErrPseudoAlreadyExists
		}
		return domain.User{}, fmt.Errorf("création de l'utilisateur : %w", err)
	}

	if _, err = tx.ExecContext(
		ctx,
		`INSERT INTO credit_transactions (user_id, montant, type)
		 VALUES ($1, $2, 'welcome')`,
		user.ID,
		domain.WelcomeCredits,
	); err != nil {
		return domain.User{}, fmt.Errorf("attribution des crédits : %w", err)
	}

	if err := tx.Commit(); err != nil {
		return domain.User{}, fmt.Errorf("validation de la transaction : %w", err)
	}

	user.CreditBalance = domain.WelcomeCredits
	user.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	return user, nil
}

func (store Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := store.db.QueryContext(ctx, `SELECT u.id, u.pseudo, u.bio, u.ville,
		COALESCE(SUM(c.montant), 0), u.created_at
		FROM users u
		LEFT JOIN credit_transactions c ON c.user_id = u.id
		GROUP BY u.id
		ORDER BY u.id`)
	if err != nil {
		return nil, fmt.Errorf("liste des utilisateurs : %w", err)
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		var user domain.User
		var createdAt time.Time
		if err := rows.Scan(&user.ID, &user.Pseudo, &user.Bio, &user.Ville, &user.CreditBalance, &createdAt); err != nil {
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

func (store Store) FindUser(ctx context.Context, id int) (domain.User, error) {
	var user domain.User
	var createdAt time.Time
	err := store.db.QueryRowContext(ctx, `SELECT u.id, u.pseudo, u.bio, u.ville,
		COALESCE(SUM(c.montant), 0), u.created_at
		FROM users u
		LEFT JOIN credit_transactions c ON c.user_id = u.id
		WHERE u.id = $1
		GROUP BY u.id`, id).Scan(
		&user.ID,
		&user.Pseudo,
		&user.Bio,
		&user.Ville,
		&user.CreditBalance,
		&createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, domain.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("lecture de l'utilisateur : %w", err)
	}

	user.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	user.Skills, err = store.ListSkills(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (store Store) GetUserStats(ctx context.Context, id int) (domain.UserStats, error) {
	var userExists bool
	if err := store.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, id).Scan(&userExists); err != nil {
		return domain.UserStats{}, fmt.Errorf("vérification de l'utilisateur : %w", err)
	}
	if !userExists {
		return domain.UserStats{}, domain.ErrUserNotFound
	}

	var stats domain.UserStats
	err := store.db.QueryRowContext(ctx, `
		SELECT
			$1 AS user_id,
			(SELECT COUNT(*) FROM services WHERE provider_id = $1 AND actif = TRUE) AS services_actifs,
			0 AS echanges_completes,
			COALESCE((SELECT SUM(montant) FROM credit_transactions WHERE user_id = $1), 0) AS credit_balance,
			0::DOUBLE PRECISION AS note_moyenne,
			0 AS nb_avis,
			COALESCE((SELECT SUM(CASE WHEN type = 'earn' THEN ABS(montant) ELSE 0 END) FROM credit_transactions WHERE user_id = $1), 0) AS total_gagne,
			COALESCE((SELECT SUM(CASE WHEN type = 'spend' THEN ABS(montant) ELSE 0 END) FROM credit_transactions WHERE user_id = $1), 0) AS total_depense
	`, id).Scan(
		&stats.UserID,
		&stats.ServicesActifs,
		&stats.EchangesCompletes,
		&stats.CreditBalance,
		&stats.NoteMoyenne,
		&stats.NbAvis,
		&stats.TotalGagne,
		&stats.TotalDepense,
	)
	if err != nil {
		return domain.UserStats{}, fmt.Errorf("lecture des statistiques utilisateur : %w", err)
	}
	return stats, nil
}

func (store Store) UpdateUser(ctx context.Context, id int, input domain.CreateUserInput) (domain.User, error) {
	result, err := store.db.ExecContext(ctx, `UPDATE users
		SET pseudo = $2, bio = $3, ville = $4
		WHERE id = $1`, id, input.Pseudo, input.Bio, input.Ville)
	if err != nil {
		if hasSQLState(err, uniqueViolationSQLState) {
			return domain.User{}, domain.ErrPseudoAlreadyExists
		}
		return domain.User{}, fmt.Errorf("modification de l'utilisateur : %w", err)
	}

	updatedRows, err := result.RowsAffected()
	if err != nil {
		return domain.User{}, fmt.Errorf("vérification de la modification : %w", err)
	}
	if updatedRows == 0 {
		return domain.User{}, domain.ErrUserNotFound
	}
	return store.FindUser(ctx, id)
}

func (store Store) DeleteUser(ctx context.Context, id int) error {
	result, err := store.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("suppression de l'utilisateur : %w", err)
	}
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification de la suppression : %w", err)
	}
	if deletedRows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (store Store) ListSkills(ctx context.Context, userID int) ([]domain.Skill, error) {
	var userExists bool
	if err := store.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&userExists); err != nil {
		return nil, fmt.Errorf("vérification de l'utilisateur : %w", err)
	}
	if !userExists {
		return nil, domain.ErrUserNotFound
	}

	rows, err := store.db.QueryContext(ctx, `SELECT nom, niveau FROM skills WHERE user_id = $1 ORDER BY nom`, userID)
	if err != nil {
		return nil, fmt.Errorf("liste des compétences : %w", err)
	}
	defer rows.Close()

	skills := []domain.Skill{}
	for rows.Next() {
		var skill domain.Skill
		if err := rows.Scan(&skill.Nom, &skill.Niveau); err != nil {
			return nil, fmt.Errorf("lecture d'une compétence : %w", err)
		}
		skills = append(skills, skill)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des compétences : %w", err)
	}
	return skills, nil
}

func (store Store) ReplaceSkills(ctx context.Context, userID int, skills []domain.Skill) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("début de la transaction : %w", err)
	}
	defer tx.Rollback()

	var userExists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&userExists); err != nil {
		return fmt.Errorf("vérification de l'utilisateur : %w", err)
	}
	if !userExists {
		return domain.ErrUserNotFound
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM skills WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("suppression des compétences : %w", err)
	}
	for _, skill := range skills {
		if _, err := tx.ExecContext(ctx, `INSERT INTO skills (user_id, nom, niveau) VALUES ($1, $2, $3)`, userID, skill.Nom, skill.Niveau); err != nil {
			return fmt.Errorf("création d'une compétence : %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("validation des compétences : %w", err)
	}
	return nil
}
