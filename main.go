package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/smatsumae/devops-camp/internal/auth"
	"github.com/smatsumae/devops-camp/internal/config"
	"github.com/smatsumae/devops-camp/internal/db"
	"github.com/smatsumae/devops-camp/internal/handlers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("invalid config: %v", err)
	}

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

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Print("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Print("server stopped")
}
