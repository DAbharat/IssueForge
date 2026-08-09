package service

import (
	"IssueForge/internal/auth"
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/redis/refreshtoken"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const passwordHashCost = bcrypt.DefaultCost
const dummyHash = "$2a$10$Az9g3YmX8.L7Z8gH4q2uOu9yVj6X7W4E8k3d1c9b8a7f6e5d4c3b2"

type UserRepo interface {
	CreateUser(ctx context.Context, email, displayName, fullname, passwordHash string) (sqlc.CreateOnboardingUserRow, error)
	GetUserForLogin(ctx context.Context, email string) (sqlc.GetUserForLoginRow, error)
	GetUserByID(ctx context.Context, id int64) (sqlc.GetUserByIDRow, error)
}

type WorkspaceMemberRepo interface {
	ListUserWorkspaces(ctx context.Context, userID int64, search string) ([]sqlc.ListUserWorkspacesRow, error)
}

type UserService struct {
	userRepo            UserRepo
	workspaceMemberRepo WorkspaceMemberRepo
	jwtSecret           string
	refreshToken        *refreshtoken.Store
}

func NewUserService(userRepo UserRepo, workspaceMemberRepo WorkspaceMemberRepo, jwtSecret string, refreshToken *refreshtoken.Store) *UserService {
	return &UserService{
		userRepo:            userRepo,
		workspaceMemberRepo: workspaceMemberRepo,
		jwtSecret:           jwtSecret,
		refreshToken:        refreshToken,
	}
}

func validatePasswordComplexity(password string) bool {
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func (s *UserService) validateCreateUser(req dto.CreateUserRequest) error {

	if len(req.Password) < 8 || len(req.Password) > 72 {
		return ErrInvalidPassword
	}
	if !validatePasswordComplexity(req.Password) {
		return ErrInvalidPassword
	}
	if utf8.RuneCountInString(req.Fullname) < 3 {
		return ErrInvalidFullName
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return ErrInvalidEmail
	}

	return nil
}

func (s *UserService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (dto.CreateUserResponse, error) {
	if err := s.validateCreateUser(req); err != nil {
		return dto.CreateUserResponse{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		passwordHashCost,
	)
	if err != nil {
		return dto.CreateUserResponse{}, fmt.Errorf("hash password failed: %w", err)
	}

	user, err := s.userRepo.CreateUser(ctx, req.Email, req.DisplayName, req.Fullname, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return dto.CreateUserResponse{}, ErrDuplicateEmail
		}
		return dto.CreateUserResponse{}, fmt.Errorf("create user: %w", err)
	}

	return dto.CreateUserResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Fullname:    user.Fullname,
		Email:       user.Email,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginUserRequest) (dto.LoginUserResponse, error) {
	if req.Email == "" || req.Password == "" {
		return dto.LoginUserResponse{}, ErrInvalidCredentials
	}

	var hashToCompare string
	userExists := true

	user, err := s.userRepo.GetUserForLogin(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			hashToCompare = dummyHash
			userExists = false
		} else {
			return dto.LoginUserResponse{}, fmt.Errorf("login failed: %w", err)
		}
	} else {
		hashToCompare = user.PasswordHash
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(hashToCompare),
		[]byte(req.Password),
	)
	if err != nil || !userExists {
		return dto.LoginUserResponse{}, ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("generate jwt: %w", err)
	}
	refToken, err := auth.GenerateRefToken()
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("generate refresh token: %w", err)
	}

	err = s.refreshToken.CreateRefreshToken(ctx, refToken, user.ID)
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("create refresh token: %w", err)
	}

	workspaces, err := s.workspaceMemberRepo.ListUserWorkspaces(ctx, user.ID, "")
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("get login workspaces: %w", err)
	}

	return dto.LoginUserResponse{
		AccessToken:  token,
		RefreshToken: refToken,
		User: dto.MeResponse{
			ID:          user.ID,
			DisplayName: user.DisplayName,
			Fullname:    user.Fullname,
			Email:       user.Email,
			Workspaces:  s.mapWorkspaces(workspaces),
		},
	}, nil
}

func (s *UserService) GetCurrentUser(ctx context.Context, userID int64) (dto.MeResponse, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return dto.MeResponse{}, err
	}

	workspaces, err := s.workspaceMemberRepo.ListUserWorkspaces(ctx, userID, "")
	if err != nil {
		return dto.MeResponse{}, fmt.Errorf("get user workspaces: %w", err)
	}
	log.Printf("%+v", workspaces)

	return dto.MeResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		Fullname:    user.Fullname,
		Email:       user.Email,
		Workspaces:  s.mapWorkspaces(workspaces),
	}, nil
}

func (s *UserService) mapWorkspaces(workspaces []sqlc.ListUserWorkspacesRow) []dto.WorkspaceSummary {
	workspaceSummaries := make([]dto.WorkspaceSummary, 0, len(workspaces))
	for _, ws := range workspaces {
		workspaceSummaries = append(workspaceSummaries, dto.WorkspaceSummary{
			ID:   ws.ID,
			Name: ws.Name,
			Role: string(ws.Role),
		})
	}
	return workspaceSummaries
}

func (s *UserService) RefreshAccessToken(ctx context.Context, req dto.RefreshTokenRequest) (string, error) {
	userID, err := s.refreshToken.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, refreshtoken.ErrTokenNotFound) {
			return "", ErrRefTokenNotFound
		}
		return "", fmt.Errorf("get refresh token: %w", err)
	}

	genNewToken, err := auth.GenerateToken(userID, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to genrate access token: %w", err)
	}

	return genNewToken, nil
}
