package postgres

import (
	"context"
	"database/sql"

	"TaskFlow/internal/domain"

	"github.com/google/uuid"
)

type ProjectRepo struct{ db *sql.DB }

func NewProjectRepo(db *sql.DB) *ProjectRepo { return &ProjectRepo{db: db} }

func (r *ProjectRepo) Create(ctx context.Context, userID, name string) (domain.Project, error) {
	p := domain.Project{
		ID:     uuid.NewString(),
		UserID: userID,
		Name:   name,
	}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO projects (id, user_id, name)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`, p.ID, p.UserID, p.Name).Scan(&p.CreatedAt, &p.UpdatedAt)

	return p, err
}

func (r *ProjectRepo) List(ctx context.Context, userID string, limit int, cursor *domain.Cursor) ([]domain.Project, *domain.Cursor, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	fetch := limit + 1

	var (
		rows *sql.Rows
		err  error
	)

	if cursor == nil {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, user_id, name, created_at, updated_at
			FROM projects
			WHERE user_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`, userID, fetch)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT id, user_id, name, created_at, updated_at
			FROM projects
			WHERE user_id = $1
			  AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`, userID, cursor.CreatedAt, cursor.ID, fetch)
	}

	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var out []domain.Project
	for rows.Next() {
		var p domain.Project
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var next *domain.Cursor
	if len(out) > limit {
		last := out[limit-1]
		next = &domain.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
		out = out[:limit]
	}

	return out, next, nil
}

func (r *ProjectRepo) Get(ctx context.Context, userID, projectID string) (domain.Project, error) {
	var p domain.Project
	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, created_at, updated_at
		FROM projects
		WHERE user_id = $1 AND id = $2
	`, userID, projectID).Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *ProjectRepo) UpdateName(ctx context.Context, userID, projectID, name string) (domain.Project, error) {
	var p domain.Project
	err := r.db.QueryRowContext(ctx, `
		UPDATE projects
		SET name = $3, updated_at = now()
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, name, created_at, updated_at
	`, userID, projectID, name).Scan(&p.ID, &p.UserID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *ProjectRepo) Delete(ctx context.Context, userID, projectID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM projects
		WHERE user_id = $1 AND id = $2
	`, userID, projectID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
