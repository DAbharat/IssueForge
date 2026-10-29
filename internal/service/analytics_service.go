package service

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

type AnalyticsService struct {
	repo  *repository.AnalyticsRepository
	authz AuthzService
}

func NewAnalyticsService(repo *repository.AnalyticsRepository, authz AuthzService) *AnalyticsService {
	return &AnalyticsService{
		repo:  repo,
		authz: authz,
	}
}

func (s *AnalyticsService) GetWorkspaceAnalytics(
	ctx context.Context,
	requesterID, workspaceID int64,
	from, to *time.Time,
	memberLimit int32,
) (dto.WorkspaceAnalyticsResponse, error) {
	if workspaceID <= 0 {
		return dto.WorkspaceAnalyticsResponse{}, ErrInvalidWorkspaceID
	}

	if err := s.authz.RequireWorkspaceAdmin(ctx, workspaceID, requesterID); err != nil {
		return dto.WorkspaceAnalyticsResponse{}, err
	}

	if _, err := s.repo.GetWorkspaceByID(ctx, workspaceID); err != nil {
		if errors.Is(err, repository.ErrWorkspaceNotFound) {
			return dto.WorkspaceAnalyticsResponse{}, ErrWorkspaceNotFound
		}
		return dto.WorkspaceAnalyticsResponse{}, fmt.Errorf("verify workspace: %w", err)
	}

	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	if from != nil && !from.IsZero() {
		startDate = from.UTC()
	}
	if to != nil && !to.IsZero() {
		endDate = to.UTC()
	}
	if memberLimit <= 0 || memberLimit > 50 {
		memberLimit = 10
	}

	dailyRows, err := s.repo.GetWorkspaceDailyActivityStats(ctx, workspaceID, startDate, endDate)
	if err != nil {
		return dto.WorkspaceAnalyticsResponse{}, fmt.Errorf("get daily activity stats: %w", err)
	}

	velocityRows, err := s.repo.GetWorkspaceIssueVelocityStats(ctx, workspaceID, startDate, endDate)
	if err != nil {
		return dto.WorkspaceAnalyticsResponse{}, fmt.Errorf("get issue velocity stats: %w", err)
	}

	breakdownRows, err := s.repo.GetWorkspaceActivityTypeBreakdown(ctx, workspaceID, startDate)
	if err != nil {
		return dto.WorkspaceAnalyticsResponse{}, fmt.Errorf("get activity breakdown: %w", err)
	}

	memberRows, err := s.repo.GetWorkspaceMemberActivityStats(ctx, workspaceID, startDate, memberLimit)
	if err != nil {
		return dto.WorkspaceAnalyticsResponse{}, fmt.Errorf("get member activity stats: %w", err)
	}

	dailyStats := make([]dto.DailyActivityStat, 0, len(dailyRows))
	for _, r := range dailyRows {
		dateStr := ""
		if r.EventDate.Valid {
			dateStr = r.EventDate.Time.Format("2006-01-02")
		}
		dailyStats = append(dailyStats, dto.DailyActivityStat{
			Date:            dateStr,
			TotalActivities: r.TotalActivities,
		})
	}

	velocityStats := make([]dto.IssueVelocityStat, 0, len(velocityRows))
	for _, r := range velocityRows {
		dateStr := ""
		if r.Date.Valid {
			dateStr = r.Date.Time.Format("2006-01-02")
		}
		velocityStats = append(velocityStats, dto.IssueVelocityStat{
			Date:            dateStr,
			IssuesCreated:   r.IssuesCreated,
			IssuesCompleted: r.IssuesCompleted,
		})
	}

	breakdown := make([]dto.ActivityTypeBreakdown, 0, len(breakdownRows))
	for _, r := range breakdownRows {
		breakdown = append(breakdown, dto.ActivityTypeBreakdown{
			ActivityType: r.ActivityType,
			Count:        r.Count,
		})
	}

	leaderboard := make([]dto.MemberActivityStat, 0, len(memberRows))
	for _, r := range memberRows {
		leaderboard = append(leaderboard, dto.MemberActivityStat{
			UserID:               r.UserID,
			Fullname:             r.Fullname,
			Username:             r.Username,
			Email:                r.Email,
			IssueActionsCount:    r.IssueActionsCount,
			CompletedIssuesCount: r.CompletedIssuesCount,
		})
	}

	return dto.WorkspaceAnalyticsResponse{
		DailyActivities:   dailyStats,
		IssueVelocity:     velocityStats,
		ActivityBreakdown: breakdown,
		MemberLeaderboard: leaderboard,
	}, nil
}

