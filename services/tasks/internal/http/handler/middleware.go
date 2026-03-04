package handler

import (
	"context"
	"kate/services/tasks/internal/client"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(authCli *client.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := parts[1]
			valid, err := authCli.VerifyToken(r.Context(), token)
			if err != nil || !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			// юзер в контекст
			ctx := context.WithValue(r.Context(), UserContextKey, "student")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
