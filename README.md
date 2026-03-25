## Task Service

gRPC service for task CRUD. Stores tasks in PostgreSQL and enforces ownership by `user_id`.

## Responsibility

`task-service` is responsible for:

- creating tasks;
- listing tasks for the authenticated user;
- updating task title/text/status;
- deleting tasks.

It does not handle authentication directly. Access tokens are verified via gRPC interceptor.

## Architecture

The service follows a layered structure:

- `cmd/task-service`
  entry point;
- `internal/app`
  bootstrap and dependency wiring;
- `internal/domain`
  models and repository contracts;
- `internal/usecase`
  business logic;
- `internal/store/postgre`
  PostgreSQL repository;
- `internal/transport/grpc`
  gRPC handlers;
- `migrations`
  database schema.

Shared infrastructure lives in `common`:

- `common/postgres`
- `common/runtime`
- `common/authsecurity`
- `common/grpcauth`

## API

The protobuf contract is defined in `api-contracts/proto/task/v1/task.proto`.

Exposed operations:

| Method | Purpose |
| --- | --- |
| `CreateTask` | Create a task |
| `GetTask` | Get a task by id |
| `ListTasks` | List tasks with filters and pagination |
| `UpdateTask` | Update title/text/status |
| `DeleteTask` | Delete a task |

## Storage Design

PostgreSQL stores task data.

Why PostgreSQL:

- task data must be durable;
- indexes are needed for user-scoped queries;
- explicit SQL keeps behavior transparent.

## Database Schema

Migration: `task-service/migrations/001_create_tasks.sql`

Table: `tasks`

- `id`
- `user_id`
- `title`
- `text`
- `status`
- `expires_at` (nullable)

## Configuration

Configuration is provided through environment variables.

Main variables:

- `TASK_GRPC_ADDR`
- `JWT_PUBLIC_KEY_PEM`
- `JWT_ISSUER`
- `JWT_AUDIENCE`
- `DATABASE_URL`

## Local Run

Requirements:

- Go
- Docker / Docker Compose
- `goose`
- RSA public key in PEM format

Start infrastructure:

```bash
cd task-service
make db-up
make migrate-up
```

Run service:

```bash
cd task-service
export JWT_PUBLIC_KEY_PEM=/absolute/path/to/public.pem
go run ./cmd/task-service
```

Defaults:

- PostgreSQL: `localhost:5432`
- gRPC: `:50051`

## Testing

Tests included:

- use case tests: `task-service/internal/usecase/task_usecase_test.go`
- PostgreSQL repo tests: `task-service/internal/store/postgre/task_repo_test.go`

Run:

```bash
GOCACHE=/tmp/go-build go test ./...
```

Repository tests use a real PostgreSQL instance and read:

- `TASK_TEST_DATABASE_URL`, or
- `DATABASE_URL`

If no DSN is set, they skip.

## Current Limitations

- no labels, deadlines, or reminders;
- no audit logging;
- no rate limiting.

## Summary

Main design choices:

- keep business logic in use cases;
- enforce ownership at repo level;
- use explicit SQL with `pgx`;
- keep configuration explicit via env vars.
