package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func scanErr(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func (s *Store) CreateUser(ctx context.Context, u User) (User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return u, err
	}
	defer tx.Rollback()
	err = tx.QueryRowContext(
		ctx,
		`INSERT INTO users(pseudo,bio,ville) VALUES($1,$2,$3) RETURNING id,created_at`,
		u.Pseudo,
		u.Bio,
		u.Ville,
	).Scan(&u.ID, &u.CreatedAt)
	if isUniqueViolation(err) {
		return u, fmt.Errorf("%w: pseudo déjà utilisé", ErrConflict)
	}
	if err != nil {
		return u, err
	}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO credit_transactions(user_id,montant,type) VALUES($1,10,'welcome')`,
		u.ID,
	)
	if err != nil {
		return u, err
	}
	if err = tx.Commit(); err != nil {
		return u, err
	}
	u.CreditBalance = 10
	return u, nil
}

func (s *Store) User(ctx context.Context, id int64) (User, error) {
	var u User
	const query = `
		SELECT u.id, u.pseudo, u.bio, u.ville, COALESCE(sum(c.montant), 0), u.created_at
		FROM users u
		LEFT JOIN credit_transactions c ON c.user_id = u.id
		WHERE u.id = $1
		GROUP BY u.id`
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Pseudo,
		&u.Bio,
		&u.Ville,
		&u.CreditBalance,
		&u.CreatedAt,
	)
	if err != nil {
		return u, scanErr(err)
	}
	u.Skills, err = s.Skills(ctx, id)
	return u, err
}

func (s *Store) UpdateUser(ctx context.Context, u User) (User, error) {
	const query = `UPDATE users SET pseudo=$2, bio=$3, ville=$4 WHERE id=$1 RETURNING created_at`
	err := s.db.QueryRowContext(ctx, query, u.ID, u.Pseudo, u.Bio, u.Ville).Scan(&u.CreatedAt)
	if isUniqueViolation(err) {
		return u, fmt.Errorf("%w: pseudo déjà utilisé", ErrConflict)
	}
	if err != nil {
		return u, scanErr(err)
	}
	return s.User(ctx, u.ID)
}

func (s *Store) Skills(ctx context.Context, id int64) ([]Skill, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, id).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}
	rows, err := s.db.QueryContext(ctx, `SELECT nom,niveau FROM skills WHERE user_id=$1 ORDER BY nom`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Skill{}
	for rows.Next() {
		var v Skill
		if err = rows.Scan(&v.Nom, &v.Niveau); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) ReplaceSkills(ctx context.Context, id int64, skills []Skill) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	r, e := tx.ExecContext(ctx, `DELETE FROM skills WHERE user_id=$1`, id)
	if e != nil {
		return e
	}
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		var ok bool
		e = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, id).Scan(&ok)
		if e != nil {
			return e
		}
		if !ok {
			return ErrNotFound
		}
	}
	for _, skill := range skills {
		_, e = tx.ExecContext(
			ctx,
			`INSERT INTO skills(user_id,nom,niveau) VALUES($1,$2,$3)`,
			id,
			skill.Nom,
			skill.Niveau,
		)
		if e != nil {
			return e
		}
	}
	return tx.Commit()
}

func (s *Store) HasSkill(ctx context.Context, id int64, name string) (bool, error) {
	var ok bool
	const query = `SELECT EXISTS(SELECT 1 FROM skills WHERE user_id=$1 AND lower(nom)=lower($2))`
	err := s.db.QueryRowContext(ctx, query, id, name).Scan(&ok)
	return ok, err
}

const serviceCols = `id,provider_id,titre,description,categorie,duree_minutes,credits,ville,actif,created_at`

func scanService(row interface{ Scan(...any) error }) (Service, error) {
	var service Service
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
		&service.CreatedAt,
	)
	return service, scanErr(err)
}

func (s *Store) CreateService(ctx context.Context, v Service) (Service, error) {
	const query = `
		INSERT INTO services(provider_id,titre,description,categorie,duree_minutes,credits,ville)
		VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING `
	row := s.db.QueryRowContext(
		ctx, query+serviceCols, v.ProviderID, v.Titre, v.Description,
		v.Categorie, v.DureeMinutes, v.Credits, v.Ville,
	)
	return scanService(row)
}

func (s *Store) Service(ctx context.Context, id int64) (Service, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+serviceCols+` FROM services WHERE id=$1`, id)
	return scanService(row)
}

func (s *Store) UpdateService(ctx context.Context, v Service) (Service, error) {
	const query = `
		UPDATE services
		SET titre=$2, description=$3, categorie=$4, duree_minutes=$5,
			credits=$6, ville=$7, actif=$8
		WHERE id=$1 RETURNING `
	row := s.db.QueryRowContext(
		ctx, query+serviceCols, v.ID, v.Titre, v.Description,
		v.Categorie, v.DureeMinutes, v.Credits, v.Ville, v.Actif,
	)
	return scanService(row)
}

func (s *Store) DeleteService(ctx context.Context, id int64) error {
	r, e := s.db.ExecContext(ctx, `UPDATE services SET actif=false WHERE id=$1`, id)
	if e != nil {
		return e
	}
	n, e := r.RowsAffected()
	if e != nil {
		return e
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Services(ctx context.Context, f ServiceFilter) ([]Service, error) {
	q := `SELECT ` + serviceCols + ` FROM services WHERE actif=true`
	args := []any{}
	add := func(col, val string) {
		if val != "" {
			args = append(args, val)
			q += fmt.Sprintf(` AND %s=$%d`, col, len(args))
		}
	}
	add("categorie", f.Categorie)
	add("ville", f.Ville)
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += fmt.Sprintf(` AND (titre ILIKE $%d OR description ILIKE $%d)`, len(args), len(args))
	}
	q += ` ORDER BY created_at DESC`
	rows, e := s.db.QueryContext(ctx, q, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Service{}
	for rows.Next() {
		v, e := scanService(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

const exchangeCols = `id,service_id,requester_id,owner_id,status,created_at,updated_at`

func scanExchange(row interface{ Scan(...any) error }) (v Exchange, e error) {
	e = row.Scan(
		&v.ID,
		&v.ServiceID,
		&v.RequesterID,
		&v.OwnerID,
		&v.Status,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	e = scanErr(e)
	return
}

func (s *Store) CreateExchange(ctx context.Context, serviceID, requesterID int64) (Exchange, error) {
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if e != nil {
		return Exchange{}, e
	}
	defer tx.Rollback()
	var owner int64
	var credits int
	var active bool
	e = tx.QueryRowContext(
		ctx,
		`SELECT provider_id,credits,actif FROM services WHERE id=$1 FOR UPDATE`,
		serviceID,
	).Scan(&owner, &credits, &active)
	if e != nil {
		return Exchange{}, scanErr(e)
	}
	if !active {
		return Exchange{}, fmt.Errorf("%w: service inactif", ErrInvalid)
	}
	if owner == requesterID {
		return Exchange{}, fmt.Errorf("%w: impossible de réserver son propre service", ErrInvalid)
	}
	var balance int
	e = tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(sum(montant),0) FROM credit_transactions WHERE user_id=$1`,
		requesterID,
	).Scan(&balance)
	if e != nil {
		return Exchange{}, e
	}
	if balance < credits {
		return Exchange{}, ErrInsufficientCredits
	}
	const insertExchange = `
		INSERT INTO exchanges(service_id,requester_id,owner_id,status)
		VALUES($1,$2,$3,'pending') RETURNING `
	row := tx.QueryRowContext(ctx, insertExchange+exchangeCols, serviceID, requesterID, owner)
	v, e := scanExchange(row)
	if isUniqueViolation(e) {
		return v, fmt.Errorf("%w: service déjà réservé", ErrConflict)
	}
	if e != nil {
		return v, e
	}
	return v, tx.Commit()
}

