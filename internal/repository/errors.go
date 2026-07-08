package repository

import "errors"

var (
	ErrWorkspaceMemberAlreadyExists = errors.New("user already exists in this workspace")
	ErrWorkspaceNotFound            = errors.New("workspace does not exist")
	ErrWorkspaceMemberNotFound      = errors.New("member does not exist in this workspace")
	ErrDuplicateEmail               = errors.New("email already exists")
	ErrUserNotFound                 = errors.New("user not found")
)
