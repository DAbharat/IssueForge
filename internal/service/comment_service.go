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

type CommentRepo interface {
	CreateComment(ctx context.Context, issueID, requesterID int64, parentCommentID *int64, content string) (sqlc.Comment, error)
	GetCommentByID(ctx context.Context, commentID int64) (sqlc.GetCommentByIDRow, error)
	ListIssueComments(ctx context.Context, issueID int64, limit, offset int32) ([]sqlc.ListIssueCommentsRow, error)
	UpdateComment(ctx context.Context, commentID int64, content string) (sqlc.Comment, error)
	DeleteComment(ctx context.Context, commentID int64) (int64, error)
}

type IssueLookupRepo interface {
	GetIssueProjectID(ctx context.Context, issueID int64) (int64, error)
}

type CommentService struct {
	repo      CommentRepo
	issueRepo IssueLookupRepo
	authz     AuthzService
}

func NewCommentService(repo CommentRepo, issueRepo IssueLookupRepo, authz AuthzService) *CommentService {
	return &CommentService{
		repo:      repo,
		issueRepo: issueRepo,
		authz:     authz,
	}
}

var (
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrInvalidOffset = errors.New("invalid offset")
)

func (s *CommentService) CreateComment(ctx context.Context, requesterID int64, req dto.CreateCommentRequest) (dto.CommentResponse, error) {
	if req.IssueID <= 0 {
		return dto.CommentResponse{}, ErrInvalidIssueID
	}

	if req.ParentCommentID != nil && *req.ParentCommentID <= 0 {
		return dto.CommentResponse{}, ErrCommentNotFound
	}

	commentContent := strings.TrimSpace(req.Content)
	if utf8.RuneCountInString(commentContent) < 3 || utf8.RuneCountInString(commentContent) > 300 {
		return dto.CommentResponse{}, ErrInvalidComment
	}

	projectID, err := s.issueRepo.GetIssueProjectID(ctx, req.IssueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.CommentResponse{}, ErrIssueNotFound
		}
		return dto.CommentResponse{}, fmt.Errorf("get issue project: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return dto.CommentResponse{}, err
	}

	comment, err := s.repo.CreateComment(ctx, req.IssueID, requesterID, req.ParentCommentID, commentContent)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrIssueNotFound):
			return dto.CommentResponse{}, ErrIssueNotFound
		case errors.Is(err, repository.ErrCommentNotFound):
			return dto.CommentResponse{}, ErrCommentNotFound
		case errors.Is(err, repository.ErrUserNotFound):
			return dto.CommentResponse{}, ErrUserNotFound
		}
		return dto.CommentResponse{}, fmt.Errorf("create comment: %w", err)
	}

	var parentComment *int64
	if comment.ParentCommentID.Valid {
		parentComment = &comment.ParentCommentID.Int64
	}

	return dto.CommentResponse{
		ID:              comment.ID,
		IssueID:         comment.IssueID,
		AuthorID:        comment.AuthorID,
		ParentCommentID: parentComment,
		Content:         commentContent,
		IsEdited:        comment.IsEdited,
		CreatedAt:       comment.CreatedAt.Time,
		UpdatedAt:       comment.UpdatedAt.Time,
	}, nil
}

func (s *CommentService) GetCommentByID(ctx context.Context, commentID, requesterID int64) (dto.CommentResponse, error) {
	if commentID <= 0 {
		return dto.CommentResponse{}, ErrInvalidCommentID
	}

	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return dto.CommentResponse{}, ErrCommentNotFound
		}
		return dto.CommentResponse{}, fmt.Errorf("get comment by id: %w", err)
	}

	projectID, err := s.issueRepo.GetIssueProjectID(ctx, comment.IssueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return dto.CommentResponse{}, ErrIssueNotFound
		}
		return dto.CommentResponse{}, fmt.Errorf("get issue project: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return dto.CommentResponse{}, err
	}

	var parentComment *int64
	if comment.ParentCommentID.Valid {
		parentComment = &comment.ParentCommentID.Int64
	}

	return dto.CommentResponse{
		ID:              comment.ID,
		IssueID:         comment.IssueID,
		AuthorID:        comment.AuthorID,
		AuthorName:      comment.AuthorName,
		ParentCommentID: parentComment,
		Content:         comment.Content,
		IsEdited:        comment.IsEdited,
		CreatedAt:       comment.CreatedAt.Time,
		UpdatedAt:       comment.UpdatedAt.Time,
	}, nil
}

