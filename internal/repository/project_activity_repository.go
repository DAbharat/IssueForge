package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProjectActivityRepository struct {
	queries *sqlc.Queries
}

func NewProjectActivityRepository(queries *sqlc.Queries) *ProjectActivityRepository {
	return &ProjectActivityRepository{
		queries: queries,
	}
}

func (r *ProjectActivityRepository) CreateProjectActivity(
	ctx context.Context,
	projectID, actorID int64,
	targetUserID *int64,
	activityType string,
	fieldName, oldValue, newValue *string,
) (sqlc.CreateProjectActivityRow, error) {
	var target_user_id pgtype.Int8
	var field_name pgtype.Text
	var old_value pgtype.Text
	var new_value pgtype.Text

	if targetUserID != nil {
		target_user_id.Int64 = *targetUserID
		target_user_id.Valid = true
	}
	if fieldName != nil {
		field_name.String = *fieldName
		field_name.Valid = true
	}
	if oldValue != nil {
		old_value.String = *oldValue
		old_value.Valid = true
	}
	if newValue != nil {
		new_value.String = *newValue
		new_value.Valid = true
	}

	params := sqlc.CreateProjectActivityParams{
		ProjectID:    projectID,
		ActorID:      actorID,
		TargetUserID: target_user_id,
		ActivityType: sqlc.ProjectActivityType(activityType),
		FieldName:    field_name,
		OldValue:     old_value,
		NewValue:     new_value,
	}

	activity, err := r.queries.CreateProjectActivity(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "project_activities_project_id_fkey":
					return sqlc.CreateProjectActivityRow{}, ErrProjectNotFound
				case "project_activities_actor_id_fkey", "project_activities_target_user_id_fkey":
					return sqlc.CreateProjectActivityRow{}, ErrUserNotFound
				}
			case "22P02":
				return sqlc.CreateProjectActivityRow{}, ErrInvalidActivityType
			}
		}
		return sqlc.CreateProjectActivityRow{}, fmt.Errorf("create project activity: %w", err)
	}

	return activity, nil
}

func (r *ProjectActivityRepository) ListProjectActivities(ctx context.Context, projectID int64, limit, offset int32) ([]sqlc.ListProjectActivitiesRow, error) {
	params := sqlc.ListProjectActivitiesParams{
		ProjectID: projectID,
		Limit:     limit,
		Offset:    offset,
	}

	activities, err := r.queries.ListProjectActivities(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list project activities: %w", err)
	}

	return activities, nil
}
