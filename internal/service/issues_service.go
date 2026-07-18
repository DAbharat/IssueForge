package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type IssueRepo interface {
	CreateIssue(ctx context.Context, creatorID, projectID int64, assignedTo *int64, title, description, status, priority string) (sqlc.Issue, error)
	GetIssueByID(ctx context.Context, issueID int64) (sqlc.GetIssueByIDRow, error)
	ListProjectIssues(ctx context.Context, projectID int64) ([]sqlc.ListProjectIssuesRow, error)
	UpdateIssueDetails(ctx context.Context, issueID int64, title, description string) (sqlc.Issue, error)
	UpdateIssueStatus(ctx context.Context, issueID int64, status string) (sqlc.Issue, error)
	UpdateIssueAssignee(ctx context.Context, issueID int64, assignedTo *int64) (sqlc.Issue, error)
	UpdateIssuePriority(ctx context.Context, issueID int64, priority string) (sqlc.Issue, error)
	ListAssignedIssues(ctx context.Context, assignedTo int64) ([]sqlc.ListAssignedIssuesRow, error)
	ListCreatedIssues(ctx context.Context, createdBy int64) ([]sqlc.ListCreatedIssuesRow, error)
	DeleteIssue(ctx context.Context, issueID int64) (int64, error)
}

type IssueService struct {
	repo  IssueRepo
	authz AuthzService
}

func NewIssueService(repo IssueRepo, authz AuthzService) *IssueService {
	return &IssueService{
		repo:  repo,
		authz: authz,
	}
}

