package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"studybuddy/backend/services/auth/domain"
	"studybuddy/backend/services/auth/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgUserRepository implements UserRepository using PostgreSQL.
type PgUserRepository struct {
	pool *pgxpool.Pool
}

// NewPgUserRepository creates a new PgUserRepository.
func NewPgUserRepository(pool *pgxpool.Pool) usecase.UserRepository {
	return &PgUserRepository{pool: pool}
}

func (r *PgUserRepository) Create(ctx context.Context, user *domain.User) error {
	// Let DB generate id and timestamps.
	const q = `
INSERT INTO users (email, password_hash, first_name, last_name, is_active)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at;
`
	err := r.pool.QueryRow(ctx, q,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailExists
		}
		return err
	}
	return nil
}

func (r *PgUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, first_name, last_name, is_active, created_at, updated_at
FROM users
WHERE email = $1;
`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *PgUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, first_name, last_name, is_active, created_at, updated_at
FROM users
WHERE id = $1;
`
	var u domain.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}
