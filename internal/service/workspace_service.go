package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/redis"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

type WorkspaceRepo interface {
	CreateWorkspace(ctx context.Context, name string) (sqlc.CreateWorkspaceRow, error)
	GetWorkspaceByID(ctx context.Context, workspaceID int64) (sqlc.GetWorkspaceByIDRow, error)
	GetWorkspaceByName(ctx context.Context, name string) (sqlc.GetWorkspaceByNameRow, error)
	UpdateWorkspaceName(ctx context.Context, name string, id int64) (sqlc.UpdateWorkspaceNameRow, error)
	DeleteWorkspace(ctx context.Context, id int64) (int64, error)
	RestoreDeletedWorkspace(ctx context.Context, id int64) (sqlc.RestoreDeletedWorkspaceRow, error)
}

type WorkspaceMemRepo interface {
	AddWorkspaceMember(ctx context.Context, workspaceID, userID int64, role string) (sqlc.WorkspaceMember, error)
}

type WorkspaceService struct {
	repo                WorkspaceRepo
	workspaceCache      redis.WorkspaceCache
	authz               AuthzService
	workspaceMemberRepo WorkspaceMemRepo
}

func NewWorkspaceService(repo WorkspaceRepo, workspaceCache redis.WorkspaceCache, workspaceMemRepo WorkspaceMemRepo, authz AuthzService) *WorkspaceService {
	return &WorkspaceService{
		repo:                repo,
		workspaceCache:      workspaceCache,
		workspaceMemberRepo: workspaceMemRepo,
		authz:               authz,
	}
}

func (s *WorkspaceService) invalidateWorkspaceCache(ctx context.Context, workspaceID int64) {
	if err := s.workspaceCache.DeleteWorkspace(ctx, workspaceID); err != nil {
		log.Printf("failed to invalidate workspace cache: %v", err)
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

	cached, found, err := s.workspaceCache.GetWorkspace(ctx, workspaceID)
	if err != nil {
		log.Printf("redis get failed: %v", err)
	}
	if found {
		return cached, nil
	}

	workspace, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNotFound
		}
		return dto.WorkspaceResponse{}, fmt.Errorf("get workspace by id: %w", err)
	}

	response := dto.WorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		CreatedAt: workspace.CreatedAt.Time,
		UpdatedAt: workspace.UpdatedAt.Time,
	}

	if err := s.workspaceCache.SetWorkspace(ctx, response); err != nil {
		log.Printf("failed to cache workspace: %v", err)
	}

	return response, nil
}

func (s *WorkspaceService) GetWorkspaceByName(ctx context.Context, name string) (dto.WorkspaceResponse, error) {
	workspace, err := s.repo.GetWorkspaceByName(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNotFound
		}
		return dto.WorkspaceResponse{}, fmt.Errorf("get workspace by name: %w", err)
	}

	response := dto.WorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		CreatedAt: workspace.CreatedAt.Time,
		UpdatedAt: workspace.UpdatedAt.Time,
	}

	if err := s.workspaceCache.SetWorkspace(ctx, response); err != nil {
		log.Printf("failed to cache worksapce: %v", err)
	}

	return response, nil
}

func (s *WorkspaceService) UpdateWorkspaceName(ctx context.Context, requesterID, workspaceID int64, req dto.UpdateWorkspaceRequest) (dto.WorkspaceResponse, error) {
	if workspaceID <= 0 {
		return dto.WorkspaceResponse{}, ErrInvalidWorkspaceID
	}

	dbWorkspace, err := s.repo.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNotFound
		}
		return dto.WorkspaceResponse{}, fmt.Errorf("get workspace by id: %w", err)
	}

	workspaceName := dbWorkspace.Name

	if req.Name != nil {
		workspaceName = strings.TrimSpace(*req.Name)

		if utf8.RuneCountInString(workspaceName) < 3 || utf8.RuneCountInString(workspaceName) > 30 {
			return dto.WorkspaceResponse{}, ErrInvalidWorkspaceName
		}
	}

	if err := s.authz.RequireWorkspaceAdmin(ctx, dbWorkspace.ID, requesterID); err != nil {
		return dto.WorkspaceResponse{}, err
	}

	workspace, err := s.repo.UpdateWorkspaceName(ctx, workspaceName, dbWorkspace.ID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNotFound
		}
		if errors.Is(err, repository.ErrWorkspaceAlreadyExists) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNameTaken
		}
		return dto.WorkspaceResponse{}, fmt.Errorf("update workspace name: %w", err)
	}

	s.invalidateWorkspaceCache(ctx, workspace.ID)

	return dto.WorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		CreatedAt: workspace.CreatedAt.Time,
		UpdatedAt: workspace.UpdatedAt.Time,
	}, nil
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, requesterID, workspaceID int64) (int64, error) {
	if workspaceID <= 0 {
		return 0, ErrInvalidWorkspaceID
	}

	if err := s.authz.RequireWorkspaceAdmin(ctx, workspaceID, requesterID); err != nil {
		return 0, err
	}

	id, err := s.repo.DeleteWorkspace(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return 0, ErrWorkspaceNotFound
		}
		return 0, fmt.Errorf("delete workspace: %w", err)
	}

	s.invalidateWorkspaceCache(ctx, workspaceID)

	return id, nil
}

func (s *WorkspaceService) RestoreDeletedWorkspace(ctx context.Context, requesterID, workspaceID int64) (dto.WorkspaceResponse, error) {
	if workspaceID <= 0 {
		return dto.WorkspaceResponse{}, ErrInvalidWorkspaceID
	}

	if err := s.authz.RequireWorkspaceAdminIncludingDeleted(ctx, workspaceID, requesterID); err != nil {
		return dto.WorkspaceResponse{}, err
	}

	workspace, err := s.repo.RestoreDeletedWorkspace(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceResponse{}, ErrWorkspaceNotFound
		}
		return dto.WorkspaceResponse{}, fmt.Errorf("restore deleted workspace: %w", err)
	}

	s.invalidateWorkspaceCache(ctx, workspace.ID)

	response := dto.WorkspaceResponse{
		ID:        workspace.ID,
		Name:      workspace.Name,
		CreatedAt: workspace.CreatedAt.Time,
		UpdatedAt: workspace.UpdatedAt.Time,
	}
	if err := s.workspaceCache.SetWorkspace(ctx, response); err != nil {
		log.Printf("failed to cache restored workspace: %v", err)
	}

	return response, nil
}
