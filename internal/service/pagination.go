package service

import "TaskFlow/internal/domain"

type Page[T any] struct {
	Items      []T
	NextCursor *domain.Cursor
}
