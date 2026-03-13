package usecase

import (
	"context"
	"errors"

	d "github.com/tasker-iniutin/task-service/internal/domain"
)

var ErrIllegalID = errors.New("id must be not null")

type GetTask struct {
	repo d.TaskRepo
}

func NewGetTask(repo d.TaskRepo) *GetTask {
	return &GetTask{repo: repo}
}

func (uc *GetTask) Exec(ctx context.Context, id d.TaskID, u_id d.UserID) (d.Task, bool, error) {
	if id == 0 {
		return d.Task{}, false, ErrIllegalID
	}
	t, b, err := uc.repo.Get(ctx, id)
	if err != nil {
		return d.Task{}, false, err
	}
	if t.UserId != u_id {
		return d.Task{}, false, ErrIllegalID
	}
	return t, b, nil
}
