package main

import (
	postgres "IssueForge/internal/db"
	"IssueForge/internal/db/config"
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/handler"
	"IssueForge/internal/middleware"
	"IssueForge/internal/redis"
	"IssueForge/internal/repository"
	"IssueForge/internal/router"
	"IssueForge/internal/service"
	"IssueForge/internal/storage"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gorilla/handlers"
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

	redisClient, err := redis.NewClient(cfg)
	if err != nil {
		log.Fatalf("load redis: %v", err)
	}
	projectCache := redis.NewRedisProjectCache(redisClient, cfg.RedisTTL)
	workspaceCache := redis.NewRedisWorkspaceCache(redisClient, cfg.RedisTTL)
	issueCache := redis.NewIssueCache(redisClient, cfg.RedisTTL)

	cld, err := cloudinary.NewFromParams(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	if err != nil {
		log.Fatalf("initialize cloudinary: %v", err)
	}

	queries := sqlc.New(pool)

	userRepo := repository.NewUserRepository(queries)
	workspaceMemberRepo := repository.NewWorkspaceMemberRepository(queries)
	userService := service.NewUserService(userRepo, workspaceMemberRepo, cfg.JWTSecret)
	userHandler := handler.NewUserHandler(userService)

	authzRepo := repository.NewAuthorizationRepository(queries)
	authzService := service.NewAuthorizationService(authzRepo)
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	projectRepo := repository.NewProjectRepository(pool, queries)
	projectService := service.NewProjectService(projectRepo, projectCache, authzService)
	projectHandler := handler.NewProjectHandler(projectService)

	projectMemberRepo := repository.NewProjectMemberRepository(queries)
	projectMemberService := service.NewProjectMemberService(projectMemberRepo, authzService)
	projectMemberHandler := handler.NewProjectMemberHandler(projectMemberService)

	workspaceMemberService := service.NewWorkspaceMemberService(workspaceMemberRepo, authzService)
	workspaceMemberHandler := handler.NewWorkspaceMemberHandler(workspaceMemberService)

	workspaceRepo := repository.NewWorkspaceRepository(queries)
	workspaceService := service.NewWorkspaceService(workspaceRepo, workspaceCache, workspaceMemberRepo, authzService)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)

	issueRepo := repository.NewIssueRepository(queries)

	issueActivityRepo := repository.NewIssueActivityRepository(queries)
	issueActivityService := service.NewIssueActivityService(issueActivityRepo, issueRepo, authzService)

	issueService := service.NewIssueService(issueRepo, issueCache, issueActivityService, authzService)
	issueHandler := handler.NewIssueHandler(issueService)

	commentRepo := repository.NewCommentRepository(queries)
	commentService := service.NewCommentService(commentRepo, issueRepo, issueActivityService, authzService)
	commentHandler := handler.NewCommentHandler(commentService)

	issueActivityHandler := handler.NewIssueActivityHandler(issueActivityService)

	cloudStorage := storage.NewCloudinaryStorage(cld)

	issueAttachmentsRepo := repository.NewIssueAttachmentsRepository(queries)
	issueAttachmentsService := service.NewIssueAttachmentsService(issueAttachmentsRepo, issueRepo, commentRepo, cloudStorage, authzService)
	issueAttachmentsHandler := handler.NewIssueAttachmentsHandler(issueAttachmentsService)

	labelsRepo := repository.NewLabelsRepository(queries)
	labelsService := service.NewLabelsService(labelsRepo, issueRepo, authzService)
	labelsHandler := handler.NewLabelsHandler(labelsService)

	r := router.New(userHandler, projectHandler, projectMemberHandler, workspaceHandler, workspaceMemberHandler, issueHandler, commentHandler, issueActivityHandler, issueAttachmentsHandler, labelsHandler, authMiddleware)

	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		handlers.AllowedMethods([]string{"POST", "GET", "PATCH", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)

	server := &http.Server{
		Addr:              serverAddr,
		Handler:           corsHandler(r),
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
