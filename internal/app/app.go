package app

import (
	"log"
	"net"

	taskpb "github.com/tasker-iniutin/api-contracts/gen/go/proto/task/v1alpha"
	authsec "github.com/tasker-iniutin/common/authsecurity"
	"github.com/tasker-iniutin/common/grpcauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/tasker-iniutin/task-service/internal/store/mem"
	handler "github.com/tasker-iniutin/task-service/internal/transport/grpc"
	"github.com/tasker-iniutin/task-service/internal/usecase"
)

type App struct {
	grpcAddr         string
	publicKeyPath    string
	jwtIssuer        string
	jwtAudience      string
	enableReflection bool
}

func CreateApp(
	grpcAddr string,
	publicKeyPath string,
	jwtIssuer string,
	jwtAudience string,
	enableReflection bool,
) *App {
	return &App{
		grpcAddr:         grpcAddr,
		publicKeyPath:    publicKeyPath,
		jwtIssuer:        jwtIssuer,
		jwtAudience:      jwtAudience,
		enableReflection: enableReflection,
	}
}

func (a *App) Run() error {
	repo := mem.NewTaskRepo()

	createTask := usecase.NewCreateTask(repo)
	getTask := usecase.NewGetTask(repo)
	listTasks := usecase.NewListTasks(repo)

	h := handler.NewServer(createTask, getTask, listTasks)

	pub, err := authsec.LoadRSAPublicKeyFromPEMFile(a.publicKeyPath)
	if err != nil {
		return err
	}

	verifier := authsec.NewRS256Verifier(pub, a.jwtIssuer, a.jwtAudience)

	// Для task-service whitelist должен быть пустой:
	// все task методы должны требовать JWT.
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

	return grpcServer.Serve(lis)
}
