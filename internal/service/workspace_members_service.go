package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
)

type WorkspaceMemberRepos interface {
	AddWorkspaceMember(ctx context.Context, workspaceID, userID int64, role sqlc.UserRole) (sqlc.WorkspaceMember, error)
	GetWorkspaceMember(ctx context.Context, workspaceID, userID int64) (sqlc.GetWorkspaceMemberRow, error)
	IsWorkspaceMember(ctx context.Context, workspaceID, userID int64) (sqlc.UserRole, error)
	ListUserWorkspaces(ctx context.Context, userID int64) ([]sqlc.ListUserWorkspacesRow, error)
	ListWorkspaceMembers(ctx context.Context, workspaceID int64) ([]sqlc.ListWorkspaceMembersRow, error)
	RemoveWorkspaceMember(ctx context.Context, workspaceID, userID int64) (sqlc.RemoveWorkspaceMemberRow, error)
}

type WorkspaceMemberService struct {
	repo WorkspaceMemberRepos
}

func NewWorkspaceMemberService(repo WorkspaceMemberRepos) *WorkspaceMemberService {
	return &WorkspaceMemberService{
		repo: repo,
	}
}

func (s *WorkspaceMemberService) AddWorkspaceMember(ctx context.Context, req dto.AddWorkspaceMemberRequest) (dto.WorkspaceMemberResponse, error) {
	member, err := s.repo.AddWorkspaceMember(ctx, req.WorkspaceID, req.UserID, sqlc.UserRole(req.Role))
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberAlreadyExists) {
			return dto.WorkspaceMemberResponse{}, ErrWorkspaceMemberAlreadyExists
		}
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceMemberResponse{}, ErrWorkspaceNotFound
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.WorkspaceMemberResponse{}, ErrUserNotFound
		}
		return dto.WorkspaceMemberResponse{}, fmt.Errorf("add workspace member: %w", err)
	}
	return dto.WorkspaceMemberResponse{
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
		Role:        string(member.Role),
		JoinedAt:    member.JoinedAt.Time,
	}, nil
}

func (s *WorkspaceMemberService) GetWorkSpaceMember(ctx context.Context, workspaceID, userID int64) (dto.WorkspaceMemberSummary, error) {
	member, err := s.repo.GetWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return dto.WorkspaceMemberSummary{}, ErrWorkspaceMemberNotFound
		}
		return dto.WorkspaceMemberSummary{}, fmt.Errorf("get workspace member: %w", err)
	}
	return dto.WorkspaceMemberSummary{
		ID:          member.UserID,
		Fullname:    member.Fullname,
		DisplayName: member.DisplayName,
		Email:       member.Email,
		Role:        string(member.Role),
		JoinedAt:    member.JoinedAt.Time,
	}, nil
}

func (s *WorkspaceMemberService) IsWorkspaceMember(ctx context.Context, workspaceID, userID int64) (sqlc.UserRole, error) {
	member, err := s.repo.IsWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return "", ErrWorkspaceMemberNotFound
		}
		return "", fmt.Errorf("check workspace membership: %w", err)
	}
	return sqlc.UserRole(member), nil
}

func (s *WorkspaceMemberService) ListUserWorkspaces(ctx context.Context, userID int64) ([]dto.WorkspaceSummary, error) {
	workspaces, err := s.repo.ListUserWorkspaces(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user workspace details: %w", err)
	}

	result := make([]dto.WorkspaceSummary, 0, len(workspaces))

	for _, ws := range workspaces {
		result = append(result, dto.WorkspaceSummary{
			ID:   ws.ID,
			Name: ws.Name,
			Role: string(ws.Role),
		})
	}
	return result, nil
}

func (s *WorkspaceMemberService) ListWorkspaceMembers(ctx context.Context, workspaceID int64) ([]dto.WorkspaceMemberDetails, error) {
	members, err := s.repo.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return []dto.WorkspaceMemberDetails{}, fmt.Errorf("list workspace members: %w", err)
	}

	result := make([]dto.WorkspaceMemberDetails, 0, len(members))

	for _, mem := range members {
		result = append(result, dto.WorkspaceMemberDetails{
			UserID:      mem.ID,
			DisplayName: mem.DisplayName,
			Fullname:    mem.Fullname,
			Email:       mem.Email,
			Role:        string(mem.Role),
			JoinedAt:    mem.JoinedAt.Time,
		})
	}
	return result, nil
}

func (s *WorkspaceMemberService) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID int64) (dto.RemoveWorkspaceMemberResponse, error) {
	member, err := s.repo.RemoveWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceMemberNotFound) {
			return dto.RemoveWorkspaceMemberResponse{}, ErrWorkspaceMemberNotFound
		}
		return dto.RemoveWorkspaceMemberResponse{}, fmt.Errorf("remove workspace member: %w", err)
	}
	return dto.RemoveWorkspaceMemberResponse{
		WorkspaceID: member.WorkspaceID,
		UserID:      member.UserID,
	}, nil
}
