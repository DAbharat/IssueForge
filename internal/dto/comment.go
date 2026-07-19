package dto

import "time"

type CreateCommentRequest struct {
	IssueID         int64  `json:"issue_id"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
	Content         string `json:"content"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

type CommentResponse struct {
	ID              int64     `json:"id"`
	IssueID         int64     `json:"issue_id"`
	AuthorID        int64     `json:"author_id"`
	AuthorName      string    `json:"author_name"`
	ParentCommentID *int64    `json:"parent_comment_id"`
	Content         string    `jsson:"content"`
	IsEdited        bool      `json:"is_edited"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CommentTreeResponse struct {
	ID              int64                 `json:"id"`
	IssueID         int64                 `json:"issue_id"`
	AuthorID        int64                 `json:"author_id"`
	AuthorName      string                `json:"author_name"`
	ParentCommentID *int64                `json:"parent_comment_id"`
	Content         string                `json:"content"`
	IsEdited        bool                  `json:"id_edited"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	Replies         []CommentTreeResponse `json:"replies"`
}
