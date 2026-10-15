package service

import (
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"fmt"
)

type ArchiveService struct {
	repo *repository.ArchiveRepository
}

func NewArchiveService(repo *repository.ArchiveRepository) *ArchiveService {
	return &ArchiveService{repo: repo}
}

func (s *ArchiveService) GetArchivedCounts(ctx context.Context, userID int64) (dto.ArchiveCountsResponse, error) {
	counts, err := s.repo.GetArchivedCounts(ctx, userID)
	if err != nil {
		return dto.ArchiveCountsResponse{}, fmt.Errorf("get archived counts service: %w", err)
	}

	return dto.ArchiveCountsResponse{
		Workspaces:  counts.WorkspacesCount,
		Projects:    counts.ProjectsCount,
		Issues:      counts.IssuesCount,
		Attachments: counts.AttachmentsCount,
	}, nil
}

func (s *ArchiveService) ListArchivedWorkspaces(ctx context.Context, userID int64, search *string) ([]dto.ArchivedWorkspaceItem, error) {
	rows, err := s.repo.ListArchivedWorkspaces(ctx, userID, search)
	if err != nil {
		return nil, fmt.Errorf("list archived workspaces service: %w", err)
	}

	result := make([]dto.ArchivedWorkspaceItem, 0, len(rows))
	for _, w := range rows {
		item := dto.ArchivedWorkspaceItem{
			ID:        w.ID,
			Name:      w.Name,
			CreatedAt: w.CreatedAt.Time,
			UpdatedAt: w.UpdatedAt.Time,
		}
		if w.DeletedAt.Valid {
			item.DeletedAt = &w.DeletedAt.Time
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *ArchiveService) ListArchivedProjects(ctx context.Context, userID int64, search *string) ([]dto.ArchivedProjectItem, error) {
	rows, err := s.repo.ListArchivedProjects(ctx, userID, search)
	if err != nil {
		return nil, fmt.Errorf("list archived projects service: %w", err)
	}

	result := make([]dto.ArchivedProjectItem, 0, len(rows))
	for _, p := range rows {
		item := dto.ArchivedProjectItem{
			ID:            p.ID,
			WorkspaceID:   p.WorkspaceID,
			WorkspaceName: p.WorkspaceName,
			LeadID:        p.LeadID,
			LeadName:      p.LeadName,
			Name:          p.Name,
			Description:   p.Description,
			CreatedAt:     p.CreatedAt.Time,
			UpdatedAt:     p.UpdatedAt.Time,
		}
		if p.DeletedAt.Valid {
			item.DeletedAt = &p.DeletedAt.Time
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *ArchiveService) ListArchivedIssues(ctx context.Context, userID int64, search *string) ([]dto.ArchivedIssueItem, error) {
	rows, err := s.repo.ListArchivedIssues(ctx, userID, search)
	if err != nil {
		return nil, fmt.Errorf("list archived issues service: %w", err)
	}

	result := make([]dto.ArchivedIssueItem, 0, len(rows))
	for _, i := range rows {
		item := dto.ArchivedIssueItem{
			ID:            i.ID,
			ProjectID:     i.ProjectID,
			ProjectName:   i.ProjectName,
			WorkspaceID:   i.WorkspaceID,
			WorkspaceName: i.WorkspaceName,
			CreatedBy:     i.CreatedBy,
			CreatorName:   i.CreatorName,
			Title:         i.Title,
			Description:   i.Description,
			Status:        string(i.Status),
			Priority:      string(i.Priority),
			CreatedAt:     i.CreatedAt.Time,
			UpdatedAt:     i.UpdatedAt.Time,
		}
		if i.AssignedTo.Valid {
			item.AssignedTo = &i.AssignedTo.Int64
		}
		if i.AssigneeName.Valid {
			item.AssigneeName = &i.AssigneeName.String
		}
		if i.DueDate.Valid {
			item.DueDate = &i.DueDate.Time
		}
		if i.DeletedAt.Valid {
			item.DeletedAt = &i.DeletedAt.Time
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *ArchiveService) ListArchivedAttachments(ctx context.Context, userID int64, search *string) ([]dto.ArchivedAttachmentItem, error) {
	rows, err := s.repo.ListArchivedAttachments(ctx, userID, search)
	if err != nil {
		return nil, fmt.Errorf("list archived attachments service: %w", err)
	}

	result := make([]dto.ArchivedAttachmentItem, 0, len(rows))
	for _, a := range rows {
		item := dto.ArchivedAttachmentItem{
			ID:           a.ID,
			IssueID:      a.IssueID,
			IssueTitle:   a.IssueTitle,
			ProjectID:    a.ProjectID,
			ProjectName:  a.ProjectName,
			UploadedBy:   a.UploadedBy,
			UploaderName: a.UploaderName,
			OriginalName: a.OriginalName,
			ResourceType: a.ResourceType,
			MimeType:     a.MimeType,
			FileSize:     a.FileSize,
			PublicURL:    a.PublicUrl,
			CreatedAt:    a.CreatedAt.Time,
		}
		if a.DeletedAt.Valid {
			item.DeletedAt = &a.DeletedAt.Time
		}
		result = append(result, item)
	}
	return result, nil
}
