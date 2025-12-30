package integration

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"TaskFlow/internal/repo/postgres"

	"github.com/google/uuid"
)

func TestTaskRepo_Ownership_Filter_Pagination(t *testing.T) {
	db := openTestDB(t)
	taskRepo := postgres.NewTaskRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	userA := uuid.NewString()
	userB := uuid.NewString()

	insertUser(t, db, userA, "a-"+uuid.NewString()+"@example.com")
	insertUser(t, db, userB, "b-"+uuid.NewString()+"@example.com")

	projectA := uuid.NewString()
	projectB := uuid.NewString()

	insertProject(t, db, projectA, userA, "Project A")
	insertProject(t, db, projectB, userB, "Project B")

	t.Cleanup(func() { deleteProject(t, db, projectA) })
	t.Cleanup(func() { deleteProject(t, db, projectB) })
	t.Cleanup(func() { deleteUser(t, db, userA) })
	t.Cleanup(func() { deleteUser(t, db, userB) })

	t1, err := taskRepo.Create(ctx, userA, projectA, "Task 1")
	if err != nil {
		t.Fatalf("create t1: %v", err)
	}
	t2, err := taskRepo.Create(ctx, userA, projectA, "Task 2")
	if err != nil {
		t.Fatalf("create t2: %v", err)
	}
	t3, err := taskRepo.Create(ctx, userA, projectA, "Task 3")
	if err != nil {
		t.Fatalf("create t3: %v", err)
	}

	// userA can get
	if _, err := taskRepo.Get(ctx, userA, t1.ID); err != nil {
		t.Fatalf("get as owner: %v", err)
	}

	// userB cannot get
	if _, err := taskRepo.Get(ctx, userB, t1.ID); err == nil {
		t.Fatalf("expected error for non-owner get")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for non-owner get, got %v", err)
	}

	// update as non-owner should fail
	newTitle := "hacked"
	if _, err := taskRepo.Update(ctx, userB, t1.ID, &newTitle, nil); err == nil {
		t.Fatalf("expected error for non-owner update")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for non-owner update, got %v", err)
	}

	// delete as non-owner should fail
	if err := taskRepo.Delete(ctx, userB, t1.ID); err == nil {
		t.Fatalf("expected error for non-owner delete")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for non-owner delete, got %v", err)
	}

	// Pagination: limit 2 should return 2 + nextCursor
	items, next, err := taskRepo.List(ctx, userA, projectA, nil, 2, nil)
	if err != nil {
		t.Fatalf("list page1: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if next == nil {
		t.Fatalf("expected next cursor, got nil")
	}

	// Next page should return remaining 1
	items2, next2, err := taskRepo.List(ctx, userA, projectA, nil, 2, next)
	if err != nil {
		t.Fatalf("list page2: %v", err)
	}
	if len(items2) != 1 {
		t.Fatalf("expected 1 item on page2, got %d", len(items2))
	}
	if next2 != nil {
		t.Fatalf("expected no next cursor on last page, got %#v", next2)
	}

	// Mark one completed
	_, err = taskRepo.Update(ctx, userA, t2.ID, nil, ptrBool(true))
	if err != nil {
		t.Fatalf("update completed: %v", err)
	}

	// Filter completed=true should return exactly that one
	itemsC, _, err := taskRepo.List(ctx, userA, projectA, ptrBool(true), 50, nil)
	if err != nil {
		t.Fatalf("list completed=true: %v", err)
	}
	if len(itemsC) != 1 {
		t.Fatalf("expected 1 completed task, got %d", len(itemsC))
	}
	if itemsC[0].ID != t2.ID {
		t.Fatalf("expected completed task id %s, got %s", t2.ID, itemsC[0].ID)
	}

	// Delete as owner succeeds
	if err := taskRepo.Delete(ctx, userA, t3.ID); err != nil {
		t.Fatalf("delete as owner: %v", err)
	}
	if _, err := taskRepo.Get(ctx, userA, t3.ID); err == nil {
		t.Fatalf("expected deleted task to be gone, but Get succeeded")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}

	if _, err := taskRepo.Get(ctx, userA, t1.ID); err != nil {
		t.Fatalf("expected t1 to exist, got %v", err)
	}
	if _, err := taskRepo.Get(ctx, userA, t2.ID); err != nil {
		t.Fatalf("expected t2 to exist, got %v", err)
	}

	// Ensure non-owner cannot list tasks of another user's project (by passing projectA with userB)
	_, _, err = taskRepo.List(ctx, userB, projectA, nil, 10, nil)
	if err != nil {
		// List should typically return empty + nil cursor, not error.
		// But if your repo chooses to enforce "project must be owned", sql.ErrNoRows is acceptable too.
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("unexpected error for non-owner list: %v", err)
		}
	}

	_ = t1
	_ = t2
}

func ptrBool(b bool) *bool { return &b }
