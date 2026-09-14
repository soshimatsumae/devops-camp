package main

import (
	"log"
	"net/http"

	"github.com/smatsumae/devops-camp/internal/auth"
	"github.com/smatsumae/devops-camp/internal/config"
	"github.com/smatsumae/devops-camp/internal/db"
	"github.com/smatsumae/devops-camp/internal/handlers"
)

func main() {
	cfg := config.Load()

	conn, err := db.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer conn.Close()

	authHandler := &handlers.AuthHandler{DB: conn, JWTSecret: cfg.JWTSecret, TokenTTL: cfg.TokenTTL}
	taskHandler := &handlers.TaskHandler{DB: conn}
	requireAuth := auth.RequireAuth(cfg.JWTSecret)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.HandleFunc("GET /tasks", requireAuth(taskHandler.List))
	mux.HandleFunc("POST /tasks", requireAuth(taskHandler.Create))
	mux.HandleFunc("GET /tasks/calendar", requireAuth(taskHandler.Calendar))
	mux.HandleFunc("GET /tasks/{id}", requireAuth(taskHandler.Get))
	mux.HandleFunc("PATCH /tasks/{id}", requireAuth(taskHandler.Update))
	mux.HandleFunc("DELETE /tasks/{id}", requireAuth(taskHandler.Delete))

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