func (s *CommentService) ListIssueComments(ctx context.Context, requesterID, issueID int64, limit, offset int32) ([]dto.CommentResponse, error) {
	if issueID <= 0 {
		return nil, ErrInvalidIssueID
	}

	if limit <= 0 {
		return nil, ErrInvalidLimit
	}

	if offset < 0 {
		return nil, ErrInvalidOffset
	}

	projectID, err := s.issueRepo.GetIssueProjectID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("get issue project: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbComments, err := s.repo.ListIssueComments(ctx, issueID, limit, offset)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("list issue comments: %w", err)
	}

	comments := make([]dto.CommentResponse, 0, len(dbComments))

	for _, c := range dbComments {
		var parentComment *int64
		if c.ParentCommentID.Valid {
			parentComment = &c.ParentCommentID.Int64
		}

		comments = append(comments, dto.CommentResponse{
			ID:              c.ID,
			IssueID:         c.IssueID,
			AuthorID:        c.AuthorID,
			AuthorName:      c.AuthorName,
			ParentCommentID: parentComment,
			Content:         c.Content,
			IsEdited:        c.IsEdited,
			CreatedAt:       c.CreatedAt.Time,
			UpdatedAt:       c.UpdatedAt.Time,
		})
	}
	return comments, nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID, requesterID int64, req dto.UpdateCommentRequest) (dto.CommentResponse, error) {
	if commentID <= 0 {
		return dto.CommentResponse{}, ErrInvalidCommentID
	}

	commentContent := strings.TrimSpace(req.Content)
	if utf8.RuneCountInString(commentContent) < 3 || utf8.RuneCountInString(commentContent) > 300 {
		return dto.CommentResponse{}, ErrInvalidComment
	}

	dbComment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return dto.CommentResponse{}, ErrCommentNotFound
		}
		return dto.CommentResponse{}, fmt.Errorf("get comment by id: %w", err)
	}

	isCommenter := requesterID == dbComment.AuthorID
	if !isCommenter {
		return dto.CommentResponse{}, ErrForbidden
	}

	comment, err := s.repo.UpdateComment(ctx, commentID, commentContent)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return dto.CommentResponse{}, ErrCommentNotFound
		}
		return dto.CommentResponse{}, fmt.Errorf("update comment: %w", err)
	}

	var parentComment *int64
	if comment.ParentCommentID.Valid {
		parentComment = &comment.ParentCommentID.Int64
	}

	return dto.CommentResponse{
		ID:              comment.ID,
		IssueID:         comment.IssueID,
		AuthorID:        comment.AuthorID,
		AuthorName:      dbComment.AuthorName,
		ParentCommentID: parentComment,
		Content:         commentContent,
		IsEdited:        comment.IsEdited,
		CreatedAt:       comment.CreatedAt.Time,
		UpdatedAt:       comment.UpdatedAt.Time,
	}, nil
}

func (s *CommentService) DeleteComment(ctx context.Context, requesterID, commentID int64) (int64, error) {
	if commentID <= 0 {
		return 0, ErrInvalidCommentID
	}

	comment, err := s.repo.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return 0, ErrCommentNotFound
		}
		return 0, fmt.Errorf("get comment by id: %w", err)
	}

	isCommenter := requesterID == comment.AuthorID
	if !isCommenter {
		return 0, ErrForbidden
	}

	id, err := s.repo.DeleteComment(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository.ErrCommentNotFound) {
			return 0, ErrCommentNotFound
		}
		return 0, fmt.Errorf("delete comment: %w", err)
	}
	return id, nil
}
