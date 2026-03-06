package mem

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"

	d "todo/task-service/internal/domain"

	taskpb "github.com/you/todo/api-contracts/gen/go/proto/task/v1alpha"
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
		Title:  t.Title,
		Text:   t.Text,
		Status: taskpb.TaskStatus_TASK_STATUS_NEW,
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

func (r *taskRepoImpl) List(ctx context.Context, filter string) ([]d.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	q := strings.ToLower(strings.TrimSpace(filter))

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]d.Task, 0, len(r.byID))
	for _, t := range r.byID {
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

	return out, nil
}
