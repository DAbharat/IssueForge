package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/redis/queue"
	"IssueForge/internal/repository"
	"IssueForge/internal/storage"
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
)

type IssueAttachmentRepo interface {
	CreateAttachment(ctx context.Context, issueID int64, commentID *int64, uploadedBy int64, originalName, storageKey, resourceType, mimeType string, fileSize int64) (sqlc.IssueAttachment, error)
	GetAttachmentByID(ctx context.Context, id int64) (sqlc.GetAttachmentByIDRow, error)
	ListCommentAttachments(ctx context.Context, commentID int64) ([]sqlc.ListCommentAttachmentsRow, error)
	ListIssueAttachments(ctx context.Context, issueID int64) ([]sqlc.ListIssueAttachmentsRow, error)
	SoftDeleteAttachments(ctx context.Context, id int64) (sqlc.SoftDeleteAttachmentRow, error)
}

type IssueAttachmentService struct {
	repo            IssueAttachmentRepo
	issueLookupRepo IssueLookupRepo
	commentRepo     CommentRepo
	storage         storage.Storage
	deleteQueue     queue.AttachmentDeleteQueue
	authz           AuthzService
}

func NewIssueAttachmentsService(repo IssueAttachmentRepo, issueLookupRepo IssueLookupRepo, commentRepo CommentRepo, storage storage.Storage, deleteQueue queue.AttachmentDeleteQueue, authz AuthzService) *IssueAttachmentService {
	return &IssueAttachmentService{
		repo:            repo,
		issueLookupRepo: issueLookupRepo,
		commentRepo:     commentRepo,
		storage:         storage,
		deleteQueue:     deleteQueue,
		authz:           authz,
	}
}

func (s *IssueAttachmentService) CreateAttachment(ctx context.Context, requesterID, issueID int64, commentID *int64, file multipart.File, header *multipart.FileHeader) (dto.AttachmentResponse, error) {
	if issueID <= 0 {
		return dto.AttachmentResponse{}, ErrInvalidIssueID
	}
	if commentID != nil && *commentID <= 0 {
		return dto.AttachmentResponse{}, ErrInvalidCommentID
	}

	projectID, err := s.issueLookupRepo.GetIssueProjectID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.AttachmentResponse{}, ErrIssueNotFound
		}
		return dto.AttachmentResponse{}, fmt.Errorf("get issue by id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return dto.AttachmentResponse{}, err
	}

	if commentID != nil {
		comment, err := s.commentRepo.GetCommentByID(ctx, *commentID)
		if err != nil {
			if errors.Is(err, repository.ErrCommentNotFound) {
				return dto.AttachmentResponse{}, ErrCommentNotFound
			}
			return dto.AttachmentResponse{}, fmt.Errorf("get comment by id: %w", err)
		}
		if comment.IssueID != issueID {
			return dto.AttachmentResponse{}, ErrCommentDoesNotBelongToIssue
		}
	}

	extension := filepath.Ext(header.Filename)
	uniqueID := uuid.NewString()

	var storageKey string

	if commentID != nil {
		storageKey = fmt.Sprintf("issueforge/comments/attachments/%s%s", uniqueID, extension)
	} else {
		storageKey = fmt.Sprintf("issueforge/issues/attachments/%s%s", uniqueID, extension)
	}

	uploadResult, err := s.storage.Upload(ctx, file, header, storageKey)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEmptyFile):
			return dto.AttachmentResponse{}, ErrEmptyFile
		case errors.Is(err, storage.ErrFileTooLarge):
			return dto.AttachmentResponse{}, ErrFileTooLarge
		case errors.Is(err, storage.ErrUnsupportedType):
			return dto.AttachmentResponse{}, ErrUnsupportedType
		default:
			return dto.AttachmentResponse{}, fmt.Errorf("upload attachment: %w", err)
		}
	}

	attachment, err := s.repo.CreateAttachment(ctx, issueID, commentID, requesterID, header.Filename, uploadResult.StorageKey, uploadResult.ResourceType, uploadResult.MIMEType, uploadResult.Size)
	if err != nil {
		job := queue.AttachmentDeleteJob{
			IssueID:      issueID,
			UserID:       requesterID,
			FilePublicID: uploadResult.StorageKey,
			ResourceType: uploadResult.ResourceType,
		}
		if err := s.deleteQueue.AddDeleteJob(ctx, job); err != nil {
			log.Printf("attachment delete job fail: %v", err)
		}

		switch {
		case errors.Is(err, repository.ErrIssueNotFound):
			return dto.AttachmentResponse{}, ErrIssueNotFound
		case errors.Is(err, repository.ErrCommentNotFound):
			return dto.AttachmentResponse{}, ErrCommentNotFound
		case errors.Is(err, repository.ErrUserNotFound):
			return dto.AttachmentResponse{}, ErrUserNotFound
		default:
			return dto.AttachmentResponse{}, fmt.Errorf("create attachment: %w", err)
		}
	}

	var comment *int64
	if attachment.CommentID.Valid {
		comment = &attachment.CommentID.Int64
	}

	return dto.AttachmentResponse{
		ID:           attachment.ID,
		IssueID:      attachment.IssueID,
		CommentID:    comment,
		UploadedBy:   attachment.UploadedBy,
		OriginalName: attachment.OriginalName,
		ResourceType: attachment.ResourceType,
		MimeType:     attachment.MimeType,
		FileSize:     attachment.FileSize,
		CreatedAt:    attachment.CreatedAt.Time,
	}, nil
}

