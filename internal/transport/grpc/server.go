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
	listTasks  *usecase.ListTasks
}

func NewServer(
	createTask *usecase.CreateTask, getTask *usecase.GetTask, listTasks *usecase.ListTasks,
) *Server {
	return &Server{
		createTask: createTask,
		getTask:    getTask,
		listTasks:  listTasks,
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

	return s.mapToTask(t), nil
}

func (s *Server) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.Task, error) {
	id := req.GetId()
	u, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "id must be uint32")
	}
	num := domain.TaskID(uint32(u))
	t, found, err := s.getTask.Exec(ctx, num)
	if err != nil {
		if errors.Is(err, usecase.ErrIllegalID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, "this id is not presented")
	}

	return s.mapToTask(t), nil
}

func (s *Server) ListTasks(ctx context.Context, req *taskpb.ListTasksRequest) (*taskpb.ListTasksResponse, error) {
	limit := req.GetLimit()
	offset := req.GetOffset()
	st := req.GetStatus()
	query := req.GetQuery()

	tasks, total, err := s.listTasks.Exec(ctx, limit, offset, st, query)
	if err != nil {
		if errors.Is(err, usecase.ErrBadPagination) || errors.Is(err, usecase.ErrBadStatus) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &taskpb.ListTasksResponse{
		Tasks: make([]*taskpb.Task, 0, len(tasks)),
		Total: total,
	}

	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, s.mapToTask(t))
	}

	return resp, nil
}

func (s *Server) mapToTask(t domain.Task) *taskpb.Task {
	return &taskpb.Task{
		Id:     fmt.Sprintf("%d", uint32(t.ID)),
		Title:  t.Title,
		Text:   t.Text,
		Status: t.Status,
	}
}
