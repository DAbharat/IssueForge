package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/redis/cache"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

type WorkspaceInvitationsRepository interface {
	CreateWorkspaceInvitation(ctx context.Context, workspaceID, invitedUserID, invitedBy int64) (sqlc.WorkspaceInvitation, error)
	GetWorkspaceInvitation(ctx context.Context, id int64) (sqlc.WorkspaceInvitation, error)
	ListPendingWorkspaceInvitations(ctx context.Context, invitedUserID int64) ([]sqlc.ListPendingWorkspaceInvitationsRow, error)
	ListPendingWorkspaceInvitationsForWorkspace(ctx context.Context, workspaceID int64) ([]sqlc.ListPendingWorkspaceInvitationsForWorkspaceRow, error)
	AcceptInvitation(ctx context.Context, id, invitedUserID int64) (sqlc.AcceptInvitationRow, error)
	DeclineInvitation(ctx context.Context, id, invitedUserID int64) (sqlc.DeclineInvitationRow, error)
	CancelInvitation(ctx context.Context, id, invitedBy int64) (sqlc.CancelInvitationRow, error)
}

type WorkspaceInvitationService struct {
	workspaceInvitationRepo  WorkspaceInvitationsRepository
	workspaceInvitationCache cache.WorkspaceInvitationCache
	authz                    AuthzService
}

func NewWorkspaceInvitationService(workspaceInvitationRepo WorkspaceInvitationsRepository, workspaceInvitationCache cache.WorkspaceInvitationCache, authz AuthzService) *WorkspaceInvitationService {
	return &WorkspaceInvitationService{
		workspaceInvitationRepo:  workspaceInvitationRepo,
		workspaceInvitationCache: workspaceInvitationCache,
		authz:                    authz,
	}
}

func (s *WorkspaceInvitationService) invalidateInvitationUserCache(ctx context.Context, userID int64) {
	if err := s.workspaceInvitationCache.DeletePendingForUser(ctx, userID); err != nil {
		log.Printf("failed to invalidate user invitation cache: %v", err)
	}
}

func (s *WorkspaceInvitationService) invalidateInvitationWorkspaceCache(ctx context.Context, workspaceID int64) {
	if err := s.workspaceInvitationCache.DeletePendingForWorkspace(ctx, workspaceID); err != nil {
		log.Printf("failed to invalidate workspace cache: %v", err)
	}
}

func (s *WorkspaceInvitationService) CreateWorkspaceInvitation(ctx context.Context, workspaceID int64, req dto.CreateWorkspaceInvitationRequest, requesterID int64) (dto.WorkspaceInvitationResponse, error) {
	if workspaceID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidWorkspaceID
	}
	if req.UserID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidUserID
	}

	if err := s.authz.RequireWorkspaceAdmin(ctx, workspaceID, requesterID); err != nil {
		return dto.WorkspaceInvitationResponse{}, err
	}

	invitation, err := s.workspaceInvitationRepo.CreateWorkspaceInvitation(ctx, workspaceID, req.UserID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrPendingInvitationExists) {
			return dto.WorkspaceInvitationResponse{}, ErrPendingInvitationExists
		}
		return dto.WorkspaceInvitationResponse{}, fmt.Errorf("create workspace invitation: %w", err)
	}

	response := dto.WorkspaceInvitationResponse{
		ID:            invitation.ID,
		WorkspaceID:   invitation.WorkspaceID,
		InvitedUserID: invitation.InvitedUserID,
		InvitedBy:     requesterID,
		Status:        invitation.Status,
		CreatedAt:     invitation.CreatedAt.Time,
	}

	s.invalidateInvitationUserCache(ctx, invitation.InvitedUserID)
	s.invalidateInvitationWorkspaceCache(ctx, invitation.WorkspaceID)

	return response, nil
}

