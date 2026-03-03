package usecase

import (
	"context"
	"errors"
	d "todo/task-service/internal/domain"
)

var IllegalID = errors.New("id must be not null")

type GetTask struct {
	repo d.TaskRepo
}

func NewGetTask(repo d.TaskRepo) *GetTask {
	return &GetTask{repo: repo}
}

func (uc *GetTask) Exec(ctx context.Context, id d.Id) (d.Task, bool, error) {
	if id == 0 {
		return d.Task{}, false, IllegalID
	}
	return uc.repo.Get(ctx, id)
}
