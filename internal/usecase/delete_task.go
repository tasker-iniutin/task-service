package usecase

import (
	"context"

	d "github.com/tasker-iniutin/task-service/internal/domain"
)

type DeleteTask struct {
	repo d.TaskRepo
}

func NewDeleteTask(repo d.TaskRepo) *DeleteTask {
	return &DeleteTask{repo: repo}
}

func (uc *DeleteTask) Exec(ctx context.Context, id d.TaskID, uID d.UserID) (bool, error) {
	if id == 0 {
		return false, ErrIllegalID
	}

	task, found, err := uc.repo.Get(ctx, id)
	if err != nil {
		return false, err
	}
	if !found || task.UserId != uID {
		return false, nil
	}

	return uc.repo.Delete(ctx, id, uID)
}
