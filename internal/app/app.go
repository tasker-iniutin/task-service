package app

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"
	authsec "github.com/tasker-iniutin/common/authsecurity"
	"github.com/tasker-iniutin/common/grpcauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tasker-iniutin/task-service/internal/store/postgre"
	handler "github.com/tasker-iniutin/task-service/internal/transport/grpc"
	"github.com/tasker-iniutin/task-service/internal/usecase"
)

type App struct {
	grpcAddr         string
	publicKeyPath    string
	jwtIssuer        string
	jwtAudience      string
	enableReflection bool
	databaseAddr     string
}

func CreateApp(
	grpcAddr string,
	publicKeyPath string,
	jwtIssuer string,
	jwtAudience string,
	enableReflection bool,
	databaseAddr string,
) *App {
	return &App{
		grpcAddr:         grpcAddr,
		publicKeyPath:    publicKeyPath,
		jwtIssuer:        jwtIssuer,
		jwtAudience:      jwtAudience,
		enableReflection: enableReflection,
		databaseAddr:     databaseAddr,
	}
}

func (a *App) Run() error {
	db, err := pgxpool.New(context.Background(), a.databaseAddr)
	if err != nil {
		return fmt.Errorf("create db pool: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		db.Close()
		return fmt.Errorf("ping db: %w", err)
	}
	defer db.Close()

	repo := postgre.NewTaskPostgreRepo(db)

	createTask := usecase.NewCreateTask(repo)
	getTask := usecase.NewGetTask(repo)
	listTasks := usecase.NewListTasks(repo)
	updateTask := usecase.NewUpdateTask(repo)
	deleteTask := usecase.NewDeleteTask(repo)

	h := handler.NewServer(createTask, getTask, listTasks, updateTask, deleteTask)

	pub, err := authsec.LoadRSAPublicKeyFromPEMFile(a.publicKeyPath)
	if err != nil {
		return err
	}

	verifier := authsec.NewRS256Verifier(pub, a.jwtIssuer, a.jwtAudience)

	whitelist := map[string]struct{}{}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcauth.UnaryAuthInterceptor(verifier, whitelist)),
	)

	taskpb.RegisterTaskServiceServer(grpcServer, h)

	if a.enableReflection {
		reflection.Register(grpcServer)
	}

	lis, err := net.Listen("tcp", a.grpcAddr)
	if err != nil {
		return err
	}

	log.Printf("task-service gRPC listening on %s", a.grpcAddr)
	log.Printf("task-service auth config: public_key=%s issuer=%s audience=%s",
		a.publicKeyPath, a.jwtIssuer, a.jwtAudience,
	)
	log.Printf("task-service database config: dsn=%s", a.databaseAddr)

	return grpcServer.Serve(lis)
}
