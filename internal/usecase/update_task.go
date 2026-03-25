package usecase

import (
	"context"
	"time"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1"
	d "github.com/tasker-iniutin/task-service/internal/domain"
)

type UpdateTask struct {
	repo d.TaskRepo
}

func NewUpdateTask(repo d.TaskRepo) *UpdateTask {
	return &UpdateTask{repo: repo}
}

func (uc *UpdateTask) Exec(
	ctx context.Context,
	id d.TaskID,
	title string,
	text string,
	status taskpb.TaskStatus,
	uID d.UserID,
	expiresAt *time.Time,
) (d.Task, bool, error) {
	if id == 0 {
		return d.Task{}, false, ErrIllegalID
	}
	if title == "" {
		return d.Task{}, false, ErrTitleRequired
	}
	if !isValidTaskStatus(status, false) {
		return d.Task{}, false, ErrBadStatus
	}

	task, found, err := uc.repo.Get(ctx, id)
	if err != nil {
		return d.Task{}, false, err
	}
	if !found || task.UserId != uID {
		return d.Task{}, false, nil
	}

	return uc.repo.Update(ctx, d.TaskUpdateRequest{
		ID:     id,
		UserId: uID,
		Title:  title,
		Text:   text,
		Status: status,
		ExpiresAt: expiresAt,
	})
}
