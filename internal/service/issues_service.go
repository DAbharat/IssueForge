package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type IssueRepo interface {
	CreateIssue(ctx context.Context, creatorID, projectID int64, assignedTo *int64, title, description, status, priority string) (sqlc.Issue, error)
	GetIssueByID(ctx context.Context, issueID int64) (sqlc.GetIssueByIDRow, error)
	ListProjectIssues(ctx context.Context, projectID int64, status, priority *string, assignedTo *int64, search *string, limit, offset int32) ([]sqlc.ListProjectIssuesRow, error)
	UpdateIssueDetails(ctx context.Context, issueID int64, title, description string) (sqlc.Issue, error)
	UpdateIssueStatus(ctx context.Context, issueID int64, status string) (sqlc.Issue, error)
	UpdateIssueAssignee(ctx context.Context, issueID int64, assignedTo *int64) (sqlc.Issue, error)
	UpdateIssuePriority(ctx context.Context, issueID int64, priority string) (sqlc.Issue, error)
	UpdateIssueDueDate(ctx context.Context, issueID int64, dueDate *time.Time) (sqlc.Issue, error)
	ListAssignedIssues(ctx context.Context, assignedTo int64) ([]sqlc.ListAssignedIssuesRow, error)
	ListCreatedIssues(ctx context.Context, createdBy int64) ([]sqlc.ListCreatedIssuesRow, error)
	DeleteIssue(ctx context.Context, issueID int64) (int64, error)
	RestoreDeletedIssue(ctx context.Context, issueID int64) (int64, error)
}

type ActivityService interface {
	CreateActivity(ctx context.Context, issueID, actorID int64, activityType string, fieldName, oldValue, newValue *string) (dto.ActivityResponse, error)
}

type IssueService struct {
	repo         IssueRepo
	activityRepo ActivityService
	authz        AuthzService
}

func NewIssueService(repo IssueRepo, activity ActivityService, authz AuthzService) *IssueService {
	return &IssueService{
		repo:         repo,
		activityRepo: activity,
		authz:        authz,
	}
}

