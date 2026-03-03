package usecase

import (
	"context"
	"errors"
	d "todo/task-service/internal/domain"
)

var ErrTitleRequired = errors.New("title must not be empty")

type CreateTask struct {
	repo d.TaskRepo
}

func NewCreateTask(repo d.TaskRepo) *CreateTask {
	return &CreateTask{repo: repo}
}

func (uc *CreateTask) Exec(ctx context.Context, title, text string) (d.Task, error) {
	if title == "" {
		return d.Task{}, ErrTitleRequired
	}
	return uc.repo.Create(ctx, d.TaskCreateRequest{Title: title, Text: text})
}
