package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(email, passwordHash string) (string, error)
	FindUserByEmail(email string) (id string, passwordHash string, err error)
}

type AuthService struct {
	users  UserRepo
	secret string
}

func NewAuthService(users UserRepo, secret string) *AuthService {
	return &AuthService{users: users, secret: secret}
}

func (s *AuthService) Register(email, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return s.users.CreateUser(email, string(hash))
}

func (s *AuthService) Login(email, password string) (string, error) {
	id, hash, err := s.users.FindUserByEmail(email)
	if err != nil {
		return "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", errors.New("invalid credentials")
	}

	claims := jwt.MapClaims{
		"sub": id,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.secret))
}
