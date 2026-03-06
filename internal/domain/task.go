package domain

import tp "github.com/you/todo/api-contracts/gen/go/proto/task/v1alpha"

type TaskID uint32

type TaskCreateRequest struct {
	Title string
	Text  string
}

type Task struct {
	ID     TaskID
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
