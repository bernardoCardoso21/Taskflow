package projects

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"TaskFlow/internal/domain"
	_service "TaskFlow/internal/service"
)

type fakeProjectRepo struct {
	createFn     func(ctx context.Context, userID, name string) (domain.Project, error)
	listFn       func(ctx context.Context, userID string, limit int, cursor *domain.Cursor) ([]domain.Project, *domain.Cursor, error)
	getFn        func(ctx context.Context, userID, projectID string) (domain.Project, error)
	updateNameFn func(ctx context.Context, userID, projectID, name string) (domain.Project, error)
	deleteFn     func(ctx context.Context, userID, projectID string) error

	lastListLimit  int
	lastListCursor *domain.Cursor
}

func (f *fakeProjectRepo) Create(ctx context.Context, userID, name string) (domain.Project, error) {
	if f.createFn != nil {
		return f.createFn(ctx, userID, name)
	}
	return domain.Project{}, nil
}

func (f *fakeProjectRepo) List(ctx context.Context, userID string, limit int, cursor *domain.Cursor) ([]domain.Project, *domain.Cursor, error) {
	f.lastListLimit = limit
	f.lastListCursor = cursor
	if f.listFn != nil {
		return f.listFn(ctx, userID, limit, cursor)
	}
	return nil, nil, nil
}

func (f *fakeProjectRepo) Get(ctx context.Context, userID, projectID string) (domain.Project, error) {
	if f.getFn != nil {
		return f.getFn(ctx, userID, projectID)
	}
	return domain.Project{}, nil
}

func (f *fakeProjectRepo) UpdateName(ctx context.Context, userID, projectID, name string) (domain.Project, error) {
	if f.updateNameFn != nil {
		return f.updateNameFn(ctx, userID, projectID, name)
	}
	return domain.Project{}, nil
}

func (f *fakeProjectRepo) Delete(ctx context.Context, userID, projectID string) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, userID, projectID)
	}
	return nil
}

func TestProjectService_List_ClampsLimit_DefaultsTo20(t *testing.T) {
	repo := &fakeProjectRepo{
		listFn: func(ctx context.Context, userID string, limit int, cursor *domain.Cursor) ([]domain.Project, *domain.Cursor, error) {
			return []domain.Project{}, nil, nil
		},
	}
	svc := _service.NewProjectService(repo)

	_, err := svc.List(context.Background(), "user-1", 0, nil) // <=0 => default 20
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if repo.lastListLimit != 20 {
		t.Fatalf("expected limit=20, got %d", repo.lastListLimit)
	}
}

func TestProjectService_List_ClampsLimit_Max100(t *testing.T) {
	repo := &fakeProjectRepo{
		listFn: func(ctx context.Context, userID string, limit int, cursor *domain.Cursor) ([]domain.Project, *domain.Cursor, error) {
			return []domain.Project{}, nil, nil
		},
	}
	svc := _service.NewProjectService(repo)

	_, err := svc.List(context.Background(), "user-1", 999, nil) // >100 => 100
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if repo.lastListLimit != 100 {
		t.Fatalf("expected limit=100, got %d", repo.lastListLimit)
	}
}

func TestProjectService_Get_MapsSqlNoRows_ToErrNotFound(t *testing.T) {
	repo := &fakeProjectRepo{
		getFn: func(ctx context.Context, userID, projectID string) (domain.Project, error) {
			return domain.Project{}, sql.ErrNoRows
		},
	}
	svc := _service.NewProjectService(repo)

	_, err := svc.Get(context.Background(), "user-1", "proj-1")
	if !errors.Is(err, _service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProjectService_UpdateName_MapsSqlNoRows_ToErrNotFound(t *testing.T) {
	repo := &fakeProjectRepo{
		updateNameFn: func(ctx context.Context, userID, projectID, name string) (domain.Project, error) {
			return domain.Project{}, sql.ErrNoRows
		},
	}
	svc := _service.NewProjectService(repo)

	_, err := svc.UpdateName(context.Background(), "user-1", "proj-1", "new")
	if !errors.Is(err, _service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestProjectService_Delete_MapsSqlNoRows_ToErrNotFound(t *testing.T) {
	repo := &fakeProjectRepo{
		deleteFn: func(ctx context.Context, userID, projectID string) error {
			return sql.ErrNoRows
		},
	}
	svc := _service.NewProjectService(repo)

	err := svc.Delete(context.Background(), "user-1", "proj-1")
	if !errors.Is(err, _service.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
