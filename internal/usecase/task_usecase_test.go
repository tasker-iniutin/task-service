package usecase

import (
	"context"
	"testing"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1"
	"github.com/tasker-iniutin/task-service/internal/store/mem"
)

func TestUpdateTaskRequiresOwnership(t *testing.T) {
	repo := mem.NewTaskRepo()
	createTask := NewCreateTask(repo)
	updateTask := NewUpdateTask(repo)

	task, err := createTask.Exec(context.Background(), "title", "text", 10, nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, found, err := updateTask.Exec(
		context.Background(),
		task.ID,
		"updated",
		"text",
		taskpb.TaskStatus_TASK_STATUS_DONE,
		99,
		nil,
	)
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if found {
		t.Fatal("expected task update to fail for another user")
	}
}

func TestDeleteTaskRemovesTask(t *testing.T) {
	repo := mem.NewTaskRepo()
	createTask := NewCreateTask(repo)
	deleteTask := NewDeleteTask(repo)
	getTask := NewGetTask(repo)

	task, err := createTask.Exec(context.Background(), "title", "text", 10, nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	deleted, err := deleteTask.Exec(context.Background(), task.ID, 10)
	if err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if !deleted {
		t.Fatal("expected task to be deleted")
	}

	_, found, err := getTask.Exec(context.Background(), task.ID, 10)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if found {
		t.Fatal("expected deleted task to be absent")
	}
}

func TestListTasksUsesOffsetPagination(t *testing.T) {
	repo := mem.NewTaskRepo()
	createTask := NewCreateTask(repo)
	listTasks := NewListTasks(repo)

	for i := 0; i < 3; i++ {
		if _, err := createTask.Exec(context.Background(), "title", "text", 10, nil); err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}

	tasks, total, err := listTasks.Exec(
		context.Background(),
		2,
		"1",
		taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED,
		"",
		10,
	)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].ID != 2 {
		t.Fatalf("expected first paginated task id 2, got %d", tasks[0].ID)
	}
}

func TestListTasksRejectsInvalidPageToken(t *testing.T) {
	repo := mem.NewTaskRepo()
	listTasks := NewListTasks(repo)

	_, _, err := listTasks.Exec(
		context.Background(),
		10,
		"bad-token",
		taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED,
		"",
		10,
	)
	if err != ErrBadPagination {
		t.Fatalf("expected ErrBadPagination, got %v", err)
	}
}

func TestCreateTaskRejectsEmptyTitle(t *testing.T) {
	repo := mem.NewTaskRepo()
	createTask := NewCreateTask(repo)

	_, err := createTask.Exec(context.Background(), "", "text", 10, nil)
	if err != ErrTitleRequired {
		t.Fatalf("expected ErrTitleRequired, got %v", err)
	}
}

func TestGetTaskHidesForeignTask(t *testing.T) {
	repo := mem.NewTaskRepo()
	createTask := NewCreateTask(repo)
	getTask := NewGetTask(repo)

	task, err := createTask.Exec(context.Background(), "title", "text", 10, nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, found, err := getTask.Exec(context.Background(), task.ID, 11)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if found {
		t.Fatal("expected foreign task to be hidden")
	}
}

func TestUpdateTaskRejectsInvalidStatus(t *testing.T) {
	repo := mem.NewTaskRepo()
	createTask := NewCreateTask(repo)
	updateTask := NewUpdateTask(repo)

	task, err := createTask.Exec(context.Background(), "title", "text", 10, nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	_, _, err = updateTask.Exec(context.Background(), task.ID, "title", "text", taskpb.TaskStatus(99), 10, nil)
	if err != ErrBadStatus {
		t.Fatalf("expected ErrBadStatus, got %v", err)
	}
}

func TestDeleteTaskRequiresOwnership(t *testing.T) {
	repo := mem.NewTaskRepo()
	createTask := NewCreateTask(repo)
	deleteTask := NewDeleteTask(repo)

	task, err := createTask.Exec(context.Background(), "title", "text", 10, nil)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	deleted, err := deleteTask.Exec(context.Background(), task.ID, 11)
	if err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if deleted {
		t.Fatal("expected foreign task delete to fail")
	}
}
