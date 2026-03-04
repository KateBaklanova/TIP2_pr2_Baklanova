package client

import (
	"context"
	"log"
	"strings"
	"time"

	"kate/proto_gen/auth"
	"kate/shared/middleware"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthGrpcClient struct {
	client auth.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthGrpcClient(addr string) (*AuthGrpcClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	client := auth.NewAuthServiceClient(conn)
	return &AuthGrpcClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *AuthGrpcClient) Close() {
	c.conn.Close()
}

// VerifyToken вызывает gRPC метод Verify с дедлайном и прокидывает request-id
func (c *AuthGrpcClient) VerifyToken(ctx context.Context, token string) (bool, string, error) {
	log.Println("[gRPC] calling grpc verify")

	// Получаем request-id для лога
	reqID := middleware.GetRequestID(ctx)
	log.Printf("[gRPC] Request-ID: %s, Token: %s", reqID, token)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Добавляем request-id в метаданные
	if reqID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", reqID)
	}

	req := &auth.VerifyRequest{Token: token}

	log.Println("[gRPC] Sending verify request to Auth service...")

	resp, err := c.client.Verify(ctx, req)

	if err != nil {
		log.Printf("[gRPC] ERROR: %v", err)

		if strings.Contains(err.Error(), "connection refused") {
			log.Println("[gRPC] Auth service is DOWN!")
			return false, "", grpc.Errorf(codes.Unavailable, "auth service unavailable")
		}

		if strings.Contains(err.Error(), "context deadline") {
			log.Println("[gRPC] Auth service TIMEOUT!")
			return false, "", grpc.Errorf(codes.DeadlineExceeded, "auth timeout")
		}

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				log.Println("[gRPC] Token invalid (unauthenticated)")
				return false, "", nil
			case codes.DeadlineExceeded:
				log.Println("[gRPC] Deadline exceeded")
				return false, "", grpc.Errorf(codes.DeadlineExceeded, "auth timeout")
			case codes.Unavailable:
				log.Println("[gRPC] Auth unavailable")
				return false, "", grpc.Errorf(codes.Unavailable, "auth service unavailable")
			default:
				log.Printf("[gRPC] Other error code: %v", st.Code())
				return false, "", err
			}
		}
		return false, "", err
	}

	log.Printf("[gRPC] Success: valid=%v, subject=%s", resp.Valid, resp.Subject)

	return resp.Valid, resp.Subject, nil
}
