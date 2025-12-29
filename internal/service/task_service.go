package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"TaskFlow/internal/domain"
)

type TaskRepo interface {
	Create(ctx context.Context, userID, projectID, title string) (domain.Task, error)
	List(ctx context.Context, userID, projectID string, completed *bool, limit int, cursor *domain.Cursor) ([]domain.Task, *domain.Cursor, error)
	Get(ctx context.Context, userID, taskID string) (domain.Task, error)
	Update(ctx context.Context, userID, taskID string, title *string, completed *bool) (domain.Task, error)
	Delete(ctx context.Context, userID, taskID string) error
}

type TaskService struct {
	repo TaskRepo
}

func NewTaskService(repo TaskRepo) *TaskService { return &TaskService{repo: repo} }

func (s *TaskService) Create(ctx context.Context, userID, projectID, title string) (domain.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return domain.Task{}, errors.New("title required")
	}
	t, err := s.repo.Create(ctx, userID, projectID, title)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, ErrNotFound
	}
	return t, err
}

func (s *TaskService) List(ctx context.Context, userID, projectID string, completed *bool, limit int, cursor *domain.Cursor) (Page[domain.Task], error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	items, next, err := s.repo.List(ctx, userID, projectID, completed, limit, cursor)
	if err != nil {
		return Page[domain.Task]{}, err
	}
	return Page[domain.Task]{Items: items, NextCursor: next}, nil
}

func (s *TaskService) Get(ctx context.Context, userID, taskID string) (domain.Task, error) {
	t, err := s.repo.Get(ctx, userID, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, ErrNotFound
	}
	return t, err
}

func (s *TaskService) Update(ctx context.Context, userID, taskID string, title *string, completed *bool) (domain.Task, error) {
	if title != nil {
		trim := strings.TrimSpace(*title)
		if trim == "" {
			return domain.Task{}, errors.New("title cannot be empty")
		}
		title = &trim
	}
	t, err := s.repo.Update(ctx, userID, taskID, title, completed)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, ErrNotFound
	}
	return t, err
}

func (s *TaskService) Delete(ctx context.Context, userID, taskID string) error {
	err := s.repo.Delete(ctx, userID, taskID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
