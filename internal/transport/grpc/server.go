package grpc

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	taskpb "github.com/you/todo/api-contracts/gen/go/proto/task/v1alpha"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"todo/task-service/internal/domain"
	"todo/task-service/internal/usecase"
)

type Server struct {
	taskpb.UnimplementedTaskServiceServer
	createTask *usecase.CreateTask
	getTask    *usecase.GetTask
}

func NewServer(createTask *usecase.CreateTask, getTask *usecase.GetTask) *Server {
	return &Server{
		createTask: createTask,
		getTask:    getTask,
	}
}

func (s *Server) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.Task, error) {
	t, err := s.createTask.Exec(ctx, req.GetTitle(), req.GetText())
	if err != nil {
		if errors.Is(err, usecase.ErrTitleRequired) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.mapToTask(t)
}

func (s *Server) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.Task, error) {
	id := req.GetId()
	u, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id must be uint32")
	}
	num := domain.Id(uint32(u))
	t, found, err := s.getTask.Exec(ctx, num)
	if err != nil {
		if errors.Is(err, usecase.IllegalID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, "this id is not presented")
	}

	return s.mapToTask(t)
}

func (s *Server) mapToTask(t domain.Task) (*taskpb.Task, error) {
	return &taskpb.Task{
		Id:     fmt.Sprintf("%d", uint32(t.ID)),
		Title:  t.Title,
		Text:   t.Text,
		Status: string(t.Status),
	}, nil
}