func (s *WorkspaceInvitationService) GetWorkspaceInvitation(ctx context.Context, requesterID, invitationID int64) (dto.WorkspaceInvitationResponse, error) {
	if invitationID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidInvitationID
	}

	invitation, err := s.workspaceInvitationRepo.GetWorkspaceInvitation(ctx, invitationID)
	if err != nil {
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return dto.WorkspaceInvitationResponse{}, ErrInvitationNotFound
		}
		return dto.WorkspaceInvitationResponse{}, fmt.Errorf("get invitation by id: %w", err)
	}

	if requesterID != invitation.InvitedUserID {
		return dto.WorkspaceInvitationResponse{}, err
	}

	return dto.WorkspaceInvitationResponse{
		ID:            invitation.ID,
		WorkspaceID:   invitation.WorkspaceID,
		InvitedUserID: invitation.InvitedUserID,
		InvitedBy:     invitation.InvitedBy,
		Status:        invitation.Status,
		CreatedAt:     invitation.CreatedAt.Time,
	}, nil
}

func (s *WorkspaceInvitationService) ListPendingWorkspaceInvitations(ctx context.Context, requesterID int64) ([]dto.PendingWorkspaceInvitationResponse, error) {
	if requesterID <= 0 {
		return nil, ErrInvalidUserID
	}

	cachedInvite, found, err := s.workspaceInvitationCache.GetPendingForUser(ctx, requesterID)
	if err != nil {
		log.Printf("redis get failed: %v", err)
	}
	if found {
		return cachedInvite, nil
	}

	dbInvitations, err := s.workspaceInvitationRepo.ListPendingWorkspaceInvitations(ctx, requesterID)
	if err != nil {
		return nil, fmt.Errorf("list pending workspace invitations: %w", err)
	}

	invitations := make([]dto.PendingWorkspaceInvitationResponse, 0, len(dbInvitations))

	for _, i := range dbInvitations {
		invitations = append(invitations, dto.PendingWorkspaceInvitationResponse{
			ID:              i.ID,
			WorkspaceID:     i.WorkspaceID,
			WorkspaceName:   i.WorkspaceName,
			InviterUsername: i.InviterUsername,
			InviterFullname: i.InviterFullname,
			CreatedAt:       i.CreatedAt.Time,
		})
	}

	if err := s.workspaceInvitationCache.SetPendingForUser(ctx, requesterID, invitations); err != nil {
		log.Printf("redis set failed: %v", err)
	}

	return invitations, nil
}

func (s *WorkspaceInvitationService) ListPendingWorkspaceInvitationsForWorkspace(ctx context.Context, requesterID, workspaceID int64) ([]dto.WorkspacePendingInvitationResponse, error) {
	if workspaceID <= 0 {
		return nil, ErrInvalidWorkspaceID
	}
	if err := s.authz.RequireWorkspaceAdmin(ctx, workspaceID, requesterID); err != nil {
		return nil, err
	}

	cachedInvite, found, err := s.workspaceInvitationCache.GetPendingForWorkspace(ctx, workspaceID)
	if err != nil {
		log.Printf("redis get failed: %v", err)
	}
	if found {
		return cachedInvite, nil
	}

	dbInvitations, err := s.workspaceInvitationRepo.ListPendingWorkspaceInvitationsForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list pending workspace invitations for workspace: %w", err)
	}

	invitations := make([]dto.WorkspacePendingInvitationResponse, 0, len(dbInvitations))

	for _, i := range dbInvitations {
		invitations = append(invitations, dto.WorkspacePendingInvitationResponse{
			ID:            i.ID,
			WorkspaceID:   i.WorkspaceID,
			InvitedUserID: i.InvitedUserID,
			Username:      i.InvitedUsername,
			Fullname:      i.InvitedFullname,
			CreatedAt:     i.CreatedAt.Time,
		})
	}

	if err := s.workspaceInvitationCache.SetPendingForWorkspace(ctx, workspaceID, invitations); err != nil {
		log.Printf("redis set failed: %v", err)
	}

	return invitations, nil
}

