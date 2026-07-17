package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

const serviceColumns = `id, provider_id, titre, description, categorie,
       duree_minutes, credits, ville, actif, created_at`

type serviceRow interface {
	Scan(...any) error
}

func scanService(row serviceRow) (domain.Service, error) {
	var service domain.Service
	var createdAt time.Time
	err := row.Scan(
		&service.ID,
		&service.ProviderID,
		&service.Titre,
		&service.Description,
		&service.Categorie,
		&service.DureeMinutes,
		&service.Credits,
		&service.Ville,
		&service.Actif,
		&createdAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Service{}, domain.ErrServiceNotFound
	}
	if err != nil {
		return domain.Service{}, err
	}
	service.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	return service, nil
}

func (store Store) CreateService(ctx context.Context, providerID int, input domain.CreateServiceInput) (domain.Service, error) {
	row := store.db.QueryRowContext(ctx, `INSERT INTO services (
		provider_id, titre, description, categorie, duree_minutes, credits, ville
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING `+serviceColumns,
		providerID,
		input.Titre,
		input.Description,
		input.Categorie,
		input.DureeMinutes,
		input.Credits,
		input.Ville,
	)
	service, err := scanService(row)
	if err != nil {
		return domain.Service{}, fmt.Errorf("création du service : %w", err)
	}
	return service, nil
}

func (store Store) ListServices(ctx context.Context, filter domain.ServiceFilter) ([]domain.Service, error) {
	rows, err := store.db.QueryContext(ctx, `SELECT `+serviceColumns+`
		FROM services
		WHERE actif = TRUE
			AND ($1 = '' OR categorie = $1)
			AND ($2 = '' OR ville = $2)
			AND ($3 = '' OR titre ILIKE '%' || $3 || '%' OR description ILIKE '%' || $3 || '%')
		ORDER BY created_at DESC`, filter.Categorie, filter.Ville, filter.Search)
	if err != nil {
		return nil, fmt.Errorf("liste des services : %w", err)
	}
	defer rows.Close()

	services := []domain.Service{}
	for rows.Next() {
		service, err := scanService(rows)
		if err != nil {
			return nil, fmt.Errorf("lecture d'un service : %w", err)
		}
		services = append(services, service)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des services : %w", err)
	}
	return services, nil
}

func (store Store) FindService(ctx context.Context, id int) (domain.Service, error) {
	row := store.db.QueryRowContext(ctx, `SELECT `+serviceColumns+` FROM services WHERE id = $1 AND actif = TRUE`, id)
	service, err := scanService(row)
	if err != nil {
		return domain.Service{}, fmt.Errorf("lecture du service : %w", err)
	}
	return service, nil
}

func (store Store) UpdateService(ctx context.Context, id int, input domain.CreateServiceInput) (domain.Service, error) {
	row := store.db.QueryRowContext(ctx, `UPDATE services
		SET titre = $2, description = $3, categorie = $4,
			duree_minutes = $5, credits = $6, ville = $7
		WHERE id = $1 AND actif = TRUE
		RETURNING `+serviceColumns,
		id,
		input.Titre,
		input.Description,
		input.Categorie,
		input.DureeMinutes,
		input.Credits,
		input.Ville,
	)
	service, err := scanService(row)
	if err != nil {
		return domain.Service{}, fmt.Errorf("modification du service : %w", err)
	}
	return service, nil
}

func (store Store) DeactivateService(ctx context.Context, id int) error {
	result, err := store.db.ExecContext(ctx, `UPDATE services SET actif = FALSE WHERE id = $1 AND actif = TRUE`, id)
	if err != nil {
		return fmt.Errorf("suppression du service : %w", err)
	}
	updatedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("vérification de la suppression : %w", err)
	}
	if updatedRows == 0 {
		return domain.ErrServiceNotFound
	}
	return nil
}
