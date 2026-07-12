package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type WorkspaceMemberRepository struct {
	queries *sqlc.Queries
}

func NewWorkspaceMemberRepository(queries *sqlc.Queries) *WorkspaceMemberRepository {
	return &WorkspaceMemberRepository{
		queries: queries,
	}
}

func (r *WorkspaceMemberRepository) AddWorkspaceMember(ctx context.Context, workspaceID, userID int64, role string) (sqlc.WorkspaceMember, error) {
	params := sqlc.AddWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        sqlc.UserRole(role),
	}

	member, err := r.queries.AddWorkspaceMember(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				switch pgErr.ConstraintName {
				case "workspace_members_pkey":
					return sqlc.WorkspaceMember{}, ErrWorkspaceMemberAlreadyExists
				}
			case "23503":
				switch pgErr.ConstraintName {
				case "workspace_members_workspace_id_fkey":
					return sqlc.WorkspaceMember{}, ErrWorkspaceNotFound
				case "workspace_members_user_id_fkey":
					return sqlc.WorkspaceMember{}, ErrUserNotFound
				}
			}
		}
		return sqlc.WorkspaceMember{}, fmt.Errorf("add workspace member: %w", err)
	}
	return member, nil
}

func (r *WorkspaceMemberRepository) GetWorkspaceMember(ctx context.Context, workspaceID, userID int64) (sqlc.GetWorkspaceMemberRow, error) {
	params := sqlc.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}

	workspaceMember, err := r.queries.GetWorkspaceMember(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetWorkspaceMemberRow{}, ErrWorkspaceMemberNotFound
		}
		return sqlc.GetWorkspaceMemberRow{}, fmt.Errorf("get workspace member: %w", err)
	}
	return workspaceMember, nil
}

func (r *WorkspaceMemberRepository) IsWorkspaceMember(ctx context.Context, workspaceID, userID int64) (sqlc.UserRole, error) {
	params := sqlc.IsWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}

	role, err := r.queries.IsWorkspaceMember(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrWorkspaceMemberNotFound
		}
		return "", fmt.Errorf("check workspace membership: %w", err)
	}
	return role, nil
}

func (r *WorkspaceMemberRepository) ListUserWorkspaces(ctx context.Context, userID int64, search string) ([]sqlc.ListUserWorkspacesRow, error) {
	params := sqlc.ListUserWorkspacesParams{
		UserID: userID,
		Search: &search,
	}

	workspaces, err := r.queries.ListUserWorkspaces(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list user workspaces: %w", err)
	}

	return workspaces, nil
}

func (r *WorkspaceMemberRepository) ListWorkspaceMembers(ctx context.Context, workspaceID int64) ([]sqlc.ListWorkspaceMembersRow, error) {
	members, err := r.queries.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list workspace members: %w", err)
	}

	return members, nil
}

func (r *WorkspaceMemberRepository) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID int64) (sqlc.RemoveWorkspaceMemberRow, error) {
	params := sqlc.RemoveWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}

	removedMember, err := r.queries.RemoveWorkspaceMember(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.RemoveWorkspaceMemberRow{}, ErrWorkspaceMemberNotFound
		}
		return sqlc.RemoveWorkspaceMemberRow{}, fmt.Errorf("remove workspace member: %w", err)
	}

	return removedMember, nil
}
