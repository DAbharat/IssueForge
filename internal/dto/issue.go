package dto

import "time"

type CreateIssueRequest struct {
	ProjectID   int64  `json:"project_id"`
	AssignedTo  *int64 `json:"assigned_to,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type CreateIssueResponse struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	CreatedBy   int64     `json:"created_by"`
	AssignedTo  *int64    `json:"assigned_to,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type IssueResponse struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"project_id"`
	CreatedBy    int64     `json:"created_by"`
	CreatorName  string    `json:"creator_name"`
	AssignedTo   *int64    `json:"assigned_to,omitempty"`
	AssigneeName *string   `json:"assignee_name,omitempty"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type IssueSummary struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"project_id"`
	CreatedBy    int64     `json:"created_by"`
	CreatorName  string    `json:"creator_name"`
	AssignedTo   *int64    `json:"assigned_to,omitempty"`
	AssigneeName *string   `json:"assignee_name"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
}

type UpdateIssueDetailsRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type UpdateIssueStatusRequest struct {
	Status string `json:"status"`
}

type UpdateIssueAssigneeRequest struct {
	AssignedTo *int64 `json:"assigned_to"`
}

type UserIssueSummary struct {
	ID          int64     `json:"id"`
	ProjectID   int64     `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}
