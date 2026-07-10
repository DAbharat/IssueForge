package service

import "errors"

var (
	ErrInvalidPassword               = errors.New("password must be atleast 8 characters")
	ErrInvalidFullName               = errors.New("name must be more than 3 characters")
	ErrInvalidEmail                  = errors.New("email is invalid")
	ErrInvalidCredentials            = errors.New("invalid credentials")
	ErrInvalidProjectName            = errors.New("project name must be between 3 and 100 characters")
	ErrInvalidDescription            = errors.New("project description must be between 10 and 300 characters")
	ErrProjectNameTaken              = errors.New("a project with this name already exists for your account")
	ErrWorkspaceNameTaken            = errors.New("this workspace name has already been taken")
	ErrInvalidWorkspaceName          = errors.New("workspace name must be more than 3 and less than 30 characters")
	ErrWorkspaceNotFound             = errors.New("workspace does not exist")
	ErrProjectMemberAlreadyExists    = errors.New("member already exists in this project")
	ErrProjectNotFound               = errors.New("project not found")
	ErrUserNotFound                  = errors.New("user not found")
	ErrProjectMemberValidationFailed = errors.New("cannot add user in this project")
	ErrWorkspaceMemberAlreadyExists  = errors.New("member already exists in this workspace")
	ErrWorkspaceMemberNotFound       = errors.New("member does not exist in this workspace")
	ErrInvalidRole                   = errors.New("invalid role")
	ErrInvalidWorkspaceID            = errors.New("invalid workspace id")
	ErrInvalidUserID                 = errors.New("invalid user id")
)
