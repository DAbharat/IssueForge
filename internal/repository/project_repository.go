package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrDuplicateProjectName = errors.New("a project with this name already exists for this user")

type ProjectRepository struct {
	queries *sqlc.Queries
}

func NewProjectRepository(queries *sqlc.Queries) *ProjectRepository {
	return &ProjectRepository{
		queries: queries,
	}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, params sqlc.CreateProjectParams) (sqlc.Project, error) {
	project, err := r.queries.CreateProject(ctx, params)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "projects_owner_id_name_key" {
				return sqlc.Project{}, ErrDuplicateProjectName
			}
		}
		return sqlc.Project{}, fmt.Errorf("create project db failure: %w", err)
	}
	return project, nil
}

func (r *ProjectRepository) ListProjectsByOwner(ctx context.Context, ownerID int64) ([]sqlc.Project, error) {
	projects, err := r.queries.ListProjectsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}
