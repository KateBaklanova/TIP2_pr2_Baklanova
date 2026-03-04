package grpc

import (
	"context"
	"log"
	"net"

	"kate/proto_gen/auth"
	"kate/services/auth/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata" // 👈 ДОБАВЬ
	"google.golang.org/grpc/status"
)

type AuthGrpcServer struct {
	authService *service.AuthService
	auth.UnimplementedAuthServiceServer
}

func NewAuthGrpcServer(authSvc *service.AuthService) *AuthGrpcServer {
	return &AuthGrpcServer{authService: authSvc}
}

// Verify реализует gRPC метод
func (s *AuthGrpcServer) Verify(ctx context.Context, req *auth.VerifyRequest) (*auth.VerifyResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if reqIDs := md.Get("x-request-id"); len(reqIDs) > 0 {
			log.Printf("[%s] gRPC Verify called with token: %s", reqIDs[0], req.Token)
		} else {
			log.Printf("[no-req-id] gRPC Verify called with token: %s", req.Token)
		}
	} else {
		log.Printf("[no-metadata] gRPC Verify called with token: %s", req.Token)
	}

	valid, subject := s.authService.VerifyToken(req.Token)
	if !valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return &auth.VerifyResponse{
		Valid:   valid,
		Subject: subject,
	}, nil
}

// Start запускает gRPC сервер
func StartGrpcServer(port string, authSvc *service.AuthService) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	auth.RegisterAuthServiceServer(s, NewAuthGrpcServer(authSvc))

	log.Printf("Auth gRPC server listening on port %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
