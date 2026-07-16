package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

type WorkspaceRepo interface {
	CreateWorkspace(ctx context.Context, name string) (sqlc.Workspace, error)
	GetWorkspaceByID(ctx context.Context, workspaceID int64) (sqlc.Workspace, error)
	GetWorkspaceByName(ctx context.Context, name string) (sqlc.Workspace, error)
}

type WorkspaceMemRepo interface {
	AddWorkspaceMember(ctx context.Context, workspaceID, userID int64, role string) (sqlc.WorkspaceMember, error)
}

type WorkspaceService struct {
	repo                WorkspaceRepo
	authz               AuthzService
	workspaceMemberRepo WorkspaceMemRepo
}

func NewWorkspaceService(repo WorkspaceRepo, workspaceMemRepo WorkspaceMemRepo, authz AuthzService) *WorkspaceService {
	return &WorkspaceService{
		repo:                repo,
		workspaceMemberRepo: workspaceMemRepo,
		authz:               authz,
	}
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, creatorID int64, req dto.CreateWorkspaceRequest) (dto.CreateWorkspaceResponse, error) {
	workspaceName := strings.TrimSpace(req.Name)
	if utf8.RuneCountInString(workspaceName) < 3 || utf8.RuneCountInString(workspaceName) > 30 {
		return dto.CreateWorkspaceResponse{}, ErrInvalidWorkspaceName
	}

	workspace, err := s.repo.CreateWorkspace(ctx, workspaceName)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceAlreadyExists) {
			return dto.CreateWorkspaceResponse{}, ErrWorkspaceNameTaken
		}
		return dto.CreateWorkspaceResponse{}, fmt.Errorf("create workspace: %w", err)
	}

	log.Println("adding creator as an admin")
	member, err := s.workspaceMemberRepo.AddWorkspaceMember(ctx, workspace.ID, creatorID, "ADMIN")
	if err != nil {
		return dto.CreateWorkspaceResponse{}, fmt.Errorf("add creator to workspace: %w", err)
	}
	log.Printf("member=%+v err=%v", member, err)

	return dto.CreateWorkspaceResponse{
		ID:   workspace.ID,
		Name: workspace.Name,
	}, nil
}

func (s *WorkspaceService) GetWorkspaceByID(ctx context.Context, workspaceID, userID int64) (dto.WorkspaceResponse, error) {
	if err := s.authz.RequireWorkspaceMember(ctx, workspaceID, userID); err != nil {
		return dto.WorkspaceResponse{}, err
	}

	workspace, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
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
