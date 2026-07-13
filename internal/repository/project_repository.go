package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewProjectRepository(db *pgxpool.Pool, queries *sqlc.Queries) *ProjectRepository {
	return &ProjectRepository{
		db:      db,
		queries: queries,
	}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, workspaceID, leadID int64, name, description string) (sqlc.CreateProjectRow, error) {
	params := sqlc.CreateProjectParams{
		WorkspaceID: workspaceID,
		LeadID:      leadID,
		Name:        name,
		Description: description,
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.CreateProjectRow{}, fmt.Errorf("create project transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := r.queries.WithTx(tx)

	project, err := qtx.CreateProject(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				switch pgErr.ConstraintName {
				case "projects_workspace_id_name_key":
					return sqlc.CreateProjectRow{}, ErrProjectAlreadyExists
				}
			case "23503":
				switch pgErr.ConstraintName {
				case "projects_workspace_id_fkey":
					return sqlc.CreateProjectRow{}, ErrWorkspaceNotFound
				case "projects_lead_id_fkey":
					return sqlc.CreateProjectRow{}, ErrUserNotFound
				}
			}
		}
		return sqlc.CreateProjectRow{}, fmt.Errorf("create project: %w", err)
	}

	_, err = qtx.AddMemberToProject(ctx, sqlc.AddMemberToProjectParams{
		ProjectID: project.ID,
		UserID:    leadID,
	})
	if err != nil {
		return sqlc.CreateProjectRow{}, fmt.Errorf("commit transaction: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.CreateProjectRow{}, fmt.Errorf("commit transaction: %w", err)
	}

	return project, nil
}

func (r *ProjectRepository) ListProjectsByLead(ctx context.Context, leadID int64) ([]sqlc.Project, error) {
	projects, err := r.queries.ListProjectsByLead(ctx, leadID)
	if err != nil {
		return []sqlc.Project{}, fmt.Errorf("list projects by lead: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) ListProjectsByWorkspace(ctx context.Context, workspaceID int64) ([]sqlc.Project, error) {
	projects, err := r.queries.ListProjectsByWorkspace(ctx, workspaceID)
	if err != nil {
		return []sqlc.Project{}, fmt.Errorf("list projects by workspace: %w", err)
	}

	return projects, nil
}
