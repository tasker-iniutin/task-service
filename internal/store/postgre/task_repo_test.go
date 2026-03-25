package postgre

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1"
	d "github.com/tasker-iniutin/task-service/internal/domain"
)

func TestTaskRepoCRUD(t *testing.T) {
	db := openTaskTestDB(t)
	repo := NewTaskPostgreRepo(db)

	exp := time.Now().UTC().Truncate(time.Second)
	task, err := repo.Create(context.Background(), d.TaskCreateRequest{
		UserId:    7,
		Title:     "write tests",
		Text:      "repo tests",
		ExpiresAt: &exp,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Status != taskpb.TaskStatus_TASK_STATUS_NEW {
		t.Fatalf("expected NEW status, got %v", task.Status)
	}

	got, found, err := repo.Get(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if !found {
		t.Fatal("expected task to exist")
	}
	if got.Title != "write tests" {
		t.Fatalf("unexpected title: %q", got.Title)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(exp) {
		t.Fatalf("unexpected expires_at: %v", got.ExpiresAt)
	}

	updated, found, err := repo.Update(context.Background(), d.TaskUpdateRequest{
		ID:     task.ID,
		UserId: 7,
		Title:  "updated",
		Text:   "updated text",
		Status: taskpb.TaskStatus_TASK_STATUS_DONE,
	})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if !found {
		t.Fatal("expected task to be updated")
	}
	if updated.Status != taskpb.TaskStatus_TASK_STATUS_DONE {
		t.Fatalf("expected DONE status, got %v", updated.Status)
	}

	list, _, err := repo.List(context.Background(), "upd", int32(taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED), 10, 0, 7)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 task, got %d", len(list))
	}

	deleted, err := repo.Delete(context.Background(), task.ID, 7)
	if err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if !deleted {
		t.Fatal("expected task to be deleted")
	}
}

func TestTaskRepoListFiltersByUser(t *testing.T) {
	db := openTaskTestDB(t)
	repo := NewTaskPostgreRepo(db)

	_, err := repo.Create(context.Background(), d.TaskCreateRequest{
		UserId: 1,
		Title:  "mine",
		Text:   "visible",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	_, err = repo.Create(context.Background(), d.TaskCreateRequest{
		UserId: 2,
		Title:  "other",
		Text:   "hidden",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	list, _, err := repo.List(context.Background(), "", int32(taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED), 10, 0, 1)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 task, got %d", len(list))
	}
	if list[0].UserId != 1 {
		t.Fatalf("expected user_id 1, got %d", list[0].UserId)
	}
}

func TestTaskRepoUpdateRequiresOwnership(t *testing.T) {
	db := openTaskTestDB(t)
	repo := NewTaskPostgreRepo(db)

	task, err := repo.Create(context.Background(), d.TaskCreateRequest{
		UserId: 1,
		Title:  "mine",
		Text:   "original",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, found, err := repo.Update(context.Background(), d.TaskUpdateRequest{
		ID:     task.ID,
		UserId: 2,
		Title:  "other",
		Text:   "other",
		Status: taskpb.TaskStatus_TASK_STATUS_DONE,
	})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if found {
		t.Fatal("expected update to fail for another user")
	}
}

func TestTaskRepoDeleteRequiresOwnership(t *testing.T) {
	db := openTaskTestDB(t)
	repo := NewTaskPostgreRepo(db)

	task, err := repo.Create(context.Background(), d.TaskCreateRequest{
		UserId: 1,
		Title:  "mine",
		Text:   "original",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	deleted, err := repo.Delete(context.Background(), task.ID, 2)
	if err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if deleted {
		t.Fatal("expected delete to fail for another user")
	}
}

func TestTaskRepoListStatusAndPagination(t *testing.T) {
	db := openTaskTestDB(t)
	repo := NewTaskPostgreRepo(db)

	t1, err := repo.Create(context.Background(), d.TaskCreateRequest{
		UserId: 1,
		Title:  "one",
		Text:   "a",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	t2, err := repo.Create(context.Background(), d.TaskCreateRequest{
		UserId: 1,
		Title:  "two",
		Text:   "b",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	t3, err := repo.Create(context.Background(), d.TaskCreateRequest{
		UserId: 1,
		Title:  "three",
		Text:   "c",
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if _, found, err := repo.Update(context.Background(), d.TaskUpdateRequest{
		ID:     t2.ID,
		UserId: 1,
		Title:  t2.Title,
		Text:   t2.Text,
		Status: taskpb.TaskStatus_TASK_STATUS_DONE,
	}); err != nil {
		t.Fatalf("update task: %v", err)
	} else if !found {
		t.Fatal("expected task to be updated")
	}

	list, total, err := repo.List(context.Background(), "", int32(taskpb.TaskStatus_TASK_STATUS_DONE), 10, 0, 1)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(list) != 1 || list[0].ID != t2.ID {
		t.Fatalf("unexpected DONE list result: %+v", list)
	}

	list, total, err = repo.List(context.Background(), "", int32(taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED), 1, 1, 1)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	if len(list) != 1 || list[0].ID != t2.ID {
		t.Fatalf("unexpected paginated result: %+v (t1=%d t2=%d t3=%d)", list, t1.ID, t2.ID, t3.ID)
	}
}

func openTaskTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TASK_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TASK_TEST_DATABASE_URL or DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	setupTaskSchema(t, db)
	return db
}

func setupTaskSchema(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	const schema = `
		CREATE TABLE IF NOT EXISTS tasks (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			title TEXT NOT NULL,
			text TEXT NOT NULL DEFAULT '',
			status INTEGER NOT NULL,
			expires_at TIMESTAMP NULL,
			CONSTRAINT tasks_title_not_empty CHECK (btrim(title) <> ''),
			CONSTRAINT tasks_status_valid CHECK (status IN (1, 2))
		);
		CREATE INDEX IF NOT EXISTS idx_tasks_user_id_id ON tasks (user_id, id);
		TRUNCATE TABLE tasks RESTART IDENTITY;
	`

	if _, err := db.Exec(context.Background(), schema); err != nil {
		t.Fatalf("setup task schema: %v", err)
	}
}
