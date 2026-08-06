package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *ProjectRepository) CreateProject(ctx context.Context, workspaceID, leadID int64, name, description string) (sqlc.Project, error) {
	params := sqlc.CreateProjectParams{
		WorkspaceID: workspaceID,
		LeadID:      leadID,
		Name:        name,
		Description: description,
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.Project{}, fmt.Errorf("create project transaction: %v", err)
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
					return sqlc.Project{}, ErrProjectAlreadyExists
				}
			case "23503":
				switch pgErr.ConstraintName {
				case "projects_workspace_id_fkey":
					return sqlc.Project{}, ErrWorkspaceNotFound
				case "projects_lead_id_fkey":
					return sqlc.Project{}, ErrUserNotFound
				}
			}
		}
		return sqlc.Project{}, fmt.Errorf("create project: %w", err)
	}

	_, err = qtx.AddMemberToProject(ctx, sqlc.AddMemberToProjectParams{
		ProjectID: project.ID,
		UserID:    leadID,
	})
	if err != nil {
		return sqlc.Project{}, fmt.Errorf("commit transaction: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Project{}, fmt.Errorf("commit transaction: %w", err)
	}

	return project, nil
}

func (r *ProjectRepository) GetProjectByID(ctx context.Context, id int64) (sqlc.GetProjectByIDRow, error) {
	project, err := r.queries.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetProjectByIDRow{}, ErrProjectNotFound
		}
		return sqlc.GetProjectByIDRow{}, fmt.Errorf("get project by id: %w", err)
	}
	return project, nil
}

func (r *ProjectRepository) UpdateProjectDetails(ctx context.Context, name, description *string, id int64) (sqlc.Project, error) {
	var projectName pgtype.Text
	var projectDesc pgtype.Text

	if name != nil {
		projectName.String = *name
		projectName.Valid = true
	}
	if description != nil {
		projectDesc.String = *description
		projectDesc.Valid = true
	}

	params := sqlc.UpdateProjectDetailsParams{
		Name:        projectName,
		Description: projectDesc,
		ID:          id,
	}

	project, err := r.queries.UpdateProjectDetails(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				if pgErr.ConstraintName == "projects_workspace_id_name_key" {
					return sqlc.Project{}, ErrProjectAlreadyExists
				}
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Project{}, ErrProjectNotFound
		}
		return sqlc.Project{}, fmt.Errorf("update project details: %w", err)
	}
	return project, nil
}

func (r *ProjectRepository) UpdateProjectLead(ctx context.Context, leadID *int64, id int64) (sqlc.Project, error) {
	var projLeadID pgtype.Int8
	if leadID != nil {
		projLeadID.Int64 = *leadID
		projLeadID.Valid = true
	}

	params := sqlc.UpdateProjectLeadParams{
		LeadID: projLeadID,
		ID:     id,
	}

	project, err := r.queries.UpdateProjectLead(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Project{}, ErrProjectNotFound
		}
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == "projects_lead_id_fkey" {
				return sqlc.Project{}, ErrUserNotFound
			}
		}
		return sqlc.Project{}, fmt.Errorf("update project lead id: %w", err)
	}
	return project, nil
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, id int64) (sqlc.DeleteProjectRow, error) {
	project, err := r.queries.DeleteProject(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.DeleteProjectRow{}, ErrProjectNotFound
		}
		return sqlc.DeleteProjectRow{}, fmt.Errorf("delete project: %w", err)
	}
	return project, nil
}

func (r *ProjectRepository) ListProjectsByLead(ctx context.Context, workspaceID int64, leadID *int64) ([]sqlc.ListProjectsByLeadRow, error) {
	var projectLead pgtype.Int8
	if leadID != nil {
		projectLead.Int64 = *leadID
		projectLead.Valid = true
	}

	params := sqlc.ListProjectsByLeadParams{
		WorkspaceID: workspaceID,
		LeadID:      projectLead,
	}

	projects, err := r.queries.ListProjectsByLead(ctx, params)
	if err != nil {
		return []sqlc.ListProjectsByLeadRow{}, fmt.Errorf("list projects by lead: %w", err)
	}

	return projects, nil
}
