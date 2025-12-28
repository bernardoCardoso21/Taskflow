package postgres

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) CreateUser(email, passwordHash string) (string, error) {
	id := uuid.NewString()
	_, err := r.db.Exec(
		`INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`,
		id, email, passwordHash,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *UserRepo) FindUserByEmail(email string) (string, string, error) {
	var id, hash string
	err := r.db.QueryRow(
		`SELECT id, password_hash FROM users WHERE email = $1`,
		email,
	).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.New("not found")
		}
		return "", "", err
	}
	return id, hash, nil
}
