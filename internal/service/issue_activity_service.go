package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
)

type IssueActivityRepo interface {
	CreateActivity(ctx context.Context, issueID, actorID int64, activityType string, fieldName, oldValue, newValue *string) (sqlc.IssueActivity, error)
	ListIssueActivities(ctx context.Context, issueID int64, limit, offset int32) ([]sqlc.ListIssueActivitiesRow, error)
}

type IssueActivityService struct {
	repo            IssueActivityRepo
	issueLookupRepo IssueLookupRepo
	authz           AuthzService
}

func NewIssueActivityService(repo IssueActivityRepo, issueLookupRepo IssueLookupRepo, authz AuthzService) *IssueActivityService {
	return &IssueActivityService{
		repo:            repo,
		issueLookupRepo: issueLookupRepo,
		authz:           authz,
	}
}

func (s *IssueActivityService) CreateActivity(ctx context.Context, issueID, actorID int64, activityType string, fieldName, oldValue, newValue *string) (dto.ActivityResponse, error) {
	if issueID <= 0 {
		return dto.ActivityResponse{}, ErrInvalidIssueID
	}
	if actorID <= 0 {
		return dto.ActivityResponse{}, ErrInvalidUserID
	}

	switch activityType {
	case "ISSUE_CREATED",
		"ISSUE_DETAILS_UPDATED",
		"ISSUE_STATUS_CHANGED",
		"ISSUE_PRIORITY_CHANGED",
		"ISSUE_ASSIGNEE_CHANGED",
		"ISSUE_DELETED",
		"ISSUE_RESTORED",
		"COMMENT_CREATED",
		"COMMENT_UPDATED",
		"COMMENT_DELETED":
	default:
		return dto.ActivityResponse{}, ErrInvalidActivityType
	}

	activity, err := s.repo.CreateActivity(ctx, issueID, actorID, activityType, fieldName, oldValue, newValue)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.ActivityResponse{}, ErrIssueNotFound
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.ActivityResponse{}, ErrUserNotFound
		}
		if errors.Is(err, repository.ErrInvalidActivityType) {
			return dto.ActivityResponse{}, ErrInvalidActivityType
		}
		return dto.ActivityResponse{}, fmt.Errorf("create activity: %w", err)
	}
	var field_name *string
	if activity.FieldName.Valid {
		field_name = &activity.FieldName.String
	}
	var old_value *string
	if activity.OldValue.Valid {
		old_value = &activity.OldValue.String
	}
	var new_value *string
	if activity.NewValue.Valid {
		new_value = &activity.NewValue.String
	}

	return dto.ActivityResponse{
		ID:           activity.ID,
		IssueID:      activity.IssueID,
		ActorID:      activity.ActorID,
		ActivityType: activityType,
		FieldName:    field_name,
		OldValue:     old_value,
		NewValue:     new_value,
		CreatedAt:    activity.CreatedAt.Time,
	}, nil
}

func (s *IssueActivityService) ListIssueActivities(ctx context.Context, requesterID, issueID int64, limit, offset int32) ([]dto.ActivityResponse, error) {
	if issueID <= 0 {
		return nil, ErrInvalidIssueID
	}
	if limit <= 0 {
		return nil, ErrInvalidLimit
	}
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	projectID, err := s.issueLookupRepo.GetIssueProjectID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("get issue project id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbActivities, err := s.repo.ListIssueActivities(ctx, issueID, limit, offset)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("list project issues: %w", err)
	}

	activity := make([]dto.ActivityResponse, 0, len(dbActivities))

	for _, a := range dbActivities {
		var field_name *string
		if a.FieldName.Valid {
			field_name = &a.FieldName.String
		}
		var old_value *string
		if a.OldValue.Valid {
			old_value = &a.OldValue.String
		}
		var new_value *string
		if a.NewValue.Valid {
			new_value = &a.NewValue.String
		}

		activity = append(activity, dto.ActivityResponse{
			ID:           a.ID,
			IssueID:      a.IssueID,
			ActorID:      a.ActorID,
			ActorName:    a.ActorName,
			ActivityType: string(a.ActivityType),
			FieldName:    field_name,
			OldValue:     old_value,
			NewValue:     new_value,
			CreatedAt:    a.CreatedAt.Time,
		})
	}
	return activity, nil
}
