package postgre

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1"

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
		INSERT INTO tasks (user_id, title, text, status, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, text, status, expires_at
	`

	var nT d.Task
	err := r.db.QueryRow(
		ctx,
		q,
		t.UserId,
		t.Title,
		t.Text,
		taskpb.TaskStatus_TASK_STATUS_NEW,
		t.ExpiresAt,
	).Scan(
		&nT.ID,
		&nT.UserId,
		&nT.Title,
		&nT.Text,
		&nT.Status,
		&nT.ExpiresAt,
	)

	if err != nil {
		return d.Task{}, fmt.Errorf("create task: %w", err)
	}

	return nT, nil
}

func (r *taskPosgtreImpl) Get(ctx context.Context, id d.TaskID) (d.Task, bool, error) {
	const q = `
		SELECT id, user_id, title, text, status, expires_at
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
		&t.ExpiresAt,
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
		SET title = $3, text = $4, status = $5, expires_at = COALESCE($6, expires_at)
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, title, text, status, expires_at
	`

	var updated d.Task
	err := r.db.QueryRow(ctx, q, t.ID, t.UserId, t.Title, t.Text, t.Status, t.ExpiresAt).Scan(
		&updated.ID,
		&updated.UserId,
		&updated.Title,
		&updated.Text,
		&updated.Status,
		&updated.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return d.Task{}, false, nil
		}
		return d.Task{}, false, fmt.Errorf("update task: %w", err)
	}

	return updated, true, nil
}

func (r *taskPosgtreImpl) Delete(ctx context.Context, id d.TaskID, uID d.UserID) (bool, error) {
	const query = `
		DELETE FROM tasks
		WHERE id = $1 AND user_id = $2
	`

	cmdTag, err := r.db.Exec(ctx, query, id, uID)
	if err != nil {
		return false, fmt.Errorf("delete task: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return false, nil
	}

	return true, nil
}

func (r *taskPosgtreImpl) List(ctx context.Context, filter string, status int32, limit, offset uint32, u_id d.UserID) ([]d.Task, uint32, error) {
	const countQ = `
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1
		  AND (
		    $2 = ''
		    OR LOWER(title) LIKE '%' || $2 || '%'
		    OR LOWER(text) LIKE '%' || $2 || '%'
		  )
		  AND ($3 = 0 OR status = $3)
	`
	const q = `
		SELECT id, user_id, title, text, status, expires_at
		FROM tasks
		WHERE user_id = $1
		  AND (
		    $2 = ''
		    OR LOWER(title) LIKE '%' || $2 || '%'
		    OR LOWER(text) LIKE '%' || $2 || '%'
		  )
		  AND ($3 = 0 OR status = $3)
		ORDER BY id ASC
		LIMIT $4 OFFSET $5
	`

	f := strings.ToLower(strings.TrimSpace(filter))

	var total int64
	if err := r.db.QueryRow(ctx, countQ, u_id, f, status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}
	if total > math.MaxUint32 {
		return nil, 0, fmt.Errorf("task total exceeds limit: %d", total)
	}

	rows, err := r.db.Query(ctx, q, u_id, f, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
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
			&task.ExpiresAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, uint32(total), nil
}
