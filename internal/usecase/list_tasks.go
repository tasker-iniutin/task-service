package usecase

import (
	"context"
	"errors"
	"strconv"

	d "github.com/tasker-iniutin/task-service/internal/domain"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"
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
	size uint32,
	token string,
	status taskpb.TaskStatus,
	query string,
	u_id d.UserID,
) ([]d.Task, uint32, error) {

	if size == 0 {
		size = defaultLimit
	}
	if size > maxLimit {
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
	ts, err := uc.repo.List(ctx, query, u_id)
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

	total := len(ts)

	t, _ := strconv.Atoi(token)
	if t >= total {
		return []d.Task{}, uint32(total), nil
	}
	end := min(t+int(size), total)
	return ts[int(size)*t : end], uint32(total), nil
}
