package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type ActivityRepository struct {
	queries *sqlc.Queries
}

func NewIssueActivityRepository(queries *sqlc.Queries) *ActivityRepository {
	return &ActivityRepository{
		queries: queries,
	}
}

func (r *ActivityRepository) CreateActivity(ctx context.Context, issueID, actorID int64, activityType string, fieldName, oldValue, newValue *string) (sqlc.IssueActivity, error) {
	var field_name pgtype.Text
	var old_value pgtype.Text
	var new_value pgtype.Text

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

	params := sqlc.CreateActivityParams{
		IssueID:      issueID,
		ActorID:      actorID,
		ActivityType: sqlc.ActivityType(activityType),
		FieldName:    field_name,
		OldValue:     old_value,
		NewValue:     new_value,
	}

	activity, err := r.queries.CreateActivity(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "issue_activities_issue_id_fkey":
					return sqlc.IssueActivity{}, ErrIssueNotFound
				case "issue_activities_actor_id_fkey":
					return sqlc.IssueActivity{}, ErrUserNotFound
				}
			}
		}
		return sqlc.IssueActivity{}, fmt.Errorf("create activity: %w", err)
	}
	return activity, nil
}

func (r *ActivityRepository) ListIssueActivities(ctx context.Context, issueID int64, limit, offset int32) ([]sqlc.ListIssueActivitiesRow, error) {
	params := sqlc.ListIssueActivitiesParams{
		IssueID: issueID,
		Limit:   limit,
		Offset:  offset,
	}

	activities, err := r.queries.ListIssueActivities(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list issue activities: %w", err)
	}
	return activities, nil
}
