package grpc

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"
	authctx "github.com/tasker-iniutin/common/authctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tasker-iniutin/task-service/internal/domain"
	"github.com/tasker-iniutin/task-service/internal/usecase"
)

type Server struct {
	taskpb.UnimplementedTaskServiceServer
	createTask *usecase.CreateTask
	getTask    *usecase.GetTask
	listTasks  *usecase.ListTasks
}

func NewServer(
	createTask *usecase.CreateTask,
	getTask *usecase.GetTask,
	listTasks *usecase.ListTasks,
) *Server {
	return &Server{
		createTask: createTask,
		getTask:    getTask,
		listTasks:  listTasks,
	}
}

func (s *Server) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.Task, error) {
	userID, ok := authctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	task, err := s.createTask.Exec(ctx, req.GetTitle(), req.GetText(), domain.UserID(userID))
	if err != nil {
		if errors.Is(err, usecase.ErrTitleRequired) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.mapToTask(task), nil
}

func (s *Server) GetTask(ctx context.Context, req *taskpb.GetTaskRequest) (*taskpb.Task, error) {
	userID, ok := authctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	taskIDNum, err := strconv.ParseUint(req.GetId(), 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	task, found, err := s.getTask.Exec(ctx, domain.TaskID(uint32(taskIDNum)), domain.UserID(userID))
	if err != nil {
		if errors.Is(err, usecase.ErrIllegalID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	return s.mapToTask(task), nil
}

func (s *Server) ListTasks(ctx context.Context, req *taskpb.ListTasksRequest) (*taskpb.ListTasksResponse, error) {
	userID, ok := authctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	tasks, total, err := s.listTasks.Exec(
		ctx,
		req.GetPageSize(),
		req.GetPageToken(),
		req.GetStatus(),
		req.GetQuery(),
		domain.UserID(userID),
	)
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

	for _, task := range tasks {
		resp.Tasks = append(resp.Tasks, s.mapToTask(task))
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
