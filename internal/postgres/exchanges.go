package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

const exchangeColumns = `id, service_id, requester_id, owner_id, status,
       created_at, updated_at, credits`

type exchangeRow interface {
	Scan(...any) error
}

func scanExchange(row exchangeRow) (domain.Exchange, error) {
	var exchange domain.Exchange
	var createdAt, updatedAt time.Time
	err := row.Scan(
		&exchange.ID,
		&exchange.ServiceID,
		&exchange.RequesterID,
		&exchange.OwnerID,
		&exchange.Status,
		&createdAt,
		&updatedAt,
		&exchange.Credits,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Exchange{}, domain.ErrExchangeNotFound
	}
	if err != nil {
		return domain.Exchange{}, err
	}
	exchange.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	exchange.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	return exchange, nil
}

func (store Store) CreditBalance(ctx context.Context, userID int) (int, error) {
	var balance int
	err := store.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(c.montant), 0)
		FROM users u
		LEFT JOIN credit_transactions c ON c.user_id = u.id
		WHERE u.id = $1
		GROUP BY u.id`, userID).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, domain.ErrUserNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("lecture du solde : %w", err)
	}
	return balance, nil
}

func (store Store) CreateExchange(ctx context.Context, requesterID int, offeredService domain.Service) (domain.Exchange, error) {
	row := store.db.QueryRowContext(ctx, `INSERT INTO exchanges (
		service_id, requester_id, owner_id, credits, status
	) VALUES ($1, $2, $3, $4, 'pending')
	RETURNING `+exchangeColumns,
		offeredService.ID,
		requesterID,
		offeredService.ProviderID,
		offeredService.Credits,
	)
	exchange, err := scanExchange(row)
	if err != nil {
		if hasSQLState(err, uniqueViolationSQLState) {
			return domain.Exchange{}, domain.ErrExchangeConflict
		}
		return domain.Exchange{}, fmt.Errorf("création de l'échange : %w", err)
	}
	return exchange, nil
}

func (store Store) ListExchanges(ctx context.Context, userID int, filter domain.ExchangeFilter) ([]domain.Exchange, error) {
	query := `SELECT ` + exchangeColumns + ` FROM exchanges
		WHERE (requester_id = $1 OR owner_id = $1)`
	args := []any{userID}
	if filter.Status != "" {
		args = append(args, filter.Status)
		query += " AND status = $2"
	}
	query += " ORDER BY created_at DESC"

	rows, err := store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("liste des échanges : %w", err)
	}
	defer rows.Close()

	exchanges := []domain.Exchange{}
	for rows.Next() {
		exchange, err := scanExchange(rows)
		if err != nil {
			return nil, fmt.Errorf("lecture d'un échange : %w", err)
		}
		exchanges = append(exchanges, exchange)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des échanges : %w", err)
	}
	return exchanges, nil
}

func (store Store) FindExchange(ctx context.Context, exchangeID int) (domain.Exchange, error) {
	row := store.db.QueryRowContext(ctx, `SELECT `+exchangeColumns+` FROM exchanges WHERE id = $1`, exchangeID)
	exchange, err := scanExchange(row)
	if err != nil {
		return domain.Exchange{}, fmt.Errorf("lecture de l'échange : %w", err)
	}
	return exchange, nil
}

func (store Store) UpdateExchangeStatus(
	ctx context.Context,
	exchangeID int,
	expectedStatus string,
	newStatus string,
	creditChanges []domain.CreditChange,
) (domain.Exchange, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Exchange{}, fmt.Errorf("début de la transaction d'échange : %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRowContext(ctx, `SELECT `+exchangeColumns+` FROM exchanges WHERE id = $1 FOR UPDATE`, exchangeID)
	exchange, err := scanExchange(row)
	if err != nil {
		return domain.Exchange{}, fmt.Errorf("verrouillage de l'échange : %w", err)
	}
	if exchange.Status != expectedStatus {
		return domain.Exchange{}, domain.ErrExchangeTransition
	}

	changes := append([]domain.CreditChange(nil), creditChanges...)
	sort.Slice(changes, func(left, right int) bool {
		return changes[left].UserID < changes[right].UserID
	})
	for _, change := range changes {
		var lockedUserID int
		if err := tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, change.UserID).Scan(&lockedUserID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.Exchange{}, domain.ErrUserNotFound
			}
			return domain.Exchange{}, fmt.Errorf("verrouillage du solde : %w", err)
		}

		if change.Montant < 0 {
			var balance int
			if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(montant), 0)
				FROM credit_transactions WHERE user_id = $1`, change.UserID).Scan(&balance); err != nil {
				return domain.Exchange{}, fmt.Errorf("vérification du solde : %w", err)
			}
			if balance+change.Montant < 0 {
				return domain.Exchange{}, domain.ErrExchangeInsufficientFund
			}
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO credit_transactions (
			user_id, exchange_id, montant, type
		) VALUES ($1, $2, $3, $4)`, change.UserID, exchangeID, change.Montant, change.Type); err != nil {
			return domain.Exchange{}, fmt.Errorf("écriture de la transaction de crédit : %w", err)
		}
	}

	row = tx.QueryRowContext(ctx, `UPDATE exchanges
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING `+exchangeColumns, exchangeID, newStatus)
	exchange, err = scanExchange(row)
	if err != nil {
		return domain.Exchange{}, fmt.Errorf("modification de l'échange : %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.Exchange{}, fmt.Errorf("validation de l'échange : %w", err)
	}
	return exchange, nil
}
