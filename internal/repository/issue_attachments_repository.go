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

type IssueAttachmentsRepository struct {
	queries *sqlc.Queries
}

func NewIssueAttachmentsRepository(queries *sqlc.Queries) *IssueAttachmentsRepository {
	return &IssueAttachmentsRepository{
		queries: queries,
	}
}

func (r *IssueAttachmentsRepository) CreateAttachment(ctx context.Context, issueID int64, commentID *int64, uploadedBy int64, originalName, storageKey, resourceType, mimeType string, fileSize int64) (sqlc.IssueAttachment, error) {
	var commenter pgtype.Int8
	if commentID != nil {
		commenter.Int64 = *commentID
		commenter.Valid = true
	}

	params := sqlc.CreateAttachmentParams{
		IssueID:      issueID,
		CommentID:    commenter,
		UploadedBy:   uploadedBy,
		OriginalName: originalName,
		StorageKey:   storageKey,
		ResourceType: resourceType,
		MimeType:     mimeType,
		FileSize:     fileSize,
	}

	attachment, err := r.queries.CreateAttachment(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "issue_attachments_issue_id_fkey":
					return sqlc.IssueAttachment{}, ErrIssueNotFound
				case "issue_attachments_comment_id_fkey":
					return sqlc.IssueAttachment{}, ErrCommentNotFound
				case "issue_attachments_uploaded_by_fkey":
					return sqlc.IssueAttachment{}, ErrUserNotFound
				}
			}
		}
		return sqlc.IssueAttachment{}, fmt.Errorf("create attachment: %w", err)
	}
	return attachment, nil
}

func (r *IssueAttachmentsRepository) GetAttachmentByID(ctx context.Context, id int64) (sqlc.GetAttachmentByIDRow, error) {
	attachment, err := r.queries.GetAttachmentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetAttachmentByIDRow{}, ErrAttachmentNotFound
		}
		return sqlc.GetAttachmentByIDRow{}, fmt.Errorf("get attachment by id: %w", err)
	}
	return attachment, nil
}

func (r *IssueAttachmentsRepository) ListCommentAttachments(ctx context.Context, commentID int64) ([]sqlc.ListCommentAttachmentsRow, error) {
	commenter := pgtype.Int8{
		Int64: commentID,
		Valid: true,
	}

	attachments, err := r.queries.ListCommentAttachments(ctx, commenter)
	if err != nil {
		return nil, fmt.Errorf("list comment attachments: %w", err)
	}
	return attachments, nil
}

func (r *IssueAttachmentsRepository) ListIssueAttachments(ctx context.Context, issueID int64) ([]sqlc.ListIssueAttachmentsRow, error) {
	attachments, err := r.queries.ListIssueAttachments(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("list issue attachments: %w", err)
	}
	return attachments, nil
}

func (r *IssueAttachmentsRepository) SoftDeleteAttachment(ctx context.Context, id int64) (sqlc.SoftDeleteAttachmentRow, error) {
	attachment, err := r.queries.SoftDeleteAttachment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.SoftDeleteAttachmentRow{}, ErrAttachmentNotFound
		}
		return sqlc.SoftDeleteAttachmentRow{}, fmt.Errorf("soft delete attachment: %w", err)
	}
	return attachment, nil
}
