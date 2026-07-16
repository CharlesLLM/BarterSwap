package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

const reviewColumns = `id, exchange_id, author_id, target_id, note, commentaire, created_at`

type reviewRow interface {
	Scan(...any) error
}

func scanReview(row reviewRow) (domain.Review, error) {
	var review domain.Review
	var createdAt time.Time
	err := row.Scan(
		&review.ID,
		&review.ExchangeID,
		&review.AuthorID,
		&review.TargetID,
		&review.Note,
		&review.Commentaire,
		&createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Review{}, domain.ErrExchangeNotFound
	}
	if err != nil {
		return domain.Review{}, err
	}
	review.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	return review, nil
}

func (store Store) CreateReview(
	ctx context.Context,
	exchangeID int,
	authorID int,
	targetID int,
	input domain.CreateReviewInput,
) (domain.Review, error) {
	row := store.db.QueryRowContext(ctx, `INSERT INTO reviews (
		exchange_id, author_id, target_id, note, commentaire
	) VALUES ($1, $2, $3, $4, $5)
	RETURNING `+reviewColumns, exchangeID, authorID, targetID, input.Note, input.Commentaire)

	review, err := scanReview(row)
	if err != nil {
		if hasSQLState(err, uniqueViolationSQLState) {
			return domain.Review{}, domain.ErrReviewAlreadyExists
		}
		return domain.Review{}, fmt.Errorf("création de l'avis : %w", err)
	}
	return review, nil
}

func (store Store) ListUserReviews(ctx context.Context, userID int) ([]domain.Review, error) {
	return store.listReviews(ctx, `SELECT `+reviewColumns+`
		FROM reviews WHERE target_id = $1 ORDER BY created_at DESC`, userID)
}

func (store Store) ListServiceReviews(ctx context.Context, serviceID int) ([]domain.Review, error) {
	return store.listReviews(ctx, `SELECT r.id, r.exchange_id, r.author_id, r.target_id,
		r.note, r.commentaire, r.created_at
		FROM reviews r
		JOIN exchanges e ON e.id = r.exchange_id
		WHERE e.service_id = $1
		ORDER BY r.created_at DESC`, serviceID)
}

func (store Store) listReviews(ctx context.Context, query string, id int) ([]domain.Review, error) {
	rows, err := store.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("liste des avis : %w", err)
	}
	defer rows.Close()

	reviews := []domain.Review{}
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, fmt.Errorf("lecture d'un avis : %w", err)
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des avis : %w", err)
	}
	return reviews, nil
}
