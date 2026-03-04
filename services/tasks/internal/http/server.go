package http

import (
	"kate/services/tasks/internal/client"
	"kate/services/tasks/internal/http/handler"
	"kate/services/tasks/internal/service"
	"kate/shared/middleware"
	"log"
	"net/http"
)

func StartServer(port string, authGrpcAddr string) {
	taskSvc := service.NewTaskService()

	authGrpc, err := client.NewAuthGrpcClient(authGrpcAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Auth gRPC: %v", err)
	}
	defer authGrpc.Close()

	taskHandler := handler.NewTaskHandler(taskSvc, authGrpc)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.GetAllTasks(w, r)
		case http.MethodPost:
			taskHandler.CreateTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.GetTaskByID(w, r)
		case http.MethodPatch:
			taskHandler.UpdateTask(w, r)
		case http.MethodDelete:
			taskHandler.DeleteTask(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	handlerWithMiddleware := middleware.RequestIDMiddleware(
		middleware.LoggingMiddleware(mux),
	)

	log.Printf("Tasks service starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handlerWithMiddleware); err != nil {
		log.Fatal(err)
	}
}
