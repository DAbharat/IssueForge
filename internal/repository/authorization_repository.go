package repository

import (
	"IssueForge/internal/auth"
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type AuthorizationRepository struct {
	queries *sqlc.Queries
}

func NewAuthorizationRepository(queries *sqlc.Queries) *AuthorizationRepository {
	return &AuthorizationRepository{
		queries: queries,
	}
}

func (r *AuthorizationRepository) IsWorkspaceMember(ctx context.Context, workspaceID, userID int64) (auth.UserRole, error) {
	params := sqlc.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}
	log.Printf("checking membership: workspaceID=%d userID=%d", workspaceID, userID)

	role, err := r.queries.IsWorkspaceMember(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", auth.ErrMembershipNotFound
		}
		return "", fmt.Errorf("check workspace member: %w", err)
	}
	log.Printf("role=%v err=%v", role, err)
	return auth.UserRole(role), nil
}

func (r *AuthorizationRepository) IsProjectLead(ctx context.Context, projectID, userID int64) (bool, error) {
	params := sqlc.IsProjectLeadParams{
		ID:     projectID,
		LeadID: userID,
	}

	ok, err := r.queries.IsProjectLead(ctx, params)
	if err != nil {
		return false, fmt.Errorf("check project lead: %w", err)
	}
	return ok, nil
}

func (r *AuthorizationRepository) IsProjectMember(ctx context.Context, projectID, userID int64) (bool, error) {
	log.Printf("IsProjectMember: projectID=%d userID=%d", projectID, userID)

	params := sqlc.IsProjectMemberParams{
		ProjectID: projectID,
		UserID:    userID,
	}

	ok, err := r.queries.IsProjectMember(ctx, params)
	log.Printf("IsProjectMember result: ok=%v err=%v", ok, err)

	if err != nil {
		return false, fmt.Errorf("check project member: %w", err)
	}

	return ok, nil
}
