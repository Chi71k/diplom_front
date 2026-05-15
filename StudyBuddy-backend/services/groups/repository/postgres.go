package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"studybuddy/backend/services/groups/domain"
	"studybuddy/backend/services/groups/usecase"
)

type PgGroupRepository struct {
	pool *pgxpool.Pool
}

func NewPgGroupRepository(pool *pgxpool.Pool) usecase.GroupRepository {
	return &PgGroupRepository{pool: pool}
}

func (r *PgGroupRepository) Create(ctx context.Context, name, description, ownerID string, courseIDs []string) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const insG = `
INSERT INTO study_groups (name, description, owner_id)
VALUES ($1, $2, $3)
RETURNING id, created_at;
`
	var g domain.Group
	err = tx.QueryRow(ctx, insG, name, description, ownerID).Scan(&g.ID, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	g.Name = name
	g.Description = description
	g.OwnerID = ownerID
	g.CourseIDs = append([]string(nil), courseIDs...)

	const insM = `
INSERT INTO group_members (group_id, user_id, role)
VALUES ($1, $2, 'owner');
`
	if _, err := tx.Exec(ctx, insM, g.ID, ownerID); err != nil {
		return nil, err
	}

	const insC = `INSERT INTO group_courses (group_id, course_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`
	for _, cid := range courseIDs {
		if _, err := tx.Exec(ctx, insC, g.ID, cid); err != nil {
			return nil, err
		}
	}

	g.Members = []domain.Member{{
		UserID:   ownerID,
		Role:     domain.RoleOwner,
		JoinedAt: g.CreatedAt,
	}}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *PgGroupRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
SELECT id::text, name, description, owner_id::text, created_at
FROM study_groups
WHERE id = $1;
`
	var g domain.Group
	err := r.pool.QueryRow(ctx, q, id).Scan(&g.ID, &g.Name, &g.Description, &g.OwnerID, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
SELECT user_id::text, role, joined_at
FROM group_members
WHERE group_id = $1
ORDER BY joined_at ASC;
`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m domain.Member
		var role string
		if err := rows.Scan(&m.UserID, &role, &m.JoinedAt); err != nil {
			return nil, err
		}
		m.Role = domain.MemberRole(role)
		g.Members = append(g.Members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	crows, err := r.pool.Query(ctx, `SELECT course_id::text FROM group_courses WHERE group_id = $1 ORDER BY course_id`, id)
	if err != nil {
		return nil, err
	}
	defer crows.Close()
	for crows.Next() {
		var cid string
		if err := crows.Scan(&cid); err != nil {
			return nil, err
		}
		g.CourseIDs = append(g.CourseIDs, cid)
	}
	return &g, crows.Err()
}

func (r *PgGroupRepository) ListForUser(ctx context.Context, userID string) ([]domain.Group, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	const q = `
SELECT g.id::text
FROM study_groups g
JOIN group_members gm ON gm.group_id = g.id
WHERE gm.user_id = $1
ORDER BY g.created_at DESC;
`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var out []domain.Group
	for _, id := range ids {
		g, err := r.GetByID(ctx, id)
		if err != nil || g == nil {
			continue
		}
		out = append(out, *g)
	}
	return out, nil
}

func (r *PgGroupRepository) UpdateMetadataAndCourses(ctx context.Context, groupID string, name, description *string, courseIDs *[]string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if name != nil || description != nil {
		var tag pgconn.CommandTag
		switch {
		case name != nil && description != nil:
			tag, err = tx.Exec(ctx, `UPDATE study_groups SET name = $2, description = $3 WHERE id = $1`, groupID, *name, *description)
		case name != nil:
			tag, err = tx.Exec(ctx, `UPDATE study_groups SET name = $2 WHERE id = $1`, groupID, *name)
		default:
			tag, err = tx.Exec(ctx, `UPDATE study_groups SET description = $2 WHERE id = $1`, groupID, *description)
		}
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domain.ErrGroupNotFound
		}
	}

	if courseIDs != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM group_courses WHERE group_id = $1`, groupID); err != nil {
			return err
		}
		const ins = `INSERT INTO group_courses (group_id, course_id) VALUES ($1, $2)`
		for _, cid := range *courseIDs {
			if _, err := tx.Exec(ctx, ins, groupID, cid); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *PgGroupRepository) Delete(ctx context.Context, groupID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := r.pool.Exec(ctx, `DELETE FROM study_groups WHERE id = $1`, groupID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

func (r *PgGroupRepository) AddMember(ctx context.Context, groupID, userID string, role domain.MemberRole) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const q = `
INSERT INTO group_members (group_id, user_id, role)
VALUES ($1, $2, $3);
`
	_, err := r.pool.Exec(ctx, q, groupID, userID, string(role))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrAlreadyMember
		}
		return err
	}
	return nil
}

func (r *PgGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tag, err := r.pool.Exec(ctx, `
DELETE FROM group_members WHERE group_id = $1 AND user_id = $2;
`, groupID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotMember
	}
	return nil
}

func (r *PgGroupRepository) CountMembers(ctx context.Context, groupID string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*)::int FROM group_members WHERE group_id = $1`, groupID).Scan(&n)
	return n, err
}

