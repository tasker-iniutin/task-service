module todo/task-service

go 1.24.0

replace github.com/you/todo/api-contracts => ../api-contracts

require (
	github.com/you/todo/api-contracts v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.1
)

require google.golang.org/genproto/googleapis/api v0.0.0-20260209200024-4cfbd4190f57 // indirect

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260209200024-4cfbd4190f57 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