func (s *IssueService) CreateIssue(ctx context.Context, creatorID int64, req dto.CreateIssueRequest) (dto.CreateIssueResponse, error) {
	if req.ProjectID <= 0 {
		return dto.CreateIssueResponse{}, ErrInvalidProjectID
	}

	if req.AssignedTo != nil && *req.AssignedTo <= 0 {
		return dto.CreateIssueResponse{}, ErrInvalidAssignee
	}

	issueTitle := strings.TrimSpace(req.Title)
	issueDescription := strings.TrimSpace(req.Description)

	if utf8.RuneCountInString(issueTitle) < 8 || utf8.RuneCountInString(issueTitle) > 50 {
		return dto.CreateIssueResponse{}, ErrInvalidTitle
	}
	if utf8.RuneCountInString(issueDescription) < 10 || utf8.RuneCountInString(issueDescription) > 300 {
		return dto.CreateIssueResponse{}, ErrInvalidDescription
	}

	switch req.Status {
	case "TODO", "IN_PROGRESS", "DONE":
	default:
		return dto.CreateIssueResponse{}, ErrInvalidStatus
	}

	switch req.Priority {
	case "LOW", "MEDIUM", "HIGH":
	default:
		return dto.CreateIssueResponse{}, ErrInvalidPriority
	}

	if err := s.authz.RequireProjectMember(ctx, req.ProjectID, creatorID); err != nil {
		return dto.CreateIssueResponse{}, err
	}

	if req.AssignedTo != nil {
		if err := s.authz.RequireProjectMember(ctx, req.ProjectID, *req.AssignedTo); err != nil {
			return dto.CreateIssueResponse{}, err
		}
	}

	issue, err := s.repo.CreateIssue(ctx, req.ProjectID, creatorID, req.AssignedTo, issueTitle, issueDescription, req.Status, req.Priority)
	if err != nil {
		return dto.CreateIssueResponse{}, fmt.Errorf("create issue: %w", err)
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}

	return dto.CreateIssueResponse{
		ID:          issue.ID,
		ProjectID:   issue.ProjectID,
		CreatedBy:   creatorID,
		AssignedTo:  assignedTo,
		Title:       issueTitle,
		Description: issueDescription,
		Status:      string(issue.Status),
		Priority:    string(issue.Priority),
		CreatedAt:   issue.CreatedAt.Time,
		UpdatedAt:   issue.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) GetIssueByID(ctx context.Context, requesterID, issueID int64) (dto.IssueResponse, error) {
	if issueID <= 0 {
		return dto.IssueResponse{}, ErrInvalidIssueID
	}

	issue, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, err
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}

	var assigneeName *string
	if issue.AssigneeName.Valid {
		assigneeName = &issue.AssigneeName.String
	}

	if err := s.authz.RequireProjectMember(ctx, issue.ProjectID, requesterID); err != nil {
		return dto.IssueResponse{}, fmt.Errorf("get issue by id: %w", err)
	}
	return dto.IssueResponse{
		ID:           issueID,
		ProjectID:    issue.ProjectID,
		CreatedBy:    issue.CreatedBy,
		CreatorName:  issue.CreatorName,
		AssignedTo:   assignedTo,
		AssigneeName: assigneeName,
		Title:        issue.Title,
		Description:  issue.Description,
		Status:       string(issue.Status),
		Priority:     string(issue.Priority),
		CreatedAt:    issue.CreatedAt.Time,
		UpdatedAt:    issue.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) ListProjectIssues(ctx context.Context, requesterID, projectID int64) ([]dto.IssueSummary, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbIssues, err := s.repo.ListProjectIssues(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("list project issues: %w", err)
	}

	issues := make([]dto.IssueSummary, 0, len(dbIssues))

	for _, i := range dbIssues {
		var assignedTo *int64
		if i.AssignedTo.Valid {
			assignedTo = &i.AssignedTo.Int64
		}

		var assigneeName *string
		if i.AssigneeName.Valid {
			assigneeName = &i.AssigneeName.String
		}

		issues = append(issues, dto.IssueSummary{
			ID:           i.ID,
			ProjectID:    i.ProjectID,
			CreatedBy:    i.CreatedBy,
			CreatorName:  i.CreatorName,
			AssignedTo:   assignedTo,
			AssigneeName: assigneeName,
			Title:        i.Title,
			Status:       string(i.Status),
			Priority:     string(i.Priority),
			CreatedAt:    i.CreatedAt.Time,
		})
	}
	return issues, nil
}

func (s *IssueService) UpdateIssueDetails(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssueDetailsRequest) (dto.IssueResponse, error) {
	if issueID <= 0 {
		return dto.IssueResponse{}, ErrInvalidIssueID
	}

	dbIssue, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		return dto.IssueResponse{}, err
	}

	description := dbIssue.Description

	issueTitle := strings.TrimSpace(req.Title)

	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)

		if utf8.RuneCountInString(description) < 10 || utf8.RuneCountInString(description) > 300 {
			return dto.IssueResponse{}, ErrInvalidDescription
		}
	}

	if utf8.RuneCountInString(issueTitle) < 8 || utf8.RuneCountInString(issueTitle) > 50 {
		return dto.IssueResponse{}, ErrInvalidTitle
	}

	isProjectLead := s.authz.RequireProjectLead(ctx, dbIssue.ProjectID, requesterID) == nil
	isCreator := requesterID == dbIssue.CreatedBy

	if !isProjectLead && !isCreator {
		return dto.IssueResponse{}, ErrForbidden
	}

	issues, err := s.repo.UpdateIssueDetails(ctx, issueID, issueTitle, description)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("update issue details: %w", err)
	}

	var assignedTo *int64
	if issues.AssignedTo.Valid {
		assignedTo = &issues.AssignedTo.Int64
	}
	return dto.IssueResponse{
		ID:          issues.ID,
		ProjectID:   issues.ProjectID,
		CreatedBy:   issues.CreatedBy,
		AssignedTo:  assignedTo,
		Title:       issues.Title,
		Description: issues.Description,
		Status:      string(issues.Status),
		Priority:    string(issues.Priority),
		CreatedAt:   issues.CreatedAt.Time,
		UpdatedAt:   issues.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) UpdateIssueStatus(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssueStatusRequest) (dto.IssueResponse, error) {
	if issueID <= 0 {
		return dto.IssueResponse{}, ErrInvalidIssueID
	}

	switch req.Status {
	case "TODO", "IN_PROGRESS", "DONE":
	default:
		return dto.IssueResponse{}, ErrInvalidStatus
	}

	dbIssue, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		return dto.IssueResponse{}, err
	}

	if err := s.authz.RequireProjectMember(ctx, dbIssue.ProjectID, requesterID); err != nil {
		return dto.IssueResponse{}, err
	}

	issue, err := s.repo.UpdateIssueStatus(ctx, issueID, req.Status)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("update issue status: %w", err)
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}
	return dto.IssueResponse{
		ID:          issue.ID,
		ProjectID:   issue.ProjectID,
		CreatedBy:   issue.CreatedBy,
		AssignedTo:  assignedTo,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      string(issue.Status),
		Priority:    string(issue.Priority),
		CreatedAt:   issue.CreatedAt.Time,
		UpdatedAt:   issue.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) UpdateIssueAssignee(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssueAssigneeRequest) (dto.IssueResponse, error) {
	if issueID <= 0 {
		return dto.IssueResponse{}, ErrInvalidIssueID
	}

	if req.AssignedTo != nil && *req.AssignedTo <= 0 {
		return dto.IssueResponse{}, ErrInvalidAssignee
	}

	dbIssue, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("get issue: %w", err)
	}

	if err := s.authz.RequireProjectLead(ctx, dbIssue.ProjectID, requesterID); err != nil {
		return dto.IssueResponse{}, err
	}

	if req.AssignedTo != nil {
		if err := s.authz.RequireProjectMember(ctx, dbIssue.ProjectID, *req.AssignedTo); err != nil {
			return dto.IssueResponse{}, err
		}
	}

	issue, err := s.repo.UpdateIssueAssignee(ctx, issueID, req.AssignedTo)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.IssueResponse{}, ErrUserNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("update issue assignee: %w", err)
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}
	return dto.IssueResponse{
		ID:          issue.ID,
		ProjectID:   issue.ProjectID,
		CreatedBy:   issue.CreatedBy,
		AssignedTo:  assignedTo,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      string(issue.Status),
		Priority:    string(issue.Priority),
		CreatedAt:   issue.CreatedAt.Time,
		UpdatedAt:   issue.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) UpdateIssuePriority(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssuePriority) (dto.IssueResponse, error) {
	if issueID <= 0 {
		return dto.IssueResponse{}, ErrInvalidIssueID
	}

	switch req.Priority {
	case "LOW", "MEDIUM", "HIGH":
	default:
		return dto.IssueResponse{}, ErrInvalidPriority
	}

	dbIssue, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("get issue by id: %w", err)
	}

	if err := s.authz.RequireProjectLead(ctx, dbIssue.ProjectID, requesterID); err != nil {
		return dto.IssueResponse{}, err
	}

	issue, err := s.repo.UpdateIssuePriority(ctx, issueID, req.Priority)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("update issue priority: %w", err)
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}

	return dto.IssueResponse{
		ID:          issue.ID,
		ProjectID:   issue.ProjectID,
		CreatedBy:   issue.CreatedBy,
		AssignedTo:  assignedTo,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      string(issue.Status),
		Priority:    string(issue.Priority),
		CreatedAt:   issue.CreatedAt.Time,
		UpdatedAt:   issue.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) ListAssignedIssues(ctx context.Context, requesterID, assignedTo int64) ([]dto.UserIssueSummary, error) {
	if assignedTo <= 0 {
		return nil, ErrInvalidAssignee
	}

	if requesterID != assignedTo {
		return nil, ErrForbidden
	}

	dbIssues, err := s.repo.ListAssignedIssues(ctx, assignedTo)
	if err != nil {
		return nil, fmt.Errorf("list assigned issues: %w", err)
	}

	issues := make([]dto.UserIssueSummary, 0, len(dbIssues))

	for _, i := range dbIssues {
		issues = append(issues, dto.UserIssueSummary{
			ID:          i.ID,
			ProjectID:   i.ProjectID,
			ProjectName: i.ProjectName,
			Title:       i.Title,
			Status:      string(i.Status),
			Priority:    string(i.Priority),
			CreatedAt:   i.CreatedAt.Time,
		})
	}
	return issues, nil
}