func (s *WorkspaceInvitationService) AcceptInvitation(ctx context.Context, invitationID, requesterID int64) (dto.WorkspaceInvitationResponse, error) {
	if invitationID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidInvitationID
	}
	if requesterID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidUserID
	}

	invitation, err := s.workspaceInvitationRepo.AcceptInvitation(ctx, invitationID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return dto.WorkspaceInvitationResponse{}, ErrInvitationNotFound
		}
		if errors.Is(err, repository.ErrWorkspaceMemberAlreadyExists) {
			return dto.WorkspaceInvitationResponse{}, ErrWorkspaceMemberAlreadyExists
		}
		return dto.WorkspaceInvitationResponse{}, fmt.Errorf("accept invitation: %w", err)
	}

	var respondTime *time.Time
	if invitation.RespondedAt.Valid {
		respondTime = &invitation.RespondedAt.Time
	}

	response := dto.WorkspaceInvitationResponse{
		ID:            invitation.ID,
		WorkspaceID:   invitation.WorkspaceID,
		InvitedUserID: invitation.InvitedUserID,
		Status:        invitation.Status,
		RespondedAt:   respondTime,
	}

	s.invalidateInvitationUserCache(ctx, invitation.InvitedUserID)
	s.invalidateInvitationWorkspaceCache(ctx, invitation.WorkspaceID)

	return response, nil
}

func (s *WorkspaceInvitationService) DeclineInvitation(ctx context.Context, invitationID, requesterID int64) (dto.WorkspaceInvitationResponse, error) {
	if invitationID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidInvitationID
	}
	if requesterID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidUserID
	}

	invitation, err := s.workspaceInvitationRepo.DeclineInvitation(ctx, invitationID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return dto.WorkspaceInvitationResponse{}, ErrInvitationNotFound
		}
		return dto.WorkspaceInvitationResponse{}, fmt.Errorf("decline invitation: %w", err)
	}

	var respondTime *time.Time
	if invitation.RespondedAt.Valid {
		respondTime = &invitation.RespondedAt.Time
	}

	response := dto.WorkspaceInvitationResponse{
		ID:            invitation.ID,
		WorkspaceID:   invitation.WorkspaceID,
		InvitedUserID: invitation.InvitedUserID,
		RespondedAt:   respondTime,
		Status:        invitation.Status,
	}

	s.invalidateInvitationUserCache(ctx, invitation.InvitedUserID)
	s.invalidateInvitationWorkspaceCache(ctx, invitation.WorkspaceID)

	return response, nil
}

func (s *WorkspaceInvitationService) CancelInvitation(ctx context.Context, invitationID, requesterID int64) (dto.WorkspaceInvitationResponse, error) {
	if invitationID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidInvitationID
	}
	if requesterID <= 0 {
		return dto.WorkspaceInvitationResponse{}, ErrInvalidUserID
	}

	getInv, err := s.workspaceInvitationRepo.GetWorkspaceInvitation(ctx, invitationID)
	if err != nil {
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return dto.WorkspaceInvitationResponse{}, ErrInvitationNotFound
		}
		return dto.WorkspaceInvitationResponse{}, fmt.Errorf("get workspace invitation: %w", err)
	}

	if err := s.authz.RequireWorkspaceAdmin(ctx, getInv.WorkspaceID, requesterID); err != nil {
		return dto.WorkspaceInvitationResponse{}, err
	}

	invitation, err := s.workspaceInvitationRepo.CancelInvitation(ctx, getInv.ID, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrInvitationNotFound) {
			return dto.WorkspaceInvitationResponse{}, ErrInvitationNotFound
		}
		return dto.WorkspaceInvitationResponse{}, fmt.Errorf("cancel invitation: %w", err)
	}

	var respondTime *time.Time
	if invitation.RespondedAt.Valid {
		respondTime = &invitation.RespondedAt.Time
	}

	response := dto.WorkspaceInvitationResponse{
		ID:          invitation.ID,
		RespondedAt: respondTime,
		Status:      invitation.Status,
	}

	s.invalidateInvitationUserCache(ctx, getInv.InvitedUserID)
	s.invalidateInvitationWorkspaceCache(ctx, getInv.WorkspaceID)

	return response, nil
}
