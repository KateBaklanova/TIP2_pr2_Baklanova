package http

import (
	"kate/services/auth/internal/http/handler"
	"kate/services/auth/internal/service"
	"kate/shared/middleware"
	"log"
	"net/http"
)

func StartServer(port string, authSvc *service.AuthService) {

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/login", handler.LoginHandler(authSvc))
	mux.HandleFunc("/v1/auth/verify", handler.VerifyHandler(authSvc))

	handlerWithMiddleware := middleware.RequestIDMiddleware(
		middleware.LoggingMiddleware(mux),
	)

	log.Printf("Auth HTTP server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handlerWithMiddleware); err != nil {
		log.Fatal(err)
	}
}
