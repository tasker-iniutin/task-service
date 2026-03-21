package domain

import tp "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"

type TaskID uint64
type UserID uint64
type TaskCreateRequest struct {
	UserId UserID
	Title  string
	Text   string
}

type TaskUpdateRequest struct {
	ID     TaskID
	UserId UserID
	Title  string
	Text   string
	Status tp.TaskStatus
}

type Task struct {
	ID     TaskID
	UserId UserID
	Title  string
	Text   string
	Status tp.TaskStatus
}

type TaskFilter struct {
	Query  string
	Status tp.TaskStatus
	Limit  uint32
	Offset uint32
}
