package dto

import "time"

type ArchiveCountsResponse struct {
	Workspaces  int64 `json:"workspaces"`
	Projects    int64 `json:"projects"`
	Issues      int64 `json:"issues"`
	Attachments int64 `json:"attachments"`
}

type ArchivedWorkspaceItem struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type ArchivedProjectItem struct {
	ID            int64      `json:"id"`
	WorkspaceID   int64      `json:"workspace_id"`
	WorkspaceName string     `json:"workspace_name"`
	LeadID        int64      `json:"lead_id"`
	LeadName      string     `json:"lead_name"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

type ArchivedIssueItem struct {
	ID            int64      `json:"id"`
	ProjectID     int64      `json:"project_id"`
	ProjectName   string     `json:"project_name"`
	WorkspaceID   int64      `json:"workspace_id"`
	WorkspaceName string     `json:"workspace_name"`
	CreatedBy     int64      `json:"created_by"`
	CreatorName   string     `json:"creator_name"`
	AssignedTo    *int64     `json:"assigned_to"`
	AssigneeName  *string    `json:"assignee_name"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DueDate       *time.Time `json:"due_date"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

type ArchivedAttachmentItem struct {
	ID           int64      `json:"id"`
	IssueID      int64      `json:"issue_id"`
	IssueTitle   string     `json:"issue_title"`
	ProjectID    int64      `json:"project_id"`
	ProjectName  string     `json:"project_name"`
	UploadedBy   int64      `json:"uploaded_by"`
	UploaderName string     `json:"uploader_name"`
	OriginalName string     `json:"original_name"`
	ResourceType string     `json:"resource_type"`
	MimeType     string     `json:"mime_type"`
	FileSize     int64      `json:"file_size"`
	PublicURL    string     `json:"public_url"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}
