package repository

import "IssueForge/internal/db/sqlc"

type WorkspaceRepository struct {
	queries *sqlc.Queries
}

func NewWorkspaceRepository(queries *sqlc.Queries) *WorkspaceRepository {
	return &WorkspaceRepository{
		queries: queries,
	}
}
