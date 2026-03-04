package handler

import (
	"encoding/json"
	"kate/services/auth/internal/service"
	"net/http"
	"strings"
)

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

func VerifyHandler(svc *service.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(verifyResponse{Valid: false, Error: "unauthorized"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(verifyResponse{Valid: false, Error: "unauthorized"})
			return
		}

		token := parts[1]

		valid, subject := svc.VerifyToken(token)

		if valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(verifyResponse{
				Valid:   true,
				Subject: subject,
			})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(verifyResponse{
				Valid: false,
				Error: "unauthorized",
			})
		}
	}
}