func (s *AnalyticsService) GetProjectAnalytics(
	ctx context.Context,
	requesterID, projectID int64,
	from, to *time.Time,
	memberLimit int32,
) (dto.ProjectAnalyticsResponse, error) {
	if projectID <= 0 {
		return dto.ProjectAnalyticsResponse{}, ErrInvalidProjectID
	}

	proj, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectAnalyticsResponse{}, ErrProjectNotFound
		}
		return dto.ProjectAnalyticsResponse{}, fmt.Errorf("verify project: %w", err)
	}

	// Authorization: Project lead or workspace admin
	isLeadErr := s.authz.RequireProjectLead(ctx, projectID, requesterID)
	isAdminErr := s.authz.RequireWorkspaceAdmin(ctx, proj.WorkspaceID, requesterID)
	if isLeadErr != nil && isAdminErr != nil {
		return dto.ProjectAnalyticsResponse{}, ErrForbidden
	}

	now := time.Now().UTC()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	if from != nil && !from.IsZero() {
		startDate = from.UTC()
	}
	if to != nil && !to.IsZero() {
		endDate = to.UTC()
	}
	if memberLimit <= 0 || memberLimit > 50 {
		memberLimit = 10
	}

	dailyRows, err := s.repo.GetProjectDailyActivityStats(ctx, projectID, startDate, endDate)
	if err != nil {
		return dto.ProjectAnalyticsResponse{}, fmt.Errorf("get daily activity stats: %w", err)
	}

	velocityRows, err := s.repo.GetProjectIssueVelocityStats(ctx, projectID, startDate, endDate)
	if err != nil {
		return dto.ProjectAnalyticsResponse{}, fmt.Errorf("get issue velocity stats: %w", err)
	}

	breakdownRows, err := s.repo.GetProjectActivityTypeBreakdown(ctx, projectID, startDate)
	if err != nil {
		return dto.ProjectAnalyticsResponse{}, fmt.Errorf("get activity breakdown: %w", err)
	}

	memberRows, err := s.repo.GetProjectMemberActivityStats(ctx, projectID, startDate, memberLimit)
	if err != nil {
		return dto.ProjectAnalyticsResponse{}, fmt.Errorf("get member activity stats: %w", err)
	}

	dailyStats := make([]dto.DailyActivityStat, 0, len(dailyRows))
	for _, r := range dailyRows {
		dateStr := ""
		if r.EventDate.Valid {
			dateStr = r.EventDate.Time.Format("2006-01-02")
		}
		dailyStats = append(dailyStats, dto.DailyActivityStat{
			Date:            dateStr,
			TotalActivities: r.TotalActivities,
		})
	}

	velocityStats := make([]dto.IssueVelocityStat, 0, len(velocityRows))
	for _, r := range velocityRows {
		dateStr := ""
		if r.Date.Valid {
			dateStr = r.Date.Time.Format("2006-01-02")
		}
		velocityStats = append(velocityStats, dto.IssueVelocityStat{
			Date:            dateStr,
			IssuesCreated:   r.IssuesCreated,
			IssuesCompleted: r.IssuesCompleted,
		})
	}

	breakdown := make([]dto.ActivityTypeBreakdown, 0, len(breakdownRows))
	for _, r := range breakdownRows {
		breakdown = append(breakdown, dto.ActivityTypeBreakdown{
			ActivityType: r.ActivityType,
			Count:        r.Count,
		})
	}

	leaderboard := make([]dto.MemberActivityStat, 0, len(memberRows))
	for _, r := range memberRows {
		leaderboard = append(leaderboard, dto.MemberActivityStat{
			UserID:               r.UserID,
			Fullname:             r.Fullname,
			Username:             r.Username,
			Email:                r.Email,
			IssueActionsCount:    r.IssueActionsCount,
			CompletedIssuesCount: r.CompletedIssuesCount,
		})
	}

	return dto.ProjectAnalyticsResponse{
		DailyActivities:   dailyStats,
		IssueVelocity:     velocityStats,
		ActivityBreakdown: breakdown,
		MemberLeaderboard: leaderboard,
	}, nil
}
