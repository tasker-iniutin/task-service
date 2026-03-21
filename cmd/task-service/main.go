package main

import (
	"log"
	"os"

	"github.com/tasker-iniutin/task-service/internal/app"
)

func main() {
	a := app.CreateApp(
		getenv("TASK_GRPC_ADDR", ":50051"),
		getenv("JWT_PUBLIC_KEY_PEM", "../auth-service/keys/public.pem"),
		getenv("JWT_ISSUER", "todo-auth"),
		getenv("JWT_AUDIENCE", "todo-api"),
		getenv("ENABLE_GRPC_REFLECTION", "false") == "true",
		getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"),
	)

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
