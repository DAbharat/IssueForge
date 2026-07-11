package auth

import "errors"

type UserRole string

const (
	RoleAdmin  UserRole = "ADMIN"
	RoleMember UserRole = "MEMBER"
)

var (
	ErrForbidden          = errors.New("forbidden")
	ErrMembershipNotFound = errors.New("workspace membership not found")
)
