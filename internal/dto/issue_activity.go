package dto

import "time"

type ActivityResponse struct {
	ID           int64     `json:"id"`
	IssueID      int64     `json:"issue_id"`
	ActorID      int64     `json:"actor_id"`
	ActorName    string    `json:"actor_name"`
	ActivityType string    `json:"activity_type"`
	FieldName    *string   `json:"field_name"`
	OldValue     *string   `json:"old_value"`
	NewValue     *string   `json:"new_value"`
	CreatedAt    time.Time `json:"created_at"`
}
