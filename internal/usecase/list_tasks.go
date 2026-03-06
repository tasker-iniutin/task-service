package usecase

import (
	"context"
	"errors"
	d "todo/task-service/internal/domain"

	taskpb "github.com/you/todo/api-contracts/gen/go/proto/task/v1alpha"
)

var ErrBadPagination = errors.New("padination failed")
var ErrBadStatus = errors.New("bad status")

type ListTasks struct {
	repo d.TaskRepo
}

func NewListTasks(repo d.TaskRepo) *ListTasks {
	return &ListTasks{repo: repo}
}

const (
	defaultLimit = uint32(50)
	maxLimit     = uint32(200)
)

func (uc *ListTasks) Exec(
	ctx context.Context,
	limit uint32,
	offset uint32,
	status taskpb.TaskStatus,
	query string,
) ([]d.Task, uint32, error) {

	if limit == 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		return nil, 0, ErrBadPagination
	}

	switch status {
	case taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED,
		taskpb.TaskStatus_TASK_STATUS_NEW,
		taskpb.TaskStatus_TASK_STATUS_DONE:
		// ok
	default:
		return nil, 0, ErrBadStatus
	}
	ts, err := uc.repo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	if status != taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED {
		out := make([]d.Task, 0, len(ts))
		for _, t := range ts {
			if t.Status == status {
				out = append(out, t)
			}
		}
		ts = out
	}

	total := uint32(len(ts))

	if offset >= total {
		return []d.Task{}, total, nil
	}
	end := min(offset+limit, total)
	return ts[offset:end], total, nil
}
