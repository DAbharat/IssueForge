package repository

import "errors"

var (
	ErrWorkspaceMemberAlreadyExists  = errors.New("user already exists in this workspace")
	ErrWorkspaceNotFound             = errors.New("workspace does not exist")
	ErrWorkspaceMemberNotFound       = errors.New("member does not exist in this workspace")
	ErrDuplicateEmail                = errors.New("email already exists")
	ErrUserNotFound                  = errors.New("user not found")
	ErrWorkspaceAlreadyExists        = errors.New("workspace with this name already exists")
	ErrProjectAlreadyExists          = errors.New("project with this name already exists")
	ErrProjectMemberAlreadyExists    = errors.New("member already exists in this project")
	ErrProjectNotFound               = errors.New("project does not exist")
	ErrProjectMemberValidationFailed = errors.New("cannot add the user in this project")
	ErrDuplicateProjectName          = errors.New("a project with this name already exists for this user")
	ErrIssueNotFound                 = errors.New("issue not found")
	ErrInvalidIssueID                = errors.New("invalid issue id")
	ErrIssueAlreadyExists            = errors.New("issue already exists")
	ErrCommentNotFound               = errors.New("comment not found")
	ErrInvalidActorID                = errors.New("invalid actor id")
	ErrInvalidActivityType           = errors.New("invalid activity type")
	ErrAttachmentNotFound            = errors.New("attachment not found")
)
