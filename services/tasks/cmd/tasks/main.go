package main

import (
	"kate/services/tasks/internal/http"
	"os"
)

func main() {
	port := os.Getenv("TASKS_PORT")
	if port == "" {
		port = "8082"
	}

	authGrpcAddr := os.Getenv("AUTH_GRPC_ADDR")
	if authGrpcAddr == "" {
		authGrpcAddr = "localhost:50051"
	}

	http.StartServer(port, authGrpcAddr)
}
