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

func (r *WorkspaceRepository) CreateWorkspace(ctx context.Context, name string) (sqlc.Workspace, error) {
	workspace, err := r.queries.CreateWorkspace(ctx, name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "workspaces_name_key":
					return sqlc.Workspace{}, ErrWorkspaceAlreadyExists
				}
			}
		}
		return sqlc.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}

	return workspace, nil
}

func (r *WorkspaceRepository) GetWorkspaceByID(ctx context.Context, id int64) (sqlc.Workspace, error) {
	workspace, err := r.queries.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Workspace{}, ErrWorkspaceNotFound
		}
		return sqlc.Workspace{}, fmt.Errorf("get workspace by id: %w", err)
	}

	return workspace, nil
}

func (r *WorkspaceRepository) GetWorkspaceByName(ctx context.Context, name string) (sqlc.Workspace, error) {
	workspace, err := r.queries.GetWorkspaceByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Workspace{}, ErrWorkspaceNotFound
		}
		return sqlc.Workspace{}, fmt.Errorf("get workspace by name: %w", err)
	}

	return workspace, nil
}
