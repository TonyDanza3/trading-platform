package internal

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound  = errors.New("instrument not found")
	ErrDuplicate = errors.New("instrument already exists")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, symbol, name, assetClass string) (Instrument, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO instruments (symbol, name, asset_class)
		VALUES ($1, $2, $3)
		RETURNING id, symbol, name, asset_class, created_at, updated_at
	`, symbol, name, assetClass)
	item, err := scanInstrument(row)
	if isUniqueViolation(err) {
		return Instrument{}, ErrDuplicate
	}
	return item, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Instrument, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, symbol, name, asset_class, created_at, updated_at
		FROM instruments WHERE id = $1
	`, id)
	item, err := scanInstrument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Instrument{}, ErrNotFound
	}
	return item, err
}

func (r *Repository) List(ctx context.Context, page, pageSize int) ([]Instrument, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT count(*) FROM instruments`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, `
		SELECT id, symbol, name, asset_class, created_at, updated_at
		FROM instruments
		ORDER BY symbol ASC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]Instrument, 0)
	for rows.Next() {
		item, err := scanInstrument(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, symbol, name, assetClass string) (Instrument, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE instruments
		SET symbol = $2, name = $3, asset_class = $4, updated_at = now()
		WHERE id = $1
		RETURNING id, symbol, name, asset_class, created_at, updated_at
	`, id, symbol, name, assetClass)
	item, err := scanInstrument(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Instrument{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return Instrument{}, ErrDuplicate
	}
	return item, err
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM instruments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanInstrument(row rowScanner) (Instrument, error) {
	var item Instrument
	err := row.Scan(&item.ID, &item.Symbol, &item.Name, &item.AssetClass, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
