package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"TaskFlow/internal/domain"

	"github.com/google/uuid"
)

type TaskRepo struct{ db *sql.DB }

func NewTaskRepo(db *sql.DB) *TaskRepo { return &TaskRepo{db: db} }

func (r *TaskRepo) Create(ctx context.Context, userID, projectID, title string) (domain.Task, error) {
	title = strings.TrimSpace(title)

	t := domain.Task{
		ID:        uuid.NewString(),
		ProjectID: projectID,
		Title:     title,
		Completed: false,
	}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO tasks (id, project_id, title)
		SELECT $1, p.id, $2
		FROM projects p
		WHERE p.id = $3 AND p.user_id = $4
		RETURNING created_at, updated_at, completed
	`, t.ID, t.Title, projectID, userID).Scan(&t.CreatedAt, &t.UpdatedAt, &t.Completed)

	return t, err
}

func (r *TaskRepo) List(
	ctx context.Context,
	userID string,
	projectID string,
	completed *bool,
	limit int,
	cursor *domain.Cursor,
) ([]domain.Task, *domain.Cursor, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	fetch := limit + 1

	var b strings.Builder
	args := make([]any, 0, 8)

	arg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	b.WriteString(
		"SELECT t.id, t.project_id, t.title, t.completed, t.created_at, t.updated_at " +
			"FROM tasks t " +
			"JOIN projects p ON p.id = t.project_id " +
			"WHERE p.user_id = ",
	)
	b.WriteString(arg(userID))
	b.WriteString(" AND t.project_id = ")
	b.WriteString(arg(projectID))

	if completed != nil {
		b.WriteString(" AND t.completed = ")
		b.WriteString(arg(*completed))
	}

	if cursor != nil {
		b.WriteString(" AND (t.created_at, t.id) < (")
		b.WriteString(arg(cursor.CreatedAt))
		b.WriteString(", ")
		b.WriteString(arg(cursor.ID))
		b.WriteString(")")
	}

	b.WriteString(" ORDER BY t.created_at DESC, t.id DESC LIMIT ")
	b.WriteString(arg(fetch))

	rows, err := r.db.QueryContext(ctx, b.String(), args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, nil, err
		}
		out = append(out, t)
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

func (r *TaskRepo) Get(ctx context.Context, userID, taskID string) (domain.Task, error) {
	var t domain.Task
	err := r.db.QueryRowContext(ctx, `
		SELECT t.id, t.project_id, t.title, t.completed, t.created_at, t.updated_at
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE t.id = $1 AND p.user_id = $2
	`, taskID, userID).Scan(&t.ID, &t.ProjectID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *TaskRepo) Update(ctx context.Context, userID, taskID string, title *string, completed *bool) (domain.Task, error) {
	var t domain.Task
	err := r.db.QueryRowContext(ctx, `
		UPDATE tasks t
		SET
			title = COALESCE($3, t.title),
			completed = COALESCE($4, t.completed),
			updated_at = now()
		FROM projects p
		WHERE p.id = t.project_id
		  AND p.user_id = $2
		  AND t.id = $1
		RETURNING t.id, t.project_id, t.title, t.completed, t.created_at, t.updated_at
	`, taskID, userID, title, completed).Scan(&t.ID, &t.ProjectID, &t.Title, &t.Completed, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *TaskRepo) Delete(ctx context.Context, userID, taskID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM tasks t
		USING projects p
		WHERE p.id = t.project_id
		  AND p.user_id = $2
		  AND t.id = $1
	`, taskID, userID)
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
