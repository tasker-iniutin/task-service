FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/task-service ./cmd/task-service

FROM alpine:3.21

WORKDIR /app

COPY --from=build /out/task-service /usr/local/bin/task-service

EXPOSE 50051

ENTRYPOINT ["/usr/local/bin/task-service"]
