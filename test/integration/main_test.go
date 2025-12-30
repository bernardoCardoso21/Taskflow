package integration

import (
	"TaskFlow/internal/repo/postgres"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile) // .../test/integration

	envPath := filepath.Join(dir, ".env.test")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("note: could not load %s: %v", envPath, err)
	} else {
		log.Printf("loaded env from %s", envPath)
	}

	os.Exit(m.Run())
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
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

func insertProject(t *testing.T, db *sql.DB, id, userID, name string) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO projects (id, user_id, name)
		VALUES ($1, $2, $3)
	`, id, userID, name)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
}

func deleteProject(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	_, err := db.Exec(`DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		t.Fatalf("delete project: %v", err)
	}
}
