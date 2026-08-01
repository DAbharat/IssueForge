package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type LabelsRepository struct {
	queries *sqlc.Queries
}

func NewLabelsRepository(queries *sqlc.Queries) *LabelsRepository {
	return &LabelsRepository{
		queries: queries,
	}
}

func (r *LabelsRepository) CreateLabel(ctx context.Context, projectID int64, name, color string, createdBy int64) (sqlc.Label, error) {
	params := sqlc.CreateLabelParams{
		ProjectID: projectID,
		Name:      name,
		Color:     color,
		CreatedBy: createdBy,
	}

	label, err := r.queries.CreateLabel(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "labels_project_id_fkey":
					return sqlc.Label{}, ErrProjectNotFound
				case "labels_created_by_fkey":
					return sqlc.Label{}, ErrUserNotFound
				}
			case "23505":
				if pgErr.ConstraintName == "idx_labels_project_name" {
					return sqlc.Label{}, ErrLabelAlreadyExists
				}
			}
		}
		return sqlc.Label{}, fmt.Errorf("create label: %w", err)
	}
	return label, nil
}

func (r *LabelsRepository) GetLabelByID(ctx context.Context, id int64) (sqlc.Label, error) {
	label, err := r.queries.GetLabelByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Label{}, ErrLabelNotFound
		}
		return sqlc.Label{}, fmt.Errorf("get label by id: %w", err)
	}
	return label, nil
}

func (r *LabelsRepository) ListProjectLabels(ctx context.Context, projectID int64) ([]sqlc.Label, error) {
	label, err := r.queries.ListProjectLabels(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project labels: %w", err)
	}
	return label, nil
}

func (r *LabelsRepository) UpdateLabel(ctx context.Context, name, color *string, id int64) (sqlc.Label, error) {
	var labelName pgtype.Text
	var labelColor pgtype.Text

	if name != nil {
		labelName.String = *name
		labelName.Valid = true
	}
	if color != nil {
		labelColor.String = *color
		labelColor.Valid = true
	}

	params := sqlc.UpdateLabelParams{
		Name:  labelName,
		Color: labelColor,
		ID:    id,
	}

	label, err := r.queries.UpdateLabel(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_labels_project_name" {
				return sqlc.Label{}, ErrLabelAlreadyExists
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Label{}, ErrLabelNotFound
		}
		return sqlc.Label{}, fmt.Errorf("update label: %w", err)
	}
	return label, nil
}

func (r *LabelsRepository) DeleteLabel(ctx context.Context, id int64) (sqlc.DeleteLabelRow, error) {
	label, err := r.queries.DeleteLabel(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.DeleteLabelRow{}, ErrLabelNotFound
		}
		return sqlc.DeleteLabelRow{}, fmt.Errorf("delete label: %w", err)
	}
	return label, nil
}

func (r *LabelsRepository) GetLabelProjectID(ctx context.Context, id int64) (int64, error) {
	label, err := r.queries.GetLabelProjectID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrLabelNotFound
		}
		return 0, fmt.Errorf("get label project id: %w", err)
	}
	return label, nil
}

func (r *LabelsRepository) AttachLabelsToIssue(ctx context.Context, issueID int64, labelIDs []int64) error {
	params := sqlc.AttachLabelsToIssueParams{
		IssueID: issueID,
		LabelID: labelIDs,
	}

	err := r.queries.AttachLabelsToIssue(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "issue_labels_issue_id_fkey":
					return ErrIssueNotFound
				case "issue_labels_label_id_fkey":
					return ErrLabelNotFound
				}
			}
		}
		return fmt.Errorf("attach labels to issue: %w", err)
	}
	return nil
}

func (r *LabelsRepository) RemoveLabelFromIssue(ctx context.Context, issueID, labelID int64) (int64, error) {
	params := sqlc.RemoveLabelFromIssueParams{
		IssueID: issueID,
		LabelID: labelID,
	}

	id, err := r.queries.RemoveLabelFromIssue(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrLabelNotAttached
		}
		return 0, fmt.Errorf("remove label from issue: %w", err)
	}
	return id, nil
}

func (r *LabelsRepository) ListIssueLabels(ctx context.Context, issueID int64) ([]sqlc.Label, error) {
	labels, err := r.queries.ListIssueLabels(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("list issue labels: %w", err)
	}
	return labels, nil
}
