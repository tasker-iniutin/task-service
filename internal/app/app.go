package app

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	taskpb "github.com/you/todo/api-contracts/gen/go/proto/task/v1alpha"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"todo/task-service/internal/store/mem"
	taskhandler "todo/task-service/internal/transport/grpc"
	"todo/task-service/internal/usecase"
)

type App struct {
	httpAddr string
	grpcAddr string
}

func CreateApp(httpAddr, grpcAddr string) *App {
	return &App{httpAddr: httpAddr, grpcAddr: grpcAddr}
}

func (a *App) Run() error {
	// infra
	repo := mem.NewTaskRepo()

	// usecases
	createTask := usecase.NewCreateTask(repo)
	getTask := usecase.NewGetTask(repo)

	// handler
	handler := taskhandler.NewServer(createTask, getTask)

	// gRPC runtime server
	grpcServer := grpc.NewServer()
	taskpb.RegisterTaskServiceServer(grpcServer, handler)
	reflection.Register(grpcServer) // dev-only

	// listener
	lis, err := net.Listen("tcp", a.grpcAddr)
	if err != nil {
		return err
	}

	// 1) стартуем gRPC в фоне
	go func() {
		log.Printf("task-service gRPC listening on %s", a.grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	// 2) поднимаем HTTP gateway (он будет работать параллельно)
	ctx := context.Background()

	// ВАЖНО: NewClient вместо DialContext
	conn, err := grpc.NewClient(
		"localhost"+a.grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	mux := runtime.NewServeMux()

	if err := taskpb.RegisterTaskServiceHandler(ctx, mux, conn); err != nil {
		return err
	}

	log.Printf("task-service HTTP gateway listening on %s", a.httpAddr)
	return http.ListenAndServe(a.httpAddr, mux)
}
