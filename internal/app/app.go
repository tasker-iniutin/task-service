package app

import (
	"log"
	"net"

	taskpb "github.com/you/todo/api-contracts/gen/go/proto/task/v1alpha"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"todo/task-service/internal/store/mem"
	handler "todo/task-service/internal/transport/grpc"
	"todo/task-service/internal/usecase"
)

type App struct {
	grpcAddr string
}

func CreateApp(grpcAddr string) *App {
	return &App{grpcAddr: grpcAddr}
}

func (a *App) Run() error {
	// infra
	repo := mem.NewTaskRepo()

	// usecases
	createTask := usecase.NewCreateTask(repo)
	getTask := usecase.NewGetTask(repo)
	listTasks := usecase.NewListTasks(repo)

	// handler
	h := handler.NewServer(createTask, getTask, listTasks)

	// gRPC runtime server
	grpcServer := grpc.NewServer()
	taskpb.RegisterTaskServiceServer(grpcServer, h)
	reflection.Register(grpcServer) // dev-only

	// listener
	lis, err := net.Listen("tcp", a.grpcAddr)
	if err != nil {
		return err
	}

	log.Printf("task-service gRPC listening on %s", a.grpcAddr)
	return grpcServer.Serve(lis)
}
