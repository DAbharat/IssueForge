package dto

import "time"

type ProjectActivityResponse struct {
	ID                 int64     `json:"id"`
	ProjectID          int64     `json:"project_id"`
	ActorID            int64     `json:"actor_id"`
	ActorName          string    `json:"actor_name,omitempty"`
	ActorUsername      string    `json:"actor_username,omitempty"`
	TargetUserID       *int64    `json:"target_user_id,omitempty"`
	TargetUserName     *string   `json:"target_user_name,omitempty"`
	TargetUserUsername *string   `json:"target_user_username,omitempty"`
	ActivityType       string    `json:"activity_type"`
	FieldName          *string   `json:"field_name"`
	OldValue           *string   `json:"old_value"`
	NewValue           *string   `json:"new_value"`
	CreatedAt          time.Time `json:"created_at"`
}