func (s *IssueAttachmentService) GetAttachmentByID(ctx context.Context, requesterID, id int64) (dto.AttachmentResponse, error) {
	attachment, err := s.repo.GetAttachmentByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAttachmentNotFound) {
			return dto.AttachmentResponse{}, ErrAttachmentNotFound
		}
		return dto.AttachmentResponse{}, fmt.Errorf("get attachmenet by id: %w", err)
	}

	projectID, err := s.issueLookupRepo.GetIssueProjectID(ctx, attachment.IssueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.AttachmentResponse{}, ErrIssueNotFound
		}
		return dto.AttachmentResponse{}, fmt.Errorf("get issue project id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return dto.AttachmentResponse{}, err
	}

	var comment *int64
	if attachment.CommentID.Valid {
		comment = &attachment.CommentID.Int64
	}

	return dto.AttachmentResponse{
		ID:           attachment.ID,
		IssueID:      attachment.IssueID,
		CommentID:    comment,
		UploadedBy:   attachment.UploadedBy,
		OriginalName: attachment.OriginalName,
		MimeType:     attachment.MimeType,
		ResourceType: attachment.ResourceType,
		FileSize:     attachment.FileSize,
		CreatedAt:    attachment.CreatedAt.Time,
	}, nil
}

func (s *IssueAttachmentService) ListIssueAttachments(ctx context.Context, requesterID, issueID int64) ([]dto.AttachmentResponse, error) {
	if issueID <= 0 {
		return nil, ErrInvalidIssueID
	}

	projectID, err := s.issueLookupRepo.GetIssueProjectID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("get issue project id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbAttachments, err := s.repo.ListIssueAttachments(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("list issue attachments: %w", err)
	}

	attachments := make([]dto.AttachmentResponse, 0, len(dbAttachments))

	for _, a := range dbAttachments {
		var comment *int64
		if a.CommentID.Valid {
			comment = &a.CommentID.Int64
		}

		attachments = append(attachments, dto.AttachmentResponse{
			ID:           a.ID,
			IssueID:      a.IssueID,
			CommentID:    comment,
			UploadedBy:   a.UploadedBy,
			OriginalName: a.OriginalName,
			MimeType:     a.MimeType,
			ResourceType: a.ResourceType,
			FileSize:     a.FileSize,
			CreatedAt:    a.CreatedAt.Time,
		})
	}
	return attachments, nil
}

func (s *IssueAttachmentService) ListCommentAttachments(ctx context.Context, requesterID, commentID int64) ([]dto.AttachmentResponse, error) {
	if commentID <= 0 {
		return nil, ErrInvalidCommentID
	}

	comment, err := s.commentRepo.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("get comment by id: %w", err)
	}

	projectID, err := s.issueLookupRepo.GetIssueProjectID(ctx, comment.IssueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("get issue project id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbAttachment, err := s.repo.ListCommentAttachments(ctx, comment.ID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return nil, ErrCommentNotFound
		}
		return nil, fmt.Errorf("list comment attachments: %w", err)
	}

	attachments := make([]dto.AttachmentResponse, 0, len(dbAttachment))

	for _, a := range dbAttachment {
		var comment *int64
		if a.CommentID.Valid {
			comment = &a.CommentID.Int64
		}

		attachments = append(attachments, dto.AttachmentResponse{
			ID:           a.ID,
			IssueID:      a.IssueID,
			CommentID:    comment,
			UploadedBy:   a.UploadedBy,
			OriginalName: a.OriginalName,
			MimeType:     a.MimeType,
			ResourceType: a.ResourceType,
			FileSize:     a.FileSize,
			CreatedAt:    a.CreatedAt.Time,
		})
	}
	return attachments, nil
}

func (s *IssueAttachmentService) SoftDeleteAttachments(ctx context.Context, requesterID, id int64) (int64, error) {
	dbAttachment, err := s.repo.GetAttachmentByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAttachmentNotFound) {
			return 0, ErrAttachmentNotFound
		}
		return 0, fmt.Errorf("get attachment by id: %w", err)
	}

	projectID, err := s.issueLookupRepo.GetIssueProjectID(ctx, dbAttachment.IssueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return 0, ErrIssueNotFound
		}
		return 0, fmt.Errorf("get issue project id: %w", err)
	}

	isUploader := dbAttachment.UploadedBy == requesterID

	if !isUploader {
		return 0, ErrForbidden
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return 0, err
	}

	attachment, err := s.repo.SoftDeleteAttachments(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrAttachmentNotFound) {
			return 0, ErrAttachmentNotFound
		}
		return 0, fmt.Errorf("soft delete attachment: %w", err)
	}

	job := queue.AttachmentDeleteJob{
		AttachmentID: attachment.ID,
		IssueID:      dbAttachment.IssueID,
		UserID:       requesterID,
		FilePublicID: attachment.StorageKey,
		ResourceType: attachment.ResourceType,
	}

	if err := s.deleteQueue.AddDeleteJob(ctx, job); err != nil {
		return 0, fmt.Errorf("enqueue attachment deletion: %w", err)
	}

	return attachment.ID, nil
}
