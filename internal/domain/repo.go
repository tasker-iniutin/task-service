package domain

import "context"

type TaskRepo interface {
	Create(ctx context.Context, t TaskCreateRequest) (Task, error)
	Get(ctx context.Context, id Id) (Task, bool, error)
	List(ctx context.Context) ([]Task, error)
}
