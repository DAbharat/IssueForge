package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ProjectMemberRepository struct {
	queries *sqlc.Queries
}

func NewProjectMemberRepository(queries *sqlc.Queries) *ProjectMemberRepository {
	return &ProjectMemberRepository{
		queries: queries,
	}
}

func (r *ProjectMemberRepository) AddMemberToProject(ctx context.Context, projectID, userID int64) (sqlc.ProjectMember, error) {
	params := sqlc.AddMemberToProjectParams{
		ProjectID: projectID,
		UserID:    userID,
	}

	member, err := r.queries.AddMemberToProject(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				switch pgErr.ConstraintName {
				case "project_members_pkey":
					return sqlc.ProjectMember{}, ErrProjectMemberAlreadyExists
				}
			case "23503":
				switch pgErr.ConstraintName {
				case "project_members_project_id_fkey":
					return sqlc.ProjectMember{}, ErrProjectNotFound
				case "project_members_user_id_fkey":
					return sqlc.ProjectMember{}, ErrUserNotFound
				}
			}
		}
		return sqlc.ProjectMember{}, fmt.Errorf("add member to project: %w", err)
	}

	return member, nil
}

func (r *ProjectMemberRepository) ListProjectMembers(ctx context.Context, projectID int64) ([]sqlc.ListProjectMembersRow, error) {
	members, err := r.queries.ListProjectMembers(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}
	return members, nil
}

func (r *ProjectMemberRepository) SafeAddMemberToProject(ctx context.Context, projectID, userID, leadID int64) (sqlc.ProjectMember, error) {
	params := sqlc.SafeAddMemberToProjectParams{
		ProjectID: projectID,
		UserID:    userID,
		LeadID:    leadID,
	}

	member, err := r.queries.SafeAddMemberToProject(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.ProjectMember{}, ErrProjectMemberValidationFailed
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				switch pgErr.ConstraintName {
				case "project_members_pkey":
					return sqlc.ProjectMember{}, ErrProjectMemberAlreadyExists
				}
			}
		}
		return sqlc.ProjectMember{}, fmt.Errorf("safe add project member: %w", err)
	}

	return member, nil
}
