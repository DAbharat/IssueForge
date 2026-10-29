package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type AnalyticsRepository struct {
	queries *sqlc.Queries
}

func NewAnalyticsRepository(queries *sqlc.Queries) *AnalyticsRepository {
	return &AnalyticsRepository{queries: queries}
}

func (r *AnalyticsRepository) GetWorkspaceByID(ctx context.Context, workspaceID int64) (sqlc.GetWorkspaceByIDRow, error) {
	ws, err := r.queries.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetWorkspaceByIDRow{}, ErrWorkspaceNotFound
		}
		return sqlc.GetWorkspaceByIDRow{}, fmt.Errorf("get workspace by id: %w", err)
	}
	return ws, nil
}

func (r *AnalyticsRepository) GetProjectByID(ctx context.Context, projectID int64) (sqlc.GetProjectByIDRow, error) {
	proj, err := r.queries.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetProjectByIDRow{}, ErrProjectNotFound
		}
		return sqlc.GetProjectByIDRow{}, fmt.Errorf("get project by id: %w", err)
	}
	return proj, nil
}

func (r *AnalyticsRepository) GetWorkspaceDailyActivityStats(ctx context.Context, workspaceID int64, startDate, endDate time.Time) ([]sqlc.GetWorkspaceDailyActivityStatsRow, error) {
	params := sqlc.GetWorkspaceDailyActivityStatsParams{
		WorkspaceID: workspaceID,
		StartDate:   pgtype.Timestamptz{Time: startDate, Valid: true},
		EndDate:     pgtype.Timestamptz{Time: endDate, Valid: true},
	}
	rows, err := r.queries.GetWorkspaceDailyActivityStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get workspace daily activity stats: %w", err)
	}
	return rows, nil
}

func (r *AnalyticsRepository) GetWorkspaceIssueVelocityStats(ctx context.Context, workspaceID int64, startDate, endDate time.Time) ([]sqlc.GetWorkspaceIssueVelocityStatsRow, error) {
	params := sqlc.GetWorkspaceIssueVelocityStatsParams{
		StartDate:   pgtype.Timestamptz{Time: startDate, Valid: true},
		EndDate:     pgtype.Timestamptz{Time: endDate, Valid: true},
		WorkspaceID: workspaceID,
	}
	rows, err := r.queries.GetWorkspaceIssueVelocityStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get workspace issue velocity stats: %w", err)
	}
	return rows, nil
}

func (r *AnalyticsRepository) GetWorkspaceActivityTypeBreakdown(ctx context.Context, workspaceID int64, startDate time.Time) ([]sqlc.GetWorkspaceActivityTypeBreakdownRow, error) {
	params := sqlc.GetWorkspaceActivityTypeBreakdownParams{
		WorkspaceID: workspaceID,
		StartDate:   pgtype.Timestamptz{Time: startDate, Valid: true},
	}
	rows, err := r.queries.GetWorkspaceActivityTypeBreakdown(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get workspace activity type breakdown: %w", err)
	}
	return rows, nil
}

func (r *AnalyticsRepository) GetWorkspaceMemberActivityStats(ctx context.Context, workspaceID int64, startDate time.Time, limit int32) ([]sqlc.GetWorkspaceMemberActivityStatsRow, error) {
	params := sqlc.GetWorkspaceMemberActivityStatsParams{
		WorkspaceID: workspaceID,
		StartDate:   pgtype.Timestamptz{Time: startDate, Valid: true},
		PageLimit:   limit,
	}
	rows, err := r.queries.GetWorkspaceMemberActivityStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get workspace member activity stats: %w", err)
	}
	return rows, nil
}

func (r *AnalyticsRepository) GetProjectDailyActivityStats(ctx context.Context, projectID int64, startDate, endDate time.Time) ([]sqlc.GetProjectDailyActivityStatsRow, error) {
	params := sqlc.GetProjectDailyActivityStatsParams{
		ProjectID: projectID,
		StartDate: pgtype.Timestamptz{Time: startDate, Valid: true},
		EndDate:   pgtype.Timestamptz{Time: endDate, Valid: true},
	}
	rows, err := r.queries.GetProjectDailyActivityStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get project daily activity stats: %w", err)
	}
	return rows, nil
}

func (r *AnalyticsRepository) GetProjectIssueVelocityStats(ctx context.Context, projectID int64, startDate, endDate time.Time) ([]sqlc.GetProjectIssueVelocityStatsRow, error) {
	params := sqlc.GetProjectIssueVelocityStatsParams{
		StartDate: pgtype.Timestamptz{Time: startDate, Valid: true},
		EndDate:   pgtype.Timestamptz{Time: endDate, Valid: true},
		ProjectID: projectID,
	}
	rows, err := r.queries.GetProjectIssueVelocityStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get project issue velocity stats: %w", err)
	}
	return rows, nil
}

func (r *AnalyticsRepository) GetProjectActivityTypeBreakdown(ctx context.Context, projectID int64, startDate time.Time) ([]sqlc.GetProjectActivityTypeBreakdownRow, error) {
	params := sqlc.GetProjectActivityTypeBreakdownParams{
		ProjectID: projectID,
		StartDate: pgtype.Timestamptz{Time: startDate, Valid: true},
	}
	rows, err := r.queries.GetProjectActivityTypeBreakdown(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get project activity type breakdown: %w", err)
	}
	return rows, nil
}

func (r *AnalyticsRepository) GetProjectMemberActivityStats(ctx context.Context, projectID int64, startDate time.Time, limit int32) ([]sqlc.GetProjectMemberActivityStatsRow, error) {
	params := sqlc.GetProjectMemberActivityStatsParams{
		ProjectID: projectID,
		StartDate: pgtype.Timestamptz{Time: startDate, Valid: true},
		PageLimit: limit,
	}
	rows, err := r.queries.GetProjectMemberActivityStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get project member activity stats: %w", err)
	}
	return rows, nil
}
