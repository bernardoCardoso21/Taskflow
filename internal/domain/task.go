package domain

import "time"

type Task struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
