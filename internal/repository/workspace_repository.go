package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type WorkspaceRepository struct {
	queries *sqlc.Queries
}

func NewWorkspaceRepository(queries *sqlc.Queries) *WorkspaceRepository {
	return &WorkspaceRepository{
		queries: queries,
	}
}

func (r *WorkspaceRepository) CreateWorkspace(ctx context.Context, name string) (sqlc.CreateWorkspaceRow, error) {
	workspace, err := r.queries.CreateWorkspace(ctx, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "workspaces_name_key":
					return sqlc.CreateWorkspaceRow{}, ErrWorkspaceAlreadyExists
				}
			}
		}
		return sqlc.CreateWorkspaceRow{}, fmt.Errorf("create workspace: %w", err)
	}

	return workspace, nil
}

func (r *WorkspaceRepository) GetWorkspaceByID(ctx context.Context, id int64) (sqlc.GetWorkspaceByIDRow, error) {
	workspace, err := r.queries.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetWorkspaceByIDRow{}, ErrWorkspaceNotFound
		}
		return sqlc.GetWorkspaceByIDRow{}, fmt.Errorf("get workspace by id: %w", err)
	}

	return workspace, nil
}

func (r *WorkspaceRepository) GetWorkspaceByName(ctx context.Context, name string) (sqlc.GetWorkspaceByNameRow, error) {
	workspace, err := r.queries.GetWorkspaceByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetWorkspaceByNameRow{}, ErrWorkspaceNotFound
		}
		return sqlc.GetWorkspaceByNameRow{}, fmt.Errorf("get workspace by name: %w", err)
	}

	return workspace, nil
}

func (r *WorkspaceRepository) UpdateWorkspaceName(ctx context.Context, name string, id int64) (sqlc.UpdateWorkspaceNameRow, error) {
	params := sqlc.UpdateWorkspaceNameParams{
		Name: name,
		ID:   id,
	}

	workspace, err := r.queries.UpdateWorkspaceName(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.UpdateWorkspaceNameRow{}, ErrWorkspaceNotFound
		}
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return sqlc.UpdateWorkspaceNameRow{}, ErrWorkspaceAlreadyExists
			}
		}
		return sqlc.UpdateWorkspaceNameRow{}, fmt.Errorf("update workspace name: %w", err)
	}
	return workspace, nil
}

func (r *WorkspaceRepository) DeleteWorkspace(ctx context.Context, id int64) (int64, error) {
	id, err := r.queries.DeleteWorkspace(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrWorkspaceNotFound
		}
		return 0, fmt.Errorf("delete workspace: %w", err)
	}
	return id, nil
}

func (r *WorkspaceRepository) RestoreDeletedWorkspace(ctx context.Context, id int64) (sqlc.RestoreDeletedWorkspaceRow, error) {
	workspace, err := r.queries.RestoreDeletedWorkspace(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.RestoreDeletedWorkspaceRow{}, ErrWorkspaceNotFound
		}
		return sqlc.RestoreDeletedWorkspaceRow{}, fmt.Errorf("restore workspace: %w", err)
	}
	return workspace, nil
}
