package service

import (
	"IssueForge/internal/auth"
	"context"
	"errors"
	"fmt"
)

type AuthorizationRepo interface {
	IsWorkspaceMember(ctx context.Context, workspaceID, userID int64) (auth.UserRole, error)
	IsProjectLead(ctx context.Context, projectID, userID int64) (bool, error)
	IsProjectMember(ctx context.Context, projectID, userID int64) (bool, error)
}

type AuthorizationService struct {
	repo AuthorizationRepo
}

func NewAuthorizationService(repo AuthorizationRepo) *AuthorizationService {
	return &AuthorizationService{
		repo: repo,
	}
}

func (a *AuthorizationService) RequireWorkspaceAdmin(ctx context.Context, workspaceID, userID int64) error {
	role, err := a.repo.IsWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, auth.ErrMembershipNotFound) {
			return auth.ErrForbidden
		}
		return fmt.Errorf("get workspace role: %w", err)
	}

	if role != auth.RoleAdmin {
		return auth.ErrForbidden
	}

	return nil
}

func (a *AuthorizationService) RequireWorkspaceMember(ctx context.Context, workspaceID, userID int64) error {
	_, err := a.repo.IsWorkspaceMember(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, auth.ErrMembershipNotFound) {
			return auth.ErrForbidden
		}
		return fmt.Errorf("check workspace member: %w", err)
	}

	return nil
}

func (a *AuthorizationService) RequireProjectLead(ctx context.Context, projectID, userID int64) error {
	ok, err := a.repo.IsProjectLead(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("check project lead: %w", err)
	}

	if !ok {
		return auth.ErrForbidden
	}

	return nil
}

func (a *AuthorizationService) RequireProjectMember(ctx context.Context, projectID, userID int64) error {
	ok, err := a.repo.IsProjectMember(ctx, projectID, userID)
	if err != nil {
		return fmt.Errorf("check project member: %w", err)
	}

	if !ok {
		return auth.ErrForbidden
	}

	return nil
}
