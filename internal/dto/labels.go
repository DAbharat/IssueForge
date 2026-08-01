package dto

import "time"

type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type UpdateLabelRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

type AttachLabelsRequest struct {
	LabelIDs []int64 `json:"label_ids"`
}

type LabelResponse struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedBy int64     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
