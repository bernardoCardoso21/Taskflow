package service

import (
	"context"
	"database/sql"
	"errors"

	"TaskFlow/internal/domain"
)

var ErrNotFound = errors.New("not found")

type ProjectRepo interface {
	Create(ctx context.Context, userID, name string) (domain.Project, error)
	List(ctx context.Context, userID string, limit int, cursor *domain.Cursor) ([]domain.Project, *domain.Cursor, error)
	Get(ctx context.Context, userID, projectID string) (domain.Project, error)
	UpdateName(ctx context.Context, userID, projectID, name string) (domain.Project, error)
	Delete(ctx context.Context, userID, projectID string) error
}

type ProjectService struct {
	repo ProjectRepo
}

func NewProjectService(repo ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, userID, name string) (domain.Project, error) {
	return s.repo.Create(ctx, userID, name)
}

func (s *ProjectService) List(ctx context.Context, userID string, limit int, cursor *domain.Cursor) (Page[domain.Project], error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	items, next, err := s.repo.List(ctx, userID, limit, cursor)
	if err != nil {
		return Page[domain.Project]{}, err
	}
	return Page[domain.Project]{Items: items, NextCursor: next}, nil
}

func (s *ProjectService) Get(ctx context.Context, userID, projectID string) (domain.Project, error) {
	p, err := s.repo.Get(ctx, userID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	return p, err
}

func (s *ProjectService) UpdateName(ctx context.Context, userID, projectID, name string) (domain.Project, error) {
	p, err := s.repo.UpdateName(ctx, userID, projectID, name)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Project{}, ErrNotFound
	}
	return p, err
}

func (s *ProjectService) Delete(ctx context.Context, userID, projectID string) error {
	err := s.repo.Delete(ctx, userID, projectID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
