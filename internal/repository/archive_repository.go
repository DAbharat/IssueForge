package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type ArchiveRepository struct {
	queries *sqlc.Queries
}

func NewArchiveRepository(queries *sqlc.Queries) *ArchiveRepository {
	return &ArchiveRepository{queries: queries}
}

func (r *ArchiveRepository) GetArchivedCounts(ctx context.Context, userID int64) (sqlc.GetArchivedCountsRow, error) {
	counts, err := r.queries.GetArchivedCounts(ctx, userID)
	if err != nil {
		return sqlc.GetArchivedCountsRow{}, fmt.Errorf("get archived counts: %w", err)
	}
	return counts, nil
}

func (r *ArchiveRepository) ListArchivedWorkspaces(ctx context.Context, userID int64, search *string) ([]sqlc.Workspace, error) {
	var searchText pgtype.Text
	if search != nil && *search != "" {
		searchText = pgtype.Text{String: *search, Valid: true}
	}
	params := sqlc.ListArchivedWorkspacesParams{
		UserID: userID,
		Search: searchText,
	}
	workspaces, err := r.queries.ListArchivedWorkspaces(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list archived workspaces: %w", err)
	}
	return workspaces, nil
}

func (r *ArchiveRepository) ListArchivedProjects(ctx context.Context, userID int64, search *string) ([]sqlc.ListArchivedProjectsRow, error) {
	var searchText pgtype.Text
	if search != nil && *search != "" {
		searchText = pgtype.Text{String: *search, Valid: true}
	}
	params := sqlc.ListArchivedProjectsParams{
		LeadID: userID,
		Search: searchText,
	}
	projects, err := r.queries.ListArchivedProjects(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list archived projects: %w", err)
	}
	return projects, nil
}

func (r *ArchiveRepository) ListArchivedIssues(ctx context.Context, userID int64, search *string) ([]sqlc.ListArchivedIssuesRow, error) {
	var searchText pgtype.Text
	if search != nil && *search != "" {
		searchText = pgtype.Text{String: *search, Valid: true}
	}
	params := sqlc.ListArchivedIssuesParams{
		CreatedBy: userID,
		Search:    searchText,
	}
	issues, err := r.queries.ListArchivedIssues(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list archived issues: %w", err)
	}
	return issues, nil
}

func (r *ArchiveRepository) ListArchivedAttachments(ctx context.Context, userID int64, search *string) ([]sqlc.ListArchivedAttachmentsRow, error) {
	var searchText pgtype.Text
	if search != nil && *search != "" {
		searchText = pgtype.Text{String: *search, Valid: true}
	}
	params := sqlc.ListArchivedAttachmentsParams{
		UploadedBy: userID,
		Search:     searchText,
	}
	attachments, err := r.queries.ListArchivedAttachments(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list archived attachments: %w", err)
	}
	return attachments, nil
}
