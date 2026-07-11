package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
)

type WorkspaceRepo interface {
	CreateWorkspace(ctx context.Context, name string) (sqlc.Workspace, error)
	GetWorkspaceByID(ctx context.Context, workspaceID, userID int64) (sqlc.Workspace, error)
	GetWorkspaceByName(ctx context.Context, name string) (sqlc.Workspace, error)
}

type WorkspaceService struct {
	repo  WorkspaceRepo
	authz AuthzService
}

func NewWorkspaceService(repo WorkspaceRepo, authz AuthzService) *WorkspaceService {
	return &WorkspaceService{
		repo:  repo,
		authz: authz,
	}
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req dto.CreateWorkspaceRequest) (dto.CreateWorkspaceResponse, error) {
	workspaceName := strings.TrimSpace(req.Name)
	if len(workspaceName) < 3 || len(workspaceName) > 30 {
		return dto.CreateWorkspaceResponse{}, ErrInvalidWorkspaceName
	}

	workspace, err := s.repo.CreateWorkspace(ctx, workspaceName)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceAlreadyExists) {
			return dto.CreateWorkspaceResponse{}, ErrWorkspaceNameTaken
		}
		return dto.CreateWorkspaceResponse{}, fmt.Errorf("create workspace: %w", err)
	}
	return dto.CreateWorkspaceResponse{
		ID:   workspace.ID,
		Name: workspace.Name,
	}, nil
}

func (s *WorkspaceService) GetWorkspaceByID(ctx context.Context, workspaceID, userID int64) (dto.WorkspaceResponse, error) {
	if err := s.authz.RequireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return dto.WorkspaceResponse{}, err
	}

	workspace, err := s.repo.GetWorkspaceByID(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNotFound
		}
		return dto.WorkspaceResponse{}, fmt.Errorf("get workspace by id: %w", err)
	}
	return dto.WorkspaceResponse{
		ID:   workspace.ID,
		Name: workspace.Name,
	}, nil
}

func (s *WorkspaceService) GetWorkspaceByName(ctx context.Context, name string) (dto.WorkspaceResponse, error) {
	workspace, err := s.repo.GetWorkspaceByName(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNotFound
		}
		return dto.WorkspaceResponse{}, fmt.Errorf("get workspace by name: %w", err)
	}
	return dto.WorkspaceResponse{
		ID:   workspace.ID,
		Name: workspace.Name,
	}, nil
}
