package mem

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"

	d "todo/task-service/internal/domain"
)

var ErrMapIsFull = errors.New("map is full")

type taskRepoImpl struct {
	mu      sync.RWMutex
	byID    map[d.Id]d.Task
	counter uint32
}

func NewTaskRepo() *taskRepoImpl {
	return &taskRepoImpl{
		byID:    make(map[d.Id]d.Task),
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

	id := d.Id(r.counter)
	nT := d.Task{
		ID:     id,
		Title:  t.Title,
		Text:   t.Text,
		Status: d.StatusNew,
	}

	r.byID[id] = nT
	r.counter++

	return nT, nil
}

func (r *taskRepoImpl) Get(ctx context.Context, id d.Id) (d.Task, bool, error) {
	if err := ctx.Err(); err != nil {
		return d.Task{}, false, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.byID[id]
	return t, ok, nil
}

func (r *taskRepoImpl) List(ctx context.Context) ([]d.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return slices.Collect(maps.Values(r.byID)), nil
}
