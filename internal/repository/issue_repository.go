package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type IssueRepository struct {
	queries *sqlc.Queries
}

func NewIssueRepository(queries *sqlc.Queries) *IssueRepository {
	return &IssueRepository{
		queries: queries,
	}
}

func (r *IssueRepository) CreateIssue(ctx context.Context, projectID, createdBy int64, assignedTo *int64, title, description, status, priority string) (sqlc.Issue, error) {
	var assignee pgtype.Int8

	if assignedTo != nil {
		assignee.Int64 = *assignedTo
		assignee.Valid = true
	}

	params := sqlc.CreateIssueParams{
		ProjectID:   projectID,
		CreatedBy:   createdBy,
		AssignedTo:  assignee,
		Title:       title,
		Description: description,
		Status:      sqlc.IssueStatus(status),
		Priority:    sqlc.IssuePriority(priority),
	}

	issue, err := r.queries.CreateIssue(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// fmt.Println("PG Code:", pgErr.Code)
			// fmt.Println("Constraint:", pgErr.ConstraintName)
			// fmt.Println("Message:", pgErr.Message)
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "issues_project_id_fkey":
					return sqlc.Issue{}, ErrProjectNotFound
				case "issues_created_by_fkey":
					return sqlc.Issue{}, ErrUserNotFound
				case "issues_assigned_to_fkey":
					return sqlc.Issue{}, ErrUserNotFound
				}
			}
		}
		return sqlc.Issue{}, fmt.Errorf("create issue: %w", err)
	}

	return issue, nil
}

func (r *IssueRepository) GetIssueByID(ctx context.Context, id int64) (sqlc.GetIssueByIDRow, error) {
	issue, err := r.queries.GetIssueByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetIssueByIDRow{}, ErrIssueNotFound
		}
		return sqlc.GetIssueByIDRow{}, fmt.Errorf("get issue by id: %w", err)
	}
	return issue, nil
}

func (r *IssueRepository) ListProjectIssues(ctx context.Context, projectID int64, status, priority *string, assignedTo *int64, search *string, limit, offset int32) ([]sqlc.ListProjectIssuesRow, error) {
	var newStatus sqlc.NullIssueStatus
	var newPriority sqlc.NullIssuePriority
	var assignee pgtype.Int8
	var newSearch pgtype.Text

	if status != nil {
		newStatus.IssueStatus = sqlc.IssueStatus(*status)
		newStatus.Valid = true
	}
	if priority != nil {
		newPriority.IssuePriority = sqlc.IssuePriority(*priority)
		newPriority.Valid = true
	}
	if assignedTo != nil {
		assignee.Int64 = *assignedTo
		assignee.Valid = true
	}
	if search != nil {
		newSearch.String = *search
		newSearch.Valid = true
	}

	params := sqlc.ListProjectIssuesParams{
		ProjectID:  projectID,
		Status:     newStatus,
		Priority:   newPriority,
		AssignedTo: assignee,
		Search:     newSearch,
		PageLimit:  limit,
		PageOffset: offset,
	}

	issues, err := r.queries.ListProjectIssues(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list project issues: %w", err)
	}
	return issues, nil
}

func (r *IssueRepository) UpdateIssueDetails(ctx context.Context, id int64, title, description string) (sqlc.Issue, error) {
	params := sqlc.UpdateIssueDetailsParams{
		ID:          id,
		Title:       title,
		Description: description,
	}

	issue, err := r.queries.UpdateIssueDetails(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Issue{}, ErrIssueNotFound
		}
		return sqlc.Issue{}, fmt.Errorf("update issue details: %w", err)
	}
	return issue, nil
}

func (r *IssueRepository) UpdateIssueStatus(ctx context.Context, id int64, status string) (sqlc.Issue, error) {
	params := sqlc.UpdateIssueStatusParams{
		ID:     id,
		Status: sqlc.IssueStatus(status),
	}

	issue, err := r.queries.UpdateIssueStatus(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Issue{}, ErrIssueNotFound
		}
		return sqlc.Issue{}, fmt.Errorf("update issue status: %w", err)
	}
	return issue, nil
}

func (r *IssueRepository) UpdateIssueAssignee(ctx context.Context, id int64, assignedTo *int64) (sqlc.Issue, error) {
	var assignee pgtype.Int8

	if assignedTo != nil {
		assignee.Int64 = *assignedTo
		assignee.Valid = true
	}

	params := sqlc.UpdateIssueAssigneeParams{
		ID:         id,
		AssignedTo: assignee,
	}

	issue, err := r.queries.UpdateIssueAssignee(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "issues_assigned_to_fkey":
					return sqlc.Issue{}, ErrUserNotFound
				}
			}
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Issue{}, ErrIssueNotFound
		}
		return sqlc.Issue{}, fmt.Errorf("update issue assignee: %w", err)
	}
	return issue, nil
}

func (r *IssueRepository) UpdateIssuePriority(ctx context.Context, id int64, priority string) (sqlc.Issue, error) {
	params := sqlc.UpdateIssuePriorityParams{
		ID:       id,
		Priority: sqlc.IssuePriority(priority),
	}

	issue, err := r.queries.UpdateIssuePriority(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Issue{}, ErrIssueNotFound
		}
		return sqlc.Issue{}, fmt.Errorf("update issue priority: %w", err)
	}
	return issue, nil
}

func (r *IssueRepository) UpdateIssueDueDate(ctx context.Context, issueID int64, dueDate *time.Time) (sqlc.Issue, error) {
	var dbDueDate pgtype.Timestamptz
	if dueDate != nil {
		dbDueDate.Time = *dueDate
		dbDueDate.Valid = true
	}

	params := sqlc.UpdateIssueDueDateParams{
		ID:      issueID,
		DueDate: dbDueDate,
	}

	issue, err := r.queries.UpdateIssueDueDate(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Issue{}, ErrIssueNotFound
		}
		return sqlc.Issue{}, fmt.Errorf("update due date: %w", err)
	}
	return issue, nil
}

func (r *IssueRepository) ListAssignedIssues(ctx context.Context, assignedTo int64) ([]sqlc.ListAssignedIssuesRow, error) {
	assignee := pgtype.Int8{
		Int64: assignedTo,
		Valid: true,
	}

	issues, err := r.queries.ListAssignedIssues(ctx, assignee)
	if err != nil {
		return nil, fmt.Errorf("list assigned issues: %w", err)
	}
	return issues, nil
}

func (r *IssueRepository) ListCreatedIssues(ctx context.Context, createdBy int64) ([]sqlc.ListCreatedIssuesRow, error) {
	issues, err := r.queries.ListCreatedIssues(ctx, createdBy)
	if err != nil {
		return nil, fmt.Errorf("list created issues: %w", err)
	}
	return issues, nil
}

func (r *IssueRepository) DeleteIssue(ctx context.Context, id int64) (int64, error) {
	issue, err := r.queries.DeleteIssue(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrIssueNotFound
		}
		return 0, fmt.Errorf("delete issue: %w", err)
	}
	return issue, nil
}

func (r *IssueRepository) GetIssueProjectID(ctx context.Context, id int64) (int64, error) {
	projectID, err := r.queries.GetIssueProjectID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrProjectNotFound
		}
		return 0, fmt.Errorf("get issue project id: %w", err)
	}
	return projectID, nil
}

func (r *IssueRepository) RestoreDeletedIssue(ctx context.Context, id int64) (int64, error) {
	issue, err := r.queries.RestoreIssue(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrIssueNotFound
		}
		return 0, fmt.Errorf("restore deleted issue: %w", err)
	}
	return issue, nil
}
