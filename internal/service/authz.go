package service

import "context"

type AuthzService interface {
	RequireWorkspaceAdmin(ctx context.Context, workspaceID, userID int64) error
	RequireWorkspaceMember(ctx context.Context, workspaceID, userID int64) error
	RequireProjectLead(ctx context.Context, projectID, userID int64) error
	RequireProjectMember(ctx context.Context, projectID, userID int64) error
	RequireWorkspaceAdminIncludingDeleted(ctx context.Context, workspaceID, userID int64) error
}
