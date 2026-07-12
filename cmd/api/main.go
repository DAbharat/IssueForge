package main

import (
	postgres "IssueForge/internal/db"
	"IssueForge/internal/db/config"
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"IssueForge/internal/repository"
	"IssueForge/internal/router"
	"IssueForge/internal/service"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}
	log.Println("configuration loaded successfully.")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := ":" + port

	pool, err := postgres.New(cfg)
	if err != nil {
		log.Fatalf("load postgres: %v", err)
	}

	queries := sqlc.New(pool)

	userRepo := repository.NewUserRepository(queries)
	workspaceMemberRepo := repository.NewWorkspaceMemberRepository(queries)
	userService := service.NewUserService(userRepo, workspaceMemberRepo, cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userService)

	authzRepo := repository.NewAuthorizationRepository(queries)
	authzService := service.NewAuthorizationService(authzRepo)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	projectRepo := repository.NewProjectRepository(queries)
	projectService := service.NewProjectService(projectRepo, authzService)
	projectHandler := handler.NewProjectHandler(projectService)

	projectMemberRepo := repository.NewProjectMemberRepository(queries)
	projectMemberService := service.NewProjectMemberService(projectMemberRepo, authzService)
	projectMemberHandler := handler.NewProjectMemberHandler(projectMemberService)

	workspaceMemberService := service.NewWorkspaceMemberService(workspaceMemberRepo, authzService)
	workspaceMemberHandler := handler.NewWorkspaceMemberHandler(workspaceMemberService)

	workspaceRepo := repository.NewWorkspaceRepository(queries)
	workspaceService := service.NewWorkspaceService(workspaceRepo, workspaceMemberRepo, authzService)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)

	r := router.New(userHandler, projectHandler, projectMemberHandler, workspaceHandler, workspaceMemberHandler, authMiddleware)

	server := &http.Server{
		Addr:              serverAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Println("server is running on port:", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server listen error: %v", err)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdownChan
	log.Printf("received signal for shutdown %v", sig)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped gracefully.")
	pool.Close()
	log.Println("disconnected from postgres")
}
