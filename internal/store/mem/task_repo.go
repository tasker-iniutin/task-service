package mem

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"

	d "github.com/tasker-iniutin/task-service/internal/domain"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1"
)

var ErrMapIsFull = errors.New("map is full")

type taskRepoImpl struct {
	mu      sync.RWMutex
	byID    map[d.TaskID]d.Task
	counter uint32
}

func NewTaskRepo() *taskRepoImpl {
	return &taskRepoImpl{
		byID:    make(map[d.TaskID]d.Task),
		counter: 1,
	}
}

func (r *taskRepoImpl) Create(ctx context.Context, t d.TaskCreateRequest) (d.Task, error) {
	if err := ctx.Err(); err != nil {
		return d.Task{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.counter == 0 {
		return d.Task{}, ErrMapIsFull
	}

	id := d.TaskID(r.counter)
	nT := d.Task{
		ID:     id,
		UserId: t.UserId,
		Title:  t.Title,
		Text:   t.Text,
		Status: taskpb.TaskStatus_TASK_STATUS_NEW,
		ExpiresAt: t.ExpiresAt,
	}

	r.byID[id] = nT
	r.counter++

	return nT, nil
}

func (r *taskRepoImpl) Get(ctx context.Context, id d.TaskID) (d.Task, bool, error) {
	if err := ctx.Err(); err != nil {
		return d.Task{}, false, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byID[id]
	return t, ok, nil
}

func (r *taskRepoImpl) Update(ctx context.Context, t d.TaskUpdateRequest) (d.Task, bool, error) {
	if err := ctx.Err(); err != nil {
		return d.Task{}, false, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byID[t.ID]
	if !ok {
		return d.Task{}, false, nil
	}
	if existing.UserId != t.UserId {
		return d.Task{}, false, nil
	}

	existing.Title = t.Title
	existing.Text = t.Text
	existing.Status = t.Status
	if t.ExpiresAt != nil {
		existing.ExpiresAt = t.ExpiresAt
	}

	r.byID[t.ID] = existing
	return existing, true, nil
}

func (r *taskRepoImpl) Delete(ctx context.Context, id d.TaskID, uID d.UserID) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.byID[id]
	if !ok {
		return false, nil
	}
	if existing.UserId != uID {
		return false, nil
	}

	delete(r.byID, id)
	return true, nil
}

func (r *taskRepoImpl) List(ctx context.Context, filter string, status int32, limit, offset uint32, u_id d.UserID) ([]d.Task, uint32, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	q := strings.ToLower(strings.TrimSpace(filter))

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]d.Task, 0, len(r.byID))
	for _, t := range r.byID {
		if t.UserId != u_id {
			continue
		}
		if status != 0 && int32(t.Status) != status {
			continue
		}
		if q == "" {
			out = append(out, t)
			continue
		}
		if strings.Contains(strings.ToLower(t.Title), q) || strings.Contains(strings.ToLower(t.Text), q) {
			out = append(out, t)
		}
	}

	slices.SortFunc(out, func(a, b d.Task) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})

	total := uint32(len(out))
	if offset >= total {
		return []d.Task{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}

	off := int(offset)
	ed := int(end)
	return out[off:ed], total, nil
}
