package integration_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"TaskFlow/internal/repo/postgres"

	"github.com/google/uuid"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func insertUser(t *testing.T, db *sql.DB, id, email string) {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)
	`, id, email, "fake-hash-for-tests")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func deleteUser(t *testing.T, db *sql.DB, id string) {
	t.Helper()

	_, err := db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		t.Fatalf("delete user: %v", err)
	}
}

func TestProjectRepo_CRUD_Ownership(t *testing.T) {
	db := openTestDB(t)
	repo := postgres.NewProjectRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userA := uuid.NewString()
	userB := uuid.NewString()

	insertUser(t, db, userA, "a-"+uuid.NewString()+"@example.com")
	insertUser(t, db, userB, "b-"+uuid.NewString()+"@example.com")

	t.Cleanup(func() { deleteUser(t, db, userA) })
	t.Cleanup(func() { deleteUser(t, db, userB) })

	// Create a project for userA
	p, err := repo.Create(ctx, userA, "Project A")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// userA can get it
	if _, err := repo.Get(ctx, userA, p.ID); err != nil {
		t.Fatalf("get as owner: %v", err)
	}

	// userB cannot get it (should look like not found)
	if _, err := repo.Get(ctx, userB, p.ID); err == nil {
		t.Fatalf("expected error for non-owner get")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for non-owner get, got %v", err)
	}

	// update as non-owner should fail (sql.ErrNoRows)
	if _, err := repo.UpdateName(ctx, userB, p.ID, "Hacked"); err == nil {
		t.Fatalf("expected error for non-owner update")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for non-owner update, got %v", err)
	}

	// delete as non-owner should fail (sql.ErrNoRows)
	if err := repo.Delete(ctx, userB, p.ID); err == nil {
		t.Fatalf("expected error for non-owner delete")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for non-owner delete, got %v", err)
	}

	// delete as owner succeeds
	if err := repo.Delete(ctx, userA, p.ID); err != nil {
		t.Fatalf("delete as owner: %v", err)
	}

	// confirm it's gone
	if _, err := repo.Get(ctx, userA, p.ID); err == nil {
		t.Fatalf("expected project to be deleted, but Get succeeded")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}
}
