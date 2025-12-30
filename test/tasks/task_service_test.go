package tasks

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"TaskFlow/internal/domain"
	_service "TaskFlow/internal/service"
)

type fakeTaskRepo struct {
	createFn func(ctx context.Context, userID, projectID, title string) (domain.Task, error)
	listFn   func(ctx context.Context, userID, projectID string, completed *bool, limit int, cursor *domain.Cursor) ([]domain.Task, *domain.Cursor, error)
	getFn    func(ctx context.Context, userID, taskID string) (domain.Task, error)
	updateFn func(ctx context.Context, userID, taskID string, title *string, completed *bool) (domain.Task, error)
	deleteFn func(ctx context.Context, userID, taskID string) error

	lastListLimit int
}

func (f *fakeTaskRepo) Create(ctx context.Context, userID, projectID, title string) (domain.Task, error) {
	if f.createFn != nil {
		return f.createFn(ctx, userID, projectID, title)
	}
	return domain.Task{}, nil
}

func (f *fakeTaskRepo) List(ctx context.Context, userID, projectID string, completed *bool, limit int, cursor *domain.Cursor) ([]domain.Task, *domain.Cursor, error) {
	f.lastListLimit = limit
	if f.listFn != nil {
		return f.listFn(ctx, userID, projectID, completed, limit, cursor)
	}
	return nil, nil, nil
}

func (f *fakeTaskRepo) Get(ctx context.Context, userID, taskID string) (domain.Task, error) {
	if f.getFn != nil {
		return f.getFn(ctx, userID, taskID)
	}
	return domain.Task{}, nil
}

func (f *fakeTaskRepo) Update(ctx context.Context, userID, taskID string, title *string, completed *bool) (domain.Task, error) {
	if f.updateFn != nil {
		return f.updateFn(ctx, userID, taskID, title, completed)
	}
	return domain.Task{}, nil
}

func (f *fakeTaskRepo) Delete(ctx context.Context, userID, taskID string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, userID, taskID)
	}
	return nil
}

func TestTaskService_Create_RejectsEmptyTitle(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := _service.NewTaskService(repo)

	_, err := svc.Create(context.Background(), "user-1", "proj-1", "   ")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "title required" {
		t.Fatalf("expected 'title required', got %v", err)
	}
}

func TestTaskService_Create_MapsSqlNoRows_ToErrNotFound(t *testing.T) {
	repo := &fakeTaskRepo{
		createFn: func(ctx context.Context, userID, projectID, title string) (domain.Task, error) {
			return domain.Task{}, sql.ErrNoRows
		},
	}
	svc := _service.NewTaskService(repo)

	_, err := svc.Create(context.Background(), "user-1", "proj-1", "hello")
	if !errors.Is(err, _service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_List_ClampsLimit_Defaults20_AndMax100(t *testing.T) {
	repo := &fakeTaskRepo{
		listFn: func(ctx context.Context, userID, projectID string, completed *bool, limit int, cursor *domain.Cursor) ([]domain.Task, *domain.Cursor, error) {
			return []domain.Task{}, nil, nil
		},
	}
	svc := _service.NewTaskService(repo)

	_, err := svc.List(context.Background(), "user-1", "proj-1", nil, 0, nil)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if repo.lastListLimit != 20 {
		t.Fatalf("expected limit=20, got %d", repo.lastListLimit)
	}

	_, err = svc.List(context.Background(), "user-1", "proj-1", nil, 999, nil)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if repo.lastListLimit != 100 {
		t.Fatalf("expected limit=100, got %d", repo.lastListLimit)
	}
}

func TestTaskService_Update_RejectsEmptyTitleWhenProvided(t *testing.T) {
	repo := &fakeTaskRepo{}
	svc := _service.NewTaskService(repo)

	empty := "   "
	_, err := svc.Update(context.Background(), "user-1", "task-1", &empty, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "title cannot be empty" {
		t.Fatalf("expected 'title cannot be empty', got %v", err)
	}
}

func TestTaskService_Get_MapsSqlNoRows_ToErrNotFound(t *testing.T) {
	repo := &fakeTaskRepo{
		getFn: func(ctx context.Context, userID, taskID string) (domain.Task, error) {
			return domain.Task{}, sql.ErrNoRows
		},
	}
	svc := _service.NewTaskService(repo)

	_, err := svc.Get(context.Background(), "user-1", "task-1")
	if !errors.Is(err, _service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_Delete_MapsSqlNoRows_ToErrNotFound(t *testing.T) {
	repo := &fakeTaskRepo{
		deleteFn: func(ctx context.Context, userID, taskID string) error {
			return sql.ErrNoRows
		},
	}
	svc := _service.NewTaskService(repo)

	err := svc.Delete(context.Background(), "user-1", "task-1")
	if !errors.Is(err, _service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTaskService_List_ReturnsNextCursor(t *testing.T) {
	// This test ensures the service just wraps repo output correctly.
	now := time.Now().UTC()
	next := &domain.Cursor{CreatedAt: now, ID: "x"}

	repo := &fakeTaskRepo{
		listFn: func(ctx context.Context, userID, projectID string, completed *bool, limit int, cursor *domain.Cursor) ([]domain.Task, *domain.Cursor, error) {
			return []domain.Task{{ID: "t1", ProjectID: "p1", Title: "a"}}, next, nil
		},
	}
	svc := _service.NewTaskService(repo)

	page, err := svc.List(context.Background(), "user-1", "p1", nil, 10, nil)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "t1" {
		t.Fatalf("unexpected items: %#v", page.Items)
	}
	if page.NextCursor == nil || page.NextCursor.ID != "x" {
		t.Fatalf("expected next cursor, got %#v", page.NextCursor)
	}
}