func (s *IssueService) ListCreatedIssues(ctx context.Context, requesterID, createdBy int64) ([]dto.UserIssueSummary, error) {
	if createdBy <= 0 {
		return nil, ErrInvalidUserID
	}

	if requesterID != createdBy {
		return nil, ErrForbidden
	}

	dbIssues, err := s.repo.ListCreatedIssues(ctx, createdBy)
	if err != nil {
		return nil, fmt.Errorf("list created issues: %w", err)
	}

	issues := make([]dto.UserIssueSummary, 0, len(dbIssues))
	for _, i := range dbIssues {
		issues = append(issues, dto.UserIssueSummary{
			ID:          i.ID,
			ProjectID:   i.ProjectID,
			ProjectName: i.ProjectName,
			Title:       i.Title,
			Status:      string(i.Status),
			Priority:    string(i.Priority),
			CreatedAt:   i.CreatedAt.Time,
		})
	}
	return issues, nil
}

func (s *IssueService) DeleteIssue(ctx context.Context, requesterID, issueID int64) (int64, error) {
	if issueID <= 0 {
		return 0, ErrInvalidIssueID
	}

	dbIssue, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		return 0, err
	}

	if err := s.authz.RequireProjectLead(ctx, dbIssue.ProjectID, requesterID); err != nil {
		return 0, err
	}

	id, err := s.repo.DeleteIssue(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return 0, ErrIssueNotFound
		}
		return 0, fmt.Errorf("delete issue: %w", err)
	}
	return id, nil
}
