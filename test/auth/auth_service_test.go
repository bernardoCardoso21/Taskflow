package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	_service "TaskFlow/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	// behavior knobs
	createErr error
	findErr   error

	// captured inputs
	createdEmail string
	createdHash  string

	// find return values
	foundID   string
	foundHash string
}

func (f *fakeUserRepo) CreateUser(email, passwordHash string) (string, error) {
	f.createdEmail = email
	f.createdHash = passwordHash
	if f.createErr != nil {
		return "", f.createErr
	}
	// return a stable fake id
	return "user-123", nil
}

func (f *fakeUserRepo) FindUserByEmail(email string) (string, string, error) {
	if f.findErr != nil {
		return "", "", f.findErr
	}
	return f.foundID, f.foundHash, nil
}

func TestAuthService_Register_HashesPasswordAndCallsRepo(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := _service.NewAuthService(repo, "secret")

	id, err := svc.Register("test@example.com", "password123")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if id != "user-123" {
		t.Fatalf("expected id user-123, got %s", id)
	}

	if repo.createdEmail != "test@example.com" {
		t.Fatalf("expected email saved, got %s", repo.createdEmail)
	}
	if repo.createdHash == "" {
		t.Fatal("expected password hash to be set")
	}
	// bcrypt hashes start with $2a$ / $2b$ / $2y$
	if !strings.HasPrefix(repo.createdHash, "$2") {
		t.Fatalf("expected bcrypt hash, got %s", repo.createdHash)
	}
	// ensure hash matches original password
	if err := bcrypt.CompareHashAndPassword([]byte(repo.createdHash), []byte("password123")); err != nil {
		t.Fatalf("expected hash to validate original password, got %v", err)
	}
}

func TestAuthService_Register_PropagatesRepoError(t *testing.T) {
	repo := &fakeUserRepo{createErr: errors.New("duplicate")}
	svc := _service.NewAuthService(repo, "secret")

	_, err := svc.Register("test@example.com", "password123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "duplicate" {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestAuthService_Login_ReturnsSignedJWTWithSubject(t *testing.T) {
	secret := "my-secret"
	repo := &fakeUserRepo{
		foundID: "abc-123",
	}
	// make a real bcrypt hash for the password we will test
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt setup failed: %v", err)
	}
	repo.foundHash = string(hash)

	svc := _service.NewAuthService(repo, secret)

	tokenStr, err := svc.Login("test@example.com", "password123")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token")
	}

	// Parse and validate the token with the same secret
	parsed, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		// extra safety: enforce HS256
		if token.Method != jwt.SigningMethodHS256 {
			t.Fatalf("expected HS256, got %v", token.Method)
		}
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("expected token to be valid")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	// Subject must match user id
	if claims["sub"] != "abc-123" {
		t.Fatalf("expected sub abc-123, got %v", claims["sub"])
	}

	// exp should be ~24h in the future (allow some slack)
	expFloat, ok := claims["exp"].(float64) // jwt MapClaims decodes numbers as float64
	if !ok {
		t.Fatalf("expected exp to be a number, got %T", claims["exp"])
	}
	exp := time.Unix(int64(expFloat), 0)
	if exp.Before(time.Now().Add(23*time.Hour)) || exp.After(time.Now().Add(25*time.Hour)) {
		t.Fatalf("expected exp about 24h from now, got %v", exp)
	}
}

func TestAuthService_Login_WrongPassword_ReturnsInvalidCredentials(t *testing.T) {
	repo := &fakeUserRepo{
		foundID: "abc-123",
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt setup failed: %v", err)
	}
	repo.foundHash = string(hash)

	svc := _service.NewAuthService(repo, "secret")

	_, err = svc.Login("test@example.com", "wrong-password")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestAuthService_Login_RepoError_Propagates(t *testing.T) {
	repo := &fakeUserRepo{
		findErr: errors.New("not found"),
	}
	svc := _service.NewAuthService(repo, "secret")

	_, err := svc.Login("missing@example.com", "password123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "not found" {
		t.Fatalf("expected not found, got %v", err)
	}
}
