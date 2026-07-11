package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
)

type ProjectMemberRepo interface {
	AddMemberToProject(ctx context.Context, projectID int64, userID int64) (sqlc.ProjectMember, error)
	ListProjectMembers(ctx context.Context, projectID int64) ([]sqlc.ListProjectMembersRow, error)
	SafeAddMemberToProject(ctx context.Context, projectID, userID, leadID int64) (sqlc.ProjectMember, error)
}

type ProjectMemberService struct {
	repo  ProjectMemberRepo
	authz AuthzService
}

func NewProjectMemberService(repo ProjectMemberRepo, authz AuthzService) *ProjectMemberService {
	return &ProjectMemberService{
		repo:  repo,
		authz: authz,
	}
}

func (s *ProjectMemberService) AddMemberToProject(ctx context.Context, req dto.AddProjectMemberRequest) (dto.ProjectMemberResponse, error) {
	member, err := s.repo.AddMemberToProject(ctx, req.ProjectID, req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectMemberAlreadyExists) {
			return dto.ProjectMemberResponse{}, ErrProjectMemberAlreadyExists
		}
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectMemberResponse{}, ErrProjectNotFound
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.ProjectMemberResponse{}, ErrUserNotFound
		}
		return dto.ProjectMemberResponse{}, fmt.Errorf("add project member: %w", err)
	}
	return dto.ProjectMemberResponse{
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		JoinedAt:  member.JoinedAt.Time,
	}, nil
}

func (s *ProjectMemberService) ListProjectMembers(ctx context.Context, projectID, userID int64) ([]dto.ProjectMemberSummary, error) {
	if err := s.authz.RequireProjectMember(ctx, projectID, userID); err != nil {
		return nil, err
	}

	members, err := s.repo.ListProjectMembers(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}

	result := make([]dto.ProjectMemberSummary, 0, len(members))

	for _, member := range members {
		result = append(result, dto.ProjectMemberSummary{
			ID:          member.ID,
			Email:       member.Email,
			Fullname:    member.Fullname,
			DisplayName: member.DisplayName,
			JoinedAt:    member.JoinedAt.Time,
		})
	}
	return result, nil
}

func (s *ProjectMemberService) SafeAddMemberToProject(ctx context.Context, req dto.AddProjectMemberRequest, leadID int64) (dto.ProjectMemberResponse, error) {
	if err := s.authz.RequireProjectLead(ctx, req.ProjectID, leadID); err != nil {
		return dto.ProjectMemberResponse{}, err
	}

	member, err := s.repo.SafeAddMemberToProject(ctx, req.ProjectID, req.UserID, leadID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectMemberValidationFailed) {
			return dto.ProjectMemberResponse{}, ErrProjectMemberValidationFailed
		}
		if errors.Is(err, repository.ErrProjectMemberAlreadyExists) {
			return dto.ProjectMemberResponse{}, ErrProjectMemberAlreadyExists
		}
		return dto.ProjectMemberResponse{}, fmt.Errorf("safe add project member: %w", err)
	}
	return dto.ProjectMemberResponse{
		ProjectID: member.ProjectID,
		UserID:    member.UserID,
		JoinedAt:  member.JoinedAt.Time,
	}, nil
}