func strPtrEqual(a, b *string) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
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

	_, err = s.activityRepo.CreateActivity(ctx, issue.ID, creatorID, "ISSUE_CREATED", nil, nil, nil)
	if err != nil {
		return dto.CreateIssueResponse{}, fmt.Errorf("create activity: %w", err)
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}
	var dueDate *time.Time
	if issue.DueDate.Valid {
		dueDate = &issue.DueDate.Time
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
		DueDate:     dueDate,
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

	var dueDate *time.Time
	if issue.DueDate.Valid {
		dueDate = &issue.DueDate.Time
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
		DueDate:      dueDate,
		CreatedAt:    issue.CreatedAt.Time,
		UpdatedAt:    issue.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) ListProjectIssues(ctx context.Context, requesterID, projectID int64, req dto.ListProjectIssuesRequest) ([]dto.IssueSummary, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}

	if req.AssignedTo != nil && *req.AssignedTo <= 0 {
		return nil, ErrInvalidAssignee
	}

	search := strings.TrimSpace(*req.Search)
	if search == "" {
		req.Search = nil
	} else {
		if utf8.RuneCountInString(search) > 50 {
			return nil, ErrInvalidSearchQuery
		}
		req.Search = &search
	}

	if req.Status != nil {
		switch *req.Status {
		case "TODO", "IN_PROGRESS", "DONE":
		default:
			return nil, ErrInvalidStatus
		}
	}

	if req.Priority != nil {
		switch *req.Priority {
		case "LOW", "MEDIUM", "HIGH":
		default:
			return nil, ErrInvalidPriority
		}
	}

	if req.Limit <= 0 || req.Limit > 100 {
		return nil, ErrInvalidLimit
	}
	if req.Offset < 0 {
		return nil, ErrInvalidOffset
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbIssues, err := s.repo.ListProjectIssues(ctx, projectID, req.Status, req.Priority, req.AssignedTo, req.Search, req.Limit, req.Offset)
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

		var dueDate *time.Time
		if i.DueDate.Valid {
			dueDate = &i.DueDate.Time
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
			DueDate:      dueDate,
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
		description = strings.TrimSpace(*req.Description)

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

	titleField := "title"
	oldTitle := dbIssue.Title
	newTitle := issues.Title

	descField := "description"
	oldDesc := dbIssue.Description
	newDesc := issues.Description

	if oldTitle != newTitle {
		_, err = s.activityRepo.CreateActivity(ctx, issues.ID, requesterID, "ISSUE_DETAILS_UPDATED", &titleField, &oldTitle, &newTitle)
		if err != nil {
			return dto.IssueResponse{}, fmt.Errorf("create activity: %w", err)
		}
	}

	if oldDesc != newDesc {
		_, err = s.activityRepo.CreateActivity(ctx, issues.ID, requesterID, "ISSUE_DETAILS_UPDATED", &descField, &oldDesc, &newDesc)
		if err != nil {
			return dto.IssueResponse{}, fmt.Errorf("create activity: %w", err)
		}
	}

	var assignedTo *int64
	if issues.AssignedTo.Valid {
		assignedTo = &issues.AssignedTo.Int64
	}
	var assigneeName *string
	if dbIssue.AssigneeName.Valid {
		assigneeName = &dbIssue.AssigneeName.String
	}
	var dueDate *time.Time
	if issues.DueDate.Valid {
		dueDate = &issues.DueDate.Time
	}
	return dto.IssueResponse{
		ID:           issues.ID,
		ProjectID:    issues.ProjectID,
		CreatedBy:    issues.CreatedBy,
		AssignedTo:   assignedTo,
		AssigneeName: assigneeName,
		Title:        issues.Title,
		Description:  issues.Description,
		Status:       string(issues.Status),
		Priority:     string(issues.Priority),
		DueDate:      dueDate,
		CreatedAt:    issues.CreatedAt.Time,
		UpdatedAt:    issues.UpdatedAt.Time,
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

	oldStatus := string(dbIssue.Status)
	issue, err := s.repo.UpdateIssueStatus(ctx, issueID, req.Status)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("update issue status: %w", err)
	}
	newStatus := string(issue.Status)

	field := "status"

	if oldStatus != newStatus {
		_, err = s.activityRepo.CreateActivity(ctx, issue.ID, requesterID, "ISSUE_STATUS_CHANGED", &field, &oldStatus, &newStatus)
		if err != nil {
			return dto.IssueResponse{}, fmt.Errorf("create activity: %w", err)
		}
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}
	var assigneeName *string
	if dbIssue.AssigneeName.Valid {
		assigneeName = &dbIssue.AssigneeName.String
	}
	var dueDate *time.Time
	if issue.DueDate.Valid {
		dueDate = &issue.DueDate.Time
	}
	return dto.IssueResponse{
		ID:           issue.ID,
		ProjectID:    issue.ProjectID,
		CreatedBy:    issue.CreatedBy,
		AssignedTo:   assignedTo,
		AssigneeName: assigneeName,
		Title:        issue.Title,
		Description:  issue.Description,
		Status:       string(issue.Status),
		Priority:     string(issue.Priority),
		DueDate:      dueDate,
		CreatedAt:    issue.CreatedAt.Time,
		UpdatedAt:    issue.UpdatedAt.Time,
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

	var oldAssignee *string
	if dbIssue.AssigneeName.Valid {
		oldAssignee = &dbIssue.AssigneeName.String
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

	updatedIssueWithNames, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("get issue by id(name): %w", err)
	}

	var newAssignee *string
	if updatedIssueWithNames.AssigneeName.Valid {
		newAssignee = &updatedIssueWithNames.AssigneeName.String
	}

	field := "assignee"

	if !strPtrEqual(oldAssignee, newAssignee) {
		_, err = s.activityRepo.CreateActivity(ctx, issue.ID, requesterID, "ISSUE_ASSIGNEE_CHANGED", &field, oldAssignee, newAssignee)
		if err != nil {
			return dto.IssueResponse{}, fmt.Errorf("create activity: %w", err)
		}
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}
	var assigneeName *string
	if updatedIssueWithNames.AssigneeName.Valid {
		assigneeName = &updatedIssueWithNames.AssigneeName.String
	}
	var dueDate *time.Time
	if issue.DueDate.Valid {
		dueDate = &issue.DueDate.Time
	}
	return dto.IssueResponse{
		ID:           issue.ID,
		ProjectID:    issue.ProjectID,
		CreatedBy:    issue.CreatedBy,
		AssignedTo:   assignedTo,
		AssigneeName: assigneeName,
		Title:        issue.Title,
		Description:  issue.Description,
		Status:       string(issue.Status),
		Priority:     string(issue.Priority),
		DueDate:      dueDate,
		CreatedAt:    issue.CreatedAt.Time,
		UpdatedAt:    issue.UpdatedAt.Time,
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

	oldPriority := string(dbIssue.Priority)
	issue, err := s.repo.UpdateIssuePriority(ctx, issueID, req.Priority)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("update issue priority: %w", err)
	}
	newPriority := string(issue.Priority)

	field := "priority"

	if oldPriority != newPriority {
		_, err = s.activityRepo.CreateActivity(ctx, issue.ID, requesterID, "ISSUE_PRIORITY_CHANGED", &field, &oldPriority, &newPriority)
		if err != nil {
			return dto.IssueResponse{}, fmt.Errorf("create activity: %w", err)
		}
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}
	var assigneeName *string
	if dbIssue.AssigneeName.Valid {
		assigneeName = &dbIssue.AssigneeName.String
	}
	var dueDate *time.Time
	if issue.DueDate.Valid {
		dueDate = &issue.DueDate.Time
	}
	return dto.IssueResponse{
		ID:           issue.ID,
		ProjectID:    issue.ProjectID,
		CreatedBy:    issue.CreatedBy,
		AssignedTo:   assignedTo,
		AssigneeName: assigneeName,
		Title:        issue.Title,
		Description:  issue.Description,
		Status:       string(issue.Status),
		Priority:     string(issue.Priority),
		DueDate:      dueDate,
		CreatedAt:    issue.CreatedAt.Time,
		UpdatedAt:    issue.UpdatedAt.Time,
	}, nil
}

func (s *IssueService) UpdateIssueDueDate(ctx context.Context, requesterID, issueID int64, req dto.UpdateIssueDueDate) (dto.IssueResponse, error) {
	if issueID <= 0 {
		return dto.IssueResponse{}, ErrInvalidIssueID
	}

	if req.DueDate != nil {
		if req.DueDate.Before(time.Now().UTC()) {
			return dto.IssueResponse{}, ErrInvalidDueDate
		}
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

	var oldDueDate *string
	if dbIssue.DueDate.Valid {
		s := dbIssue.DueDate.Time.Format(time.RFC3339)
		oldDueDate = &s
	}

	issue, err := s.repo.UpdateIssueDueDate(ctx, issueID, req.DueDate)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.IssueResponse{}, ErrIssueNotFound
		}
		return dto.IssueResponse{}, fmt.Errorf("update due date: %w", err)
	}

	var newDueDate *string
	if issue.DueDate.Valid {
		s := issue.DueDate.Time.Format(time.RFC3339)
		newDueDate = &s
	}

	field := "due date"

	if !strPtrEqual(oldDueDate, newDueDate) {
		_, err := s.activityRepo.CreateActivity(ctx, issue.ID, requesterID, "DUE_DATE_CHANGED", &field, oldDueDate, newDueDate)
		if err != nil {
			return dto.IssueResponse{}, fmt.Errorf("create activity: %w", err)
		}
	}

	var assignedTo *int64
	if issue.AssignedTo.Valid {
		assignedTo = &issue.AssignedTo.Int64
	}
	var assigneeName *string
	if dbIssue.AssigneeName.Valid {
		assigneeName = &dbIssue.AssigneeName.String
	}

	var dueDate *time.Time
	if issue.DueDate.Valid {
		dueDate = &issue.DueDate.Time
	}

	return dto.IssueResponse{
		ID:           issue.ID,
		ProjectID:    issue.ProjectID,
		CreatedBy:    issue.CreatedBy,
		AssignedTo:   assignedTo,
		AssigneeName: assigneeName,
		Title:        issue.Title,
		Description:  issue.Description,
		Status:       string(issue.Status),
		Priority:     string(issue.Priority),
		CreatedAt:    issue.CreatedAt.Time,
		DueDate:      dueDate,
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
		var dueDate *time.Time
		if i.DueDate.Valid {
			dueDate = &i.DueDate.Time
		}
		issues = append(issues, dto.UserIssueSummary{
			ID:          i.ID,
			ProjectID:   i.ProjectID,
			ProjectName: i.ProjectName,
			Title:       i.Title,
			Status:      string(i.Status),
			Priority:    string(i.Priority),
			CreatedAt:   i.CreatedAt.Time,
			DueDate:     dueDate,
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
		var dueDate *time.Time
		if i.DueDate.Valid {
			dueDate = &i.DueDate.Time
		}
		issues = append(issues, dto.UserIssueSummary{
			ID:          i.ID,
			ProjectID:   i.ProjectID,
			ProjectName: i.ProjectName,
			Title:       i.Title,
			Status:      string(i.Status),
			Priority:    string(i.Priority),
			CreatedAt:   i.CreatedAt.Time,
			DueDate:     dueDate,
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

	_, err = s.activityRepo.CreateActivity(ctx, dbIssue.ID, requesterID, "ISSUE_DELETED", nil, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("create activity: %w", err)
	}
	return id, nil
}

func (s *IssueService) RestoreDeletedIssue(ctx context.Context, requesterID, issueID int64) (int64, error) {
	if issueID <= 0 {
		return 0, ErrInvalidIssueID
	}

	dbIssue, err := s.repo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return 0, ErrIssueNotFound
		}
		return 0, fmt.Errorf("get issue by id(restore): %w", err)
	}

	isCreator := requesterID == dbIssue.CreatedBy
	isLead := s.authz.RequireProjectLead(ctx, dbIssue.ProjectID, requesterID) == nil

	if !isCreator && !isLead {
		return 0, ErrForbidden
	}

	issue, err := s.repo.RestoreDeletedIssue(ctx, dbIssue.ID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return 0, ErrIssueNotFound
		}
		return 0, fmt.Errorf("restore deleted issue: %w", err)
	}

	_, err = s.activityRepo.CreateActivity(ctx, dbIssue.ID, requesterID, "ISSUE_RESTORED", nil, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("create activity: %w", err)
	}
	return issue, nil
}
