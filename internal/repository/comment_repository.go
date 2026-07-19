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

type CommentRepository struct {
	queries *sqlc.Queries
}

func NewCommentRepository(queries *sqlc.Queries) *CommentRepository {
	return &CommentRepository{
		queries: queries,
	}
}

func (r *CommentRepository) CreateComment(ctx context.Context, issueID, authorID int64, parentCommentID *int64, content string) (sqlc.Comment, error) {
	var parentComment pgtype.Int8

	if parentCommentID != nil {
		parentComment.Int64 = *parentCommentID
		parentComment.Valid = true
	}

	params := sqlc.CreateCommentParams{
		IssueID:         issueID,
		AuthorID:        authorID,
		ParentCommentID: parentComment,
		Content:         content,
	}

	comment, err := r.queries.CreateComment(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				switch pgErr.ConstraintName {
				case "comments_issue_id_fkey":
					return sqlc.Comment{}, ErrIssueNotFound
				case "comments_author_id_fkey":
					return sqlc.Comment{}, ErrUserNotFound
				case "comments_parent_comment_id_fkey":
					return sqlc.Comment{}, ErrCommentNotFound
				}
			}
		}
		return sqlc.Comment{}, fmt.Errorf("create comment: %w", err)
	}
	return comment, nil
}

func (r *CommentRepository) GetCommentByID(ctx context.Context, commentID int64) (sqlc.GetCommentByIDRow, error) {
	comment, err := r.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetCommentByIDRow{}, ErrCommentNotFound
		}
		return sqlc.GetCommentByIDRow{}, fmt.Errorf("get comment by id: %w", err)
	}
	return comment, nil
}

func (r *CommentRepository) ListIssueComments(ctx context.Context, issueID int64, limit, offset int32) ([]sqlc.ListIssueCommentsRow, error) {
	params := sqlc.ListIssueCommentsParams{
		IssueID: issueID,
		Limit:   limit,
		Offset:  offset,
	}

	comments, err := r.queries.ListIssueComments(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list issue comments: %w", err)
	}
	return comments, nil
}

func (r *CommentRepository) UpdateComment(ctx context.Context, id int64, content string) (sqlc.Comment, error) {
	params := sqlc.UpdateCommentParams{
		ID:      id,
		Content: content,
	}

	comment, err := r.queries.UpdateComment(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Comment{}, ErrCommentNotFound
		}
		return sqlc.Comment{}, fmt.Errorf("update comment: %w", err)
	}
	return comment, nil
}

func (r *CommentRepository) DeleteComment(ctx context.Context, id int64) (int64, error) {
	id, err := r.queries.DeleteComment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrCommentNotFound
		}
		return 0, fmt.Errorf("delete comment: %w", err)
	}
	return id, nil
}