func (s *Store) Exchange(ctx context.Context, id int64) (Exchange, error) {
	return scanExchange(s.db.QueryRowContext(ctx, `SELECT `+exchangeCols+` FROM exchanges WHERE id=$1`, id))
}

func (s *Store) Exchanges(ctx context.Context, user int64, status string) ([]Exchange, error) {
	q := `SELECT ` + exchangeCols + ` FROM exchanges WHERE (requester_id=$1 OR owner_id=$1)`
	a := []any{user}
	if status != "" {
		q += ` AND status=$2`
		a = append(a, status)
	}
	q += ` ORDER BY created_at DESC`
	rows, e := s.db.QueryContext(ctx, q, a...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Exchange{}
	for rows.Next() {
		v, e := scanExchange(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) Transition(ctx context.Context, id, user int64, action string) (Exchange, error) {
	tx, e := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if e != nil {
		return Exchange{}, e
	}
	defer tx.Rollback()
	v, e := scanExchange(tx.QueryRowContext(ctx, `SELECT `+exchangeCols+` FROM exchanges WHERE id=$1 FOR UPDATE`, id))
	if e != nil {
		return v, e
	}
	var credits int
	e = tx.QueryRowContext(ctx, `SELECT credits FROM services WHERE id=$1`, v.ServiceID).Scan(&credits)
	if e != nil {
		return v, e
	}
	next := ""
	switch action {
	case "accept":
		if user != v.OwnerID {
			return v, ErrForbidden
		}
		if v.Status != StatusPending {
			return v, ErrInvalid
		}
		var bal int
		e = tx.QueryRowContext(
			ctx,
			`SELECT COALESCE(sum(montant),0) FROM credit_transactions WHERE user_id=$1`,
			v.RequesterID,
		).Scan(&bal)
		if e != nil {
			return v, e
		}
		if bal < credits {
			return v, ErrInsufficientCredits
		}
		_, e = tx.ExecContext(
			ctx,
			`INSERT INTO credit_transactions(user_id,exchange_id,montant,type) VALUES($1,$2,$3,'spend')`,
			v.RequesterID, v.ID, -credits,
		)
		next = StatusAccepted
	case "reject":
		if user != v.OwnerID {
			return v, ErrForbidden
		}
		if v.Status != StatusPending {
			return v, ErrInvalid
		}
		next = StatusRejected
	case "complete":
		if user != v.RequesterID && user != v.OwnerID {
			return v, ErrForbidden
		}
		if v.Status != StatusAccepted {
			return v, ErrInvalid
		}
		_, e = tx.ExecContext(
			ctx,
			`INSERT INTO credit_transactions(user_id,exchange_id,montant,type) VALUES($1,$2,$3,'earn')`,
			v.OwnerID, v.ID, credits,
		)
		next = StatusCompleted
	case "cancel":
		if user != v.RequesterID && user != v.OwnerID {
			return v, ErrForbidden
		}
		if v.Status != StatusPending && v.Status != StatusAccepted {
			return v, ErrInvalid
		}
		if v.Status == StatusAccepted {
			_, e = tx.ExecContext(
				ctx,
				`INSERT INTO credit_transactions(user_id,exchange_id,montant,type) VALUES($1,$2,$3,'refund')`,
				v.RequesterID, v.ID, credits,
			)
		}
		next = StatusCancelled
	default:
		return v, ErrInvalid
	}
	if e != nil {
		return v, e
	}
	row := tx.QueryRowContext(
		ctx,
		`UPDATE exchanges SET status=$2,updated_at=now() WHERE id=$1 RETURNING `+exchangeCols,
		id,
		next,
	)
	v, e = scanExchange(row)
	if e != nil {
		return v, e
	}
	return v, tx.Commit()
}

func (s *Store) CreateReview(ctx context.Context, r Review) (Review, error) {
	var ex Exchange
	var e error
	ex, e = s.Exchange(ctx, r.ExchangeID)
	if e != nil {
		return r, e
	}
	if ex.Status != StatusCompleted {
		return r, fmt.Errorf("%w: échange non terminé", ErrInvalid)
	}
	if r.AuthorID != ex.RequesterID && r.AuthorID != ex.OwnerID {
		return r, ErrForbidden
	}
	if r.AuthorID == ex.RequesterID {
		r.TargetID = ex.OwnerID
	} else {
		r.TargetID = ex.RequesterID
	}
	const query = `
		INSERT INTO reviews(exchange_id,author_id,target_id,note,commentaire)
		VALUES($1,$2,$3,$4,$5) RETURNING id,created_at`
	e = s.db.QueryRowContext(
		ctx, query, r.ExchangeID, r.AuthorID, r.TargetID, r.Note, r.Commentaire,
	).Scan(&r.ID, &r.CreatedAt)
	if isUniqueViolation(e) {
		return r, fmt.Errorf("%w: avis déjà publié", ErrInvalid)
	}
	return r, e
}

func (s *Store) Reviews(ctx context.Context, where string, id int64) ([]Review, error) {
	allowed := map[string]string{"user": "target_id", "service": "e.service_id"}
	col, ok := allowed[where]
	if !ok {
		return nil, ErrInvalid
	}
	q := `
		SELECT r.id,r.exchange_id,r.author_id,r.target_id,r.note,r.commentaire,r.created_at
		FROM reviews r
		JOIN exchanges e ON e.id=r.exchange_id
		WHERE ` + col + `=$1
		ORDER BY r.created_at DESC`
	rows, e := s.db.QueryContext(ctx, q, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []Review{}
	for rows.Next() {
		var r Review
		e = rows.Scan(
			&r.ID,
			&r.ExchangeID,
			&r.AuthorID,
			&r.TargetID,
			&r.Note,
			&r.Commentaire,
			&r.CreatedAt,
		)
		if e != nil {
			return nil, e
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Stats(ctx context.Context, id int64) (UserStats, error) {
	var x UserStats
	x.UserID = id
	const query = `
		SELECT
			(SELECT count(*) FROM services WHERE provider_id=$1 AND actif),
			(SELECT count(*) FROM exchanges WHERE status='completed' AND (requester_id=$1 OR owner_id=$1)),
			COALESCE((SELECT sum(montant) FROM credit_transactions WHERE user_id=$1),0),
			COALESCE((SELECT avg(note) FROM reviews WHERE target_id=$1),0),
			(SELECT count(*) FROM reviews WHERE target_id=$1),
			COALESCE((SELECT sum(montant) FROM credit_transactions WHERE user_id=$1 AND type='earn'),0),
			COALESCE(-(SELECT sum(montant) FROM credit_transactions WHERE user_id=$1 AND type='spend'),0)`
	e := s.db.QueryRowContext(ctx, query, id).Scan(
		&x.ServicesActifs,
		&x.EchangesCompletes,
		&x.CreditBalance,
		&x.NoteMoyenne,
		&x.NbAvis,
		&x.TotalGagne,
		&x.TotalDepense,
	)
	if e != nil {
		return x, e
	}
	var exists bool
	e = s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, id).Scan(&exists)
	if e == nil && !exists {
		e = ErrNotFound
	}
	return x, e
}

func validStatus(v string) bool {
	return v == StatusPending || v == StatusAccepted || v == StatusRejected || v == StatusCancelled || v == StatusCompleted
}

func clean(v string) string {
	return strings.TrimSpace(v)
}
