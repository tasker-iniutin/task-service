package app

import (
	"context"
	"log"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"
	authsec "github.com/tasker-iniutin/common/authsecurity"
	"github.com/tasker-iniutin/common/grpcauth"
	"github.com/tasker-iniutin/common/postgres"
	"github.com/tasker-iniutin/common/runtime"
	"google.golang.org/grpc"

	"github.com/tasker-iniutin/task-service/internal/store/postgre"
	handler "github.com/tasker-iniutin/task-service/internal/transport/grpc"
	"github.com/tasker-iniutin/task-service/internal/usecase"
)

type App struct {
	cfg Config
}

func New(cfg Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Run(ctx context.Context) error {
	db, err := postgres.Open(context.Background(), a.cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := postgre.NewTaskPostgreRepo(db)

	createTask := usecase.NewCreateTask(repo)
	getTask := usecase.NewGetTask(repo)
	listTasks := usecase.NewListTasks(repo)
	updateTask := usecase.NewUpdateTask(repo)
	deleteTask := usecase.NewDeleteTask(repo)

	h := handler.NewServer(createTask, getTask, listTasks, updateTask, deleteTask)

	pub, err := authsec.LoadRSAPublicKeyFromPEMFile(a.cfg.PublicKeyPath)
	if err != nil {
		return err
	}

	verifier := authsec.NewRS256Verifier(pub, a.cfg.JWTIssuer, a.cfg.JWTAudience)

	whitelist := map[string]struct{}{}

	log.Printf("task-service gRPC listening on %s", a.cfg.GRPCAddr)
	log.Printf("task-service auth config: public_key=%s issuer=%s audience=%s",
		a.cfg.PublicKeyPath, a.cfg.JWTIssuer, a.cfg.JWTAudience,
	)
	log.Printf("task-service database config: dsn=%s", a.cfg.DatabaseURL)

	return runtime.ServeGRPCWithContext(
		ctx,
		a.cfg.GRPCAddr,
		func(server *grpc.Server) {
			taskpb.RegisterTaskServiceServer(server, h)
		},
		a.cfg.EnableReflection,
		grpc.UnaryInterceptor(grpcauth.UnaryAuthInterceptor(verifier, whitelist)),
	)
}
