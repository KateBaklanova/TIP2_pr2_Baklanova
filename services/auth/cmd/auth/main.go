package main

import (
	"kate/services/auth/internal/grpc"
	"kate/services/auth/internal/http"
	"kate/services/auth/internal/service"
	"os"
)

func main() {
	httpPort := os.Getenv("AUTH_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	grpcPort := os.Getenv("AUTH_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	authSvc := service.NewAuthService()

	go http.StartServer(httpPort, authSvc)

	grpc.StartGrpcServer(grpcPort, authSvc)
}
