package grpc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1"
	authctx "github.com/tasker-iniutin/common/authctx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tasker-iniutin/task-service/internal/domain"
	"github.com/tasker-iniutin/task-service/internal/usecase"
)

type Server struct {
	taskpb.UnimplementedTaskServiceServer
	createTask *usecase.CreateTask
	getTask    *usecase.GetTask
	listTasks  *usecase.ListTasks
	updateTask *usecase.UpdateTask
	deleteTask *usecase.DeleteTask
}

func NewServer(
	createTask *usecase.CreateTask,
	getTask *usecase.GetTask,
	listTasks *usecase.ListTasks,
	updateTask *usecase.UpdateTask,
	deleteTask *usecase.DeleteTask,
) *Server {
	return &Server{
		createTask: createTask,
		getTask:    getTask,
		listTasks:  listTasks,
		updateTask: updateTask,
		deleteTask: deleteTask,
	}
}

func (s *Server) CreateTask(ctx context.Context, req *taskpb.CreateTaskRequest) (*taskpb.Task, error) {
	userID, ok := authctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	expiresAt, err := toTimePtr(req.GetExpiresAt())
	if err != nil {
		return nil, err
	}

	task, err := s.createTask.Exec(ctx, req.GetTitle(), req.GetText(), domain.UserID(userID), expiresAt)
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

	taskIDNum, err := strconv.ParseUint(req.GetId(), 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	task, found, err := s.getTask.Exec(ctx, domain.TaskID(taskIDNum), domain.UserID(userID))
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

func (s *Server) UpdateTask(ctx context.Context, req *taskpb.UpdateTaskRequest) (*taskpb.Task, error) {
	userID, ok := authctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	taskIDNum, err := strconv.ParseUint(req.GetId(), 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	expiresAt, err := toTimePtr(req.GetExpiresAt())
	if err != nil {
		return nil, err
	}

	task, found, err := s.updateTask.Exec(
		ctx,
		domain.TaskID(taskIDNum),
		req.GetTitle(),
		req.GetText(),
		req.GetStatus(),
		domain.UserID(userID),
		expiresAt,
	)
	if err != nil {
		if errors.Is(err, usecase.ErrIllegalID) ||
			errors.Is(err, usecase.ErrTitleRequired) ||
			errors.Is(err, usecase.ErrBadStatus) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !found {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	return s.mapToTask(task), nil
}

func (s *Server) DeleteTask(ctx context.Context, req *taskpb.DeleteTaskRequest) (*emptypb.Empty, error) {
	userID, ok := authctx.UserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing authenticated user")
	}

	taskIDNum, err := strconv.ParseUint(req.GetId(), 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task id")
	}

	deleted, err := s.deleteTask.Exec(ctx, domain.TaskID(taskIDNum), domain.UserID(userID))
	if err != nil {
		if errors.Is(err, usecase.ErrIllegalID) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !deleted {
		return nil, status.Error(codes.NotFound, "task not found")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) mapToTask(t domain.Task) *taskpb.Task {
	return &taskpb.Task{
		Id:     fmt.Sprintf("%d", uint64(t.ID)),
		UserId: fmt.Sprintf("%d", uint64(t.UserId)),
		Title:  t.Title,
		Text:   t.Text,
		Status: t.Status,
		ExpiresAt: toTimestamp(t.ExpiresAt),
	}
}

func toTimePtr(ts *timestamppb.Timestamp) (*time.Time, error) {
	if ts == nil {
		return nil, nil
	}
	if err := ts.CheckValid(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid expires_at")
	}
	t := ts.AsTime()
	return &t, nil
}

func toTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
