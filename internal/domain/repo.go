package domain

import "context"

type TaskRepo interface {
	Create(ctx context.Context, t TaskCreateRequest) (Task, error)
	Get(ctx context.Context, id TaskID) (Task, bool, error)
	Update(ctx context.Context, t TaskUpdateRequest) (Task, bool, error)
	Delete(ctx context.Context, id TaskID, uID UserID) (bool, error)
	List(ctx context.Context, filter string, status int32, limit, offset uint32, uID UserID) ([]Task, uint32, error)
}
