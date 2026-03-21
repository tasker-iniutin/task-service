package domain

import "context"

type TaskRepo interface {
	Create(ctx context.Context, t TaskCreateRequest) (Task, error)
	Get(ctx context.Context, id TaskID) (Task, bool, error)
	Update(ctx context.Context, t TaskUpdateRequest) (Task, bool, error)
	Delete(ctx context.Context, id TaskID) (bool, error)
	List(ctx context.Context, filter string, u_id UserID) ([]Task, error)
}