func (r *PgGroupRepository) ListCourseTitlesForGroup(ctx context.Context, groupID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
SELECT c.title
FROM group_courses gc
JOIN courses c ON c.id = gc.course_id
WHERE gc.group_id = $1
ORDER BY c.title;
`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list course titles: %w", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		titles = append(titles, t)
	}
	return titles, rows.Err()
}

func (r *PgGroupRepository) ListDistinctInterestNamesForGroup(ctx context.Context, groupID string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx, `
SELECT DISTINCT i.name
FROM group_members gm
JOIN user_interests ui ON ui.user_id = gm.user_id
JOIN interests i ON i.id = ui.interest_id
WHERE gm.group_id = $1
ORDER BY i.name;
`, groupID)
	if err != nil {
		return nil, fmt.Errorf("list interest names: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, rows.Err()
}

func (r *PgGroupRepository) ListCandidateUserIDs(ctx context.Context, groupID string, limit int) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 200
	}

	rows, err := r.pool.Query(ctx, `
SELECT u.id::text
FROM users u
WHERE u.is_active = true
  AND u.id NOT IN (SELECT user_id FROM group_members WHERE group_id = $1::uuid)
ORDER BY u.created_at DESC
LIMIT $2;
`, groupID, limit)
	if err != nil {
		return nil, fmt.Errorf("list candidates: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PgGroupRepository) ListCourseOverlapCandidates(ctx context.Context, groupID string, limit int) ([]usecase.OverlapCandidate, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if limit <= 0 {
		limit = 10
	}

	rows, err := r.pool.Query(ctx, `
SELECT s.user_id::text, s.cnt::int
FROM (
  SELECT u.id AS user_id, COUNT(*)::int AS cnt
  FROM users u
  JOIN courses c ON c.owner_user_id = u.id
  JOIN group_courses gc ON gc.course_id = c.id AND gc.group_id = $1::uuid
  WHERE u.is_active = true
    AND u.id NOT IN (SELECT user_id FROM group_members WHERE group_id = $1::uuid)
  GROUP BY u.id
) s
ORDER BY s.cnt DESC, s.user_id ASC
LIMIT $2;
`, groupID, limit)
	if err != nil {
		return nil, fmt.Errorf("overlap candidates: %w", err)
	}
	defer rows.Close()

	var out []usecase.OverlapCandidate
	for rows.Next() {
		var o usecase.OverlapCandidate
		if err := rows.Scan(&o.UserID, &o.OverlapCount); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *PgGroupRepository) ListProfiles(ctx context.Context, userIDs []string) ([]usecase.ProfileSnippet, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(userIDs) == 0 {
		return nil, nil
	}

	uuids := make([]uuid.UUID, 0, len(userIDs))
	for _, s := range userIDs {
		u, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		uuids = append(uuids, u)
	}
	if len(uuids) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, `
SELECT id::text, first_name, last_name, COALESCE(avatar_url, '')
FROM users
WHERE id = ANY($1)
`, uuids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []usecase.ProfileSnippet
	for rows.Next() {
		var p usecase.ProfileSnippet
		if err := rows.Scan(&p.UserID, &p.FirstName, &p.LastName, &p.AvatarURL); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
