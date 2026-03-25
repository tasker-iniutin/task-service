package usecase

import (
	"context"
	"errors"
	"math"
	"strconv"

	d "github.com/tasker-iniutin/task-service/internal/domain"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"
)

var ErrBadPagination = errors.New("pagination failed")
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

	if !isValidTaskStatus(status, true) {
		return nil, 0, ErrBadStatus
	}
	offset, err := parsePageToken(token)
	if err != nil {
		return nil, 0, ErrBadPagination
	}

	tasks, total, err := uc.repo.List(ctx, query, int32(status), size, offset, u_id)
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func parsePageToken(token string) (uint32, error) {
	if token == "" {
		return 0, nil
	}

	offset, err := strconv.ParseUint(token, 10, 64)
	if err != nil {
		return 0, ErrBadPagination
	}
	if offset > math.MaxUint32 {
		return 0, ErrBadPagination
	}

	return uint32(offset), nil
}

func isValidTaskStatus(status taskpb.TaskStatus, allowUnspecified bool) bool {
	switch status {
	case taskpb.TaskStatus_TASK_STATUS_NEW, taskpb.TaskStatus_TASK_STATUS_DONE:
		return true
	case taskpb.TaskStatus_TASK_STATUS_UNSPECIFIED:
		return allowUnspecified
	default:
		return false
	}
}
