package postgre

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"

	d "github.com/tasker-iniutin/task-service/internal/domain"
)

type taskPosgtreImpl struct {
	db *pgxpool.Pool
}

func NewTaskPostgreRepo(db *pgxpool.Pool) *taskPosgtreImpl {
	return &taskPosgtreImpl{db: db}
}

func (r *taskPosgtreImpl) Create(ctx context.Context, t d.TaskCreateRequest) (d.Task, error) {
	const q = `
		INSERT INTO tasks (user_id, title, text, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, title, text, status
	`

	var nT d.Task
	err := r.db.QueryRow(
		ctx,
		q,
		t.UserId,
		t.Title,
		t.Text,
		taskpb.TaskStatus_TASK_STATUS_NEW,
	).Scan(
		&nT.ID,
		&nT.UserId,
		&nT.Title,
		&nT.Text,
		&nT.Status,
	)

	if err != nil {
		return d.Task{}, fmt.Errorf("create task: %w", err)
	}

	return nT, nil
}

func (r *taskPosgtreImpl) Get(ctx context.Context, id d.TaskID) (d.Task, bool, error) {
	const q = `
		SELECT id, user_id, title, text, status
		FROM tasks
		WHERE id = $1
	`

	var t d.Task
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID,
		&t.UserId,
		&t.Title,
		&t.Text,
		&t.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Task{}, false, nil
		}
		return d.Task{}, false, fmt.Errorf("get task: %w", err)
	}

	return t, true, nil
}

func (r *taskPosgtreImpl) Update(ctx context.Context, t d.TaskUpdateRequest) (d.Task, bool, error) {
	const q = `
		UPDATE tasks
		SET user_id = $2, title = $3, text = $4, status = $5
		WHERE id = $1
		RETURNING id, user_id, title, text, status
	`

	var updated d.Task
	err := r.db.QueryRow(ctx, q, t.ID, t.UserId, t.Title, t.Text, t.Status).Scan(
		&updated.ID,
		&updated.UserId,
		&updated.Title,
		&updated.Text,
		&updated.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Task{}, false, nil
		}
		return d.Task{}, false, fmt.Errorf("update task: %w", err)
	}

	return updated, true, nil
}

func (r *taskPosgtreImpl) Delete(ctx context.Context, id d.TaskID) (bool, error) {
	const query = `
		DELETE FROM tasks
		WHERE id = $1
	`

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("delete task: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}

func (r *taskPosgtreImpl) List(ctx context.Context, filter string, u_id d.UserID) ([]d.Task, error) {
	const q = `
		SELECT id, user_id, title, text, status
		FROM tasks
		WHERE user_id = $1
		  AND (
		    $2 = ''
		    OR LOWER(title) LIKE '%' || $2 || '%'
		    OR LOWER(text) LIKE '%' || $2 || '%'
		  )
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, q, u_id, strings.ToLower(strings.TrimSpace(filter)))
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]d.Task, 0)
	for rows.Next() {
		var task d.Task
		if err := rows.Scan(
			&task.ID,
			&task.UserId,
			&task.Title,
			&task.Text,
			&task.Status,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}
