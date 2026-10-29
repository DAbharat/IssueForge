package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
)

type ProjectActivityRepo interface {
	CreateProjectActivity(ctx context.Context, projectID, actorID int64, targetUserID *int64, activityType string, fieldName, oldValue, newValue *string) (sqlc.ProjectActivity, error)
	ListProjectActivities(ctx context.Context, projectID int64, limit, offset int32) ([]sqlc.ListProjectActivitiesRow, error)
}

type ProjectActivityRecorder interface {
	CreateProjectActivity(ctx context.Context, projectID, actorID int64, targetUserID *int64, activityType string, fieldName, oldValue, newValue *string) (dto.ProjectActivityResponse, error)
}

type ProjectActivityService struct {
	repo  ProjectActivityRepo
	authz AuthzService
}

func NewProjectActivityService(repo ProjectActivityRepo, authz AuthzService) *ProjectActivityService {
	return &ProjectActivityService{
		repo:  repo,
		authz: authz,
	}
}

func (s *ProjectActivityService) CreateProjectActivity(
	ctx context.Context,
	projectID, actorID int64,
	targetUserID *int64,
	activityType string,
	fieldName, oldValue, newValue *string,
) (dto.ProjectActivityResponse, error) {
	if projectID <= 0 {
		return dto.ProjectActivityResponse{}, ErrInvalidProjectID
	}
	if actorID <= 0 {
		return dto.ProjectActivityResponse{}, ErrInvalidUserID
	}
	if targetUserID != nil && *targetUserID <= 0 {
		return dto.ProjectActivityResponse{}, ErrInvalidUserID
	}

	switch activityType {
	case "PROJECT_CREATED",
		"PROJECT_DETAILS_UPDATED",
		"PROJECT_MEMBER_ADDED",
		"PROJECT_MEMBER_REMOVED",
		"NEW_LEAD_ASSIGNED",
		"PROJECT_ARCHIVED",
		"PROJECT_DELETED",
		"PROJECT_RESTORED",
		"NEW_ISSUE_ADDED":
	default:
		return dto.ProjectActivityResponse{}, ErrInvalidActivityType
	}

	activity, err := s.repo.CreateProjectActivity(ctx, projectID, actorID, targetUserID, activityType, fieldName, oldValue, newValue)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectActivityResponse{}, ErrProjectNotFound
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.ProjectActivityResponse{}, ErrUserNotFound
		}
		if errors.Is(err, repository.ErrInvalidActivityType) {
			return dto.ProjectActivityResponse{}, ErrInvalidActivityType
		}
		return dto.ProjectActivityResponse{}, fmt.Errorf("create project activity: %w", err)
	}

	var target_user_id *int64
	if activity.TargetUserID.Valid {
		target_user_id = &activity.TargetUserID.Int64
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

	return dto.ProjectActivityResponse{
		ID:           activity.ID,
		ProjectID:    activity.ProjectID,
		ActorID:      activity.ActorID,
		TargetUserID: target_user_id,
		ActivityType: activityType,
		FieldName:    field_name,
		OldValue:     old_value,
		NewValue:     new_value,
		CreatedAt:    activity.CreatedAt.Time,
	}, nil
}

func (s *ProjectActivityService) ListProjectActivities(ctx context.Context, requesterID, projectID int64, limit, offset int32) ([]dto.ProjectActivityResponse, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}
	if limit <= 0 {
		return nil, ErrInvalidLimit
	}
	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbActivities, err := s.repo.ListProjectActivities(ctx, projectID, limit, offset)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("list project activities: %w", err)
	}

	activities := make([]dto.ProjectActivityResponse, 0, len(dbActivities))
	for _, a := range dbActivities {
		var target_user_id *int64
		if a.TargetUserID.Valid {
			target_user_id = &a.TargetUserID.Int64
		}
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
		var target_user_name *string
		if a.Fullname_2.Valid {
			target_user_name = &a.Fullname_2.String
		}
		var target_user_username *string
		if a.Username_2.Valid {
			target_user_username = &a.Username_2.String
		}

		activities = append(activities, dto.ProjectActivityResponse{
			ID:                 a.ID,
			ProjectID:          a.ProjectID,
			ActorID:            a.ActorID,
			ActorName:          a.Fullname,
			ActorUsername:      a.Username,
			TargetUserID:       target_user_id,
			TargetUserName:     target_user_name,
			TargetUserUsername: target_user_username,
			ActivityType:       string(a.ActivityType),
			FieldName:          field_name,
			OldValue:           old_value,
			NewValue:           new_value,
			CreatedAt:          a.CreatedAt.Time,
		})
	}

	return activities, nil
}
