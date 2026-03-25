package domain

import (
	"time"

	tp "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1"
)

type TaskID uint64
type UserID uint64
type TaskCreateRequest struct {
	UserId UserID
	Title  string
	Text   string
	// ExpiresAt is optional; nil means no expiration.
	ExpiresAt *time.Time
}

type TaskUpdateRequest struct {
	ID     TaskID
	UserId UserID
	Title  string
	Text   string
	Status tp.TaskStatus
	// ExpiresAt is optional; nil means keep current value.
	ExpiresAt *time.Time
}

type Task struct {
	ID     TaskID
	UserId UserID
	Title  string
	Text   string
	Status tp.TaskStatus
	ExpiresAt *time.Time
}

type TaskFilter struct {
	Query  string
	Status tp.TaskStatus
	Limit  uint32
	Offset uint32
}
