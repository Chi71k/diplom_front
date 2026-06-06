package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"studybuddy/backend/services/users/domain"
	"studybuddy/backend/services/users/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgProfileRepository implements ProfileRepository using PostgreSQL users table.
type PgProfileRepository struct {
	pool *pgxpool.Pool
}

// NewPgProfileRepository creates a new PgProfileRepository.
func NewPgProfileRepository(pool *pgxpool.Pool) *PgProfileRepository {
	return &PgProfileRepository{pool: pool}
}

// NewPgAdminRepository returns the same store for admin operations.
func NewPgAdminRepository(pool *pgxpool.Pool) usecase.AdminRepository {
	return &PgProfileRepository{pool: pool}
}

func (r *PgProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT id, email, first_name, last_name, bio, avatar_url,
       COALESCE(telegram_tag, ''), created_at, updated_at
FROM users
WHERE id = $1;
`
	var p domain.Profile
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&p.UserID,
		&p.Email,
		&p.FirstName,
		&p.LastName,
		&p.Bio,
		&p.AvatarURL,
		&p.TelegramTag,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PgProfileRepository) Upsert(ctx context.Context, profile *domain.Profile) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Users service must not create new users; it only updates existing profiles.
	// We only update profile fields and leave credentials (password_hash) untouched.
	const q = `
UPDATE users
SET first_name = $2,
    last_name  = $3,
    bio        = $4,
    avatar_url = $5,
    telegram_tag = $6,
    updated_at = now()
WHERE id = $1;
`
	_, err := r.pool.Exec(ctx, q,
		profile.UserID,
		profile.FirstName,
		profile.LastName,
		profile.Bio,
		profile.AvatarURL,
		profile.TelegramTag,
	)
	return err
}

// DeleteByUserID performs logical deletion: mark user as inactive.
// Auth service will then reject logins for this user.
func (r *PgProfileRepository) DeleteByUserID(ctx context.Context, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
UPDATE users
SET is_active = false,
    updated_at = now()
WHERE id = $1;
`
	_, err := r.pool.Exec(ctx, q, userID)
	return err
}

// ListUsers returns a paginated, filterable list of all users.
func (r *PgProfileRepository) ListUsers(ctx context.Context, filter domain.AdminUserFilter, limit, offset int) ([]domain.UserSummary, int, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var (
		conds  []string
		args   []any
		argPos = 1
	)
	if filter.Role != nil {
		conds = append(conds, fmt.Sprintf("role = $%d", argPos))
		args = append(args, *filter.Role)
		argPos++
	}
	if filter.IsActive != nil {
		conds = append(conds, fmt.Sprintf("is_active = $%d", argPos))
		args = append(args, *filter.IsActive)
		argPos++
	}
	if filter.Search != nil && *filter.Search != "" {
		conds = append(conds, fmt.Sprintf(
			"(first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d OR CONCAT(first_name, ' ', last_name) ILIKE $%d)",
			argPos, argPos, argPos, argPos,
		))
		args = append(args, "%"+*filter.Search+"%")
		argPos++
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	countQ := "SELECT COUNT(*) FROM users " + where
	var total int
	if err := r.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := fmt.Sprintf(`
SELECT id, email, first_name, last_name, role, is_active, created_at
FROM users
%s
ORDER BY created_at DESC
LIMIT $%d OFFSET $%d;
`, where, argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []domain.UserSummary
	for rows.Next() {
		var u domain.UserSummary
		if err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if out == nil {
		out = []domain.UserSummary{}
	}
	return out, total, nil
}

// SetUserActive sets is_active for the given user ID.
func (r *PgProfileRepository) SetUserActive(ctx context.Context, userID string, active bool) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
UPDATE users
SET is_active = $2,
    updated_at = now()
WHERE id = $1;
`
	tag, err := r.pool.Exec(ctx, q, userID, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// GetPlatformStats returns aggregate counts across the platform.
func (r *PgProfileRepository) GetPlatformStats(ctx context.Context) (domain.PlatformStats, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	const q = `
SELECT
    (SELECT COUNT(*)::int FROM users),
    (SELECT COUNT(*)::int FROM users WHERE is_active = true),
    (SELECT COUNT(*)::int FROM matches WHERE status = 'accepted'),
    (SELECT COUNT(*)::int FROM study_groups),
    (SELECT COUNT(*)::int FROM study_sessions WHERE status = 'confirmed');
`
	var s domain.PlatformStats
	err := r.pool.QueryRow(ctx, q).Scan(
		&s.TotalUsers,
		&s.ActiveUsers,
		&s.TotalMatches,
		&s.TotalGroups,
		&s.TotalSessions,
	)
	return s, err
}
