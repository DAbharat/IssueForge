package service

import (
	"IssueForge/internal/auth"
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/redis/refreshtoken"
	"IssueForge/internal/repository"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const passwordHashCost = bcrypt.DefaultCost
const dummyHash = "$2a$10$Az9g3YmX8.L7Z8gH4q2uOu9yVj6X7W4E8k3d1c9b8a7f6e5d4c3b2"
const oidcPlaceholderHash = "!oidc-no-password!"

type UserRepo interface {
	CreateUser(ctx context.Context, email, displayName, fullname, passwordHash string) (sqlc.CreateOnboardingUserRow, error)
	GetUserForLogin(ctx context.Context, email string) (sqlc.GetUserForLoginRow, error)
	GetUserByID(ctx context.Context, id int64) (sqlc.GetUserByIDRow, error)
	GetUserByUsername(ctx context.Context, username string) (sqlc.GetUserByUsernameRow, error)
	SearchUserByUsername(ctx context.Context, search *string) ([]sqlc.SearchUserByUsernameRow, error)
	DeleteUser(ctx context.Context, id int64) (int64, error)
	GetUserByGoogleSub(ctx context.Context, sub *string) (sqlc.GetUserByGoogleSubRow, error)
	CreateOIDCUser(ctx context.Context, email, fullname, username, passwordHash string, googleSub *string) (sqlc.CreateOIDCUserRow, error)
	BindGoogleSub(ctx context.Context, id int64, sub *string) (int64, error)
}

type WorkspaceMemberRepo interface {
	ListUserWorkspaces(ctx context.Context, userID int64, search string) ([]sqlc.ListUserWorkspacesRow, error)
}

type UserService struct {
	userRepo            UserRepo
	workspaceMemberRepo WorkspaceMemberRepo
	jwtSecret           string
	refreshToken        *refreshtoken.Store
	googleOIDC          *auth.GoogleOIDC
}

func NewUserService(userRepo UserRepo, workspaceMemberRepo WorkspaceMemberRepo, jwtSecret string, refreshToken *refreshtoken.Store, googleOIDC *auth.GoogleOIDC) *UserService {
	return &UserService{
		userRepo:            userRepo,
		workspaceMemberRepo: workspaceMemberRepo,
		jwtSecret:           jwtSecret,
		refreshToken:        refreshToken,
		googleOIDC:          googleOIDC,
	}
}

func isValidUsername(username string) bool {
	pattern := `^[a-z0-9_.]+$`
	matched, _ := regexp.MatchString(pattern, username)
	return matched
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
	if !isValidUsername(req.Username) {
		return ErrInvalidUsername
	}
	if !validatePasswordComplexity(req.Password) {
		return ErrInvalidPassword
	}
	if utf8.RuneCountInString(req.Fullname) < 3 {
		return ErrInvalidFullName
	}
	if utf8.RuneCountInString(req.Username) < 3 || utf8.RuneCountInString(req.Username) > 30 {
		return ErrInvalidUsername
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

	user, err := s.userRepo.CreateUser(ctx, req.Email, req.Username, req.Fullname, string(hashedPassword))
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			return dto.CreateUserResponse{}, ErrDuplicateEmail
		}
		if errors.Is(err, repository.ErrDuplicateUsername) {
			return dto.CreateUserResponse{}, ErrDuplicateUsername
		}
		return dto.CreateUserResponse{}, fmt.Errorf("create user: %w", err)
	}

	return dto.CreateUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Fullname: user.Fullname,
		Email:    user.Email,
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
			ID:         user.ID,
			Username:   user.Username,
			Fullname:   user.Fullname,
			Email:      user.Email,
			Workspaces: s.mapWorkspaces(workspaces),
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
		ID:         user.ID,
		Username:   user.Username,
		Fullname:   user.Fullname,
		Email:      user.Email,
		Workspaces: s.mapWorkspaces(workspaces),
	}, nil
}

func (s *UserService) mapWorkspaces(workspaces []sqlc.ListUserWorkspacesRow) []dto.WorkspaceSummary {
	workspaceSummaries := make([]dto.WorkspaceSummary, 0, len(workspaces))
	for _, ws := range workspaces {
		workspaceSummaries = append(workspaceSummaries, dto.WorkspaceSummary{
			ID:        ws.ID,
			Name:      ws.Name,
			Role:      string(ws.Role),
			CreatedAt: ws.CreatedAt.Time,
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

func (s *UserService) GetUserByID(ctx context.Context, id int64) (dto.UserResponse, error) {
	if id < 0 {
		return dto.UserResponse{}, ErrInvalidUserID
	}

	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.UserResponse{}, ErrUserNotFound
		}
		return dto.UserResponse{}, fmt.Errorf("get user by id: %w", err)
	}
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Fullname:  user.Fullname,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, username string) (dto.UserResponse, error) {
	username = strings.TrimSpace(username)

	if username == "" {
		return dto.UserResponse{}, ErrInvalidUsername
	}

	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.UserResponse{}, ErrUserNotFound
		}
		return dto.UserResponse{}, fmt.Errorf("get user by username: %w", err)
	}
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Fullname:  user.Fullname,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

func (s *UserService) SearchUserByUsername(ctx context.Context, search *string) ([]dto.UserResponse, error) {
	newSearch := ""

	if search != nil {
		newSearch = *search
	}

	if newSearch == "" {
		search = nil
	} else {
		if utf8.RuneCountInString(newSearch) > 50 {
			return nil, ErrInvalidSearchQuery
		}
		search = &newSearch
	}

	dbUser, err := s.userRepo.SearchUserByUsername(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("search user by username: %w", err)
	}

	user := make([]dto.UserResponse, 0, len(dbUser))

	for _, u := range dbUser {
		user = append(user, dto.UserResponse{
			ID:       u.ID,
			Username: u.Username,
			Fullname: u.Fullname,
		})
	}
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int64) (int64, error) {
	if id < 0 {
		return 0, ErrInvalidUserID
	}

	userID, err := s.userRepo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return 0, ErrUserNotFound
		}
		return 0, fmt.Errorf("delete user: %w", err)
	}
	return userID, nil
}

func (s *UserService) GoogleAuthURL(state string) (string, error) {
	if s.googleOIDC == nil || s.googleOIDC.Config == nil {
		return "", errors.New("google oauth is not configured")
	}
	if len(state) < 32 {
		return "", errors.New("invalid oauth state")
	}

	return s.googleOIDC.Config.AuthCodeURL(state), nil
}

func (s *UserService) GoogleLogin(ctx context.Context, code string) (dto.LoginUserResponse, error) {
	if s.googleOIDC == nil || s.googleOIDC.Config == nil || s.googleOIDC.Verifier == nil {
		return dto.LoginUserResponse{}, errors.New("google oidc is not configured")
	}
	if strings.TrimSpace(code) == "" {
		return dto.LoginUserResponse{}, ErrInvalidCredentials
	}

	oauthToken, err := s.googleOIDC.Config.Exchange(ctx, code)
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("exchange google authorization code: %w", err)
	}

	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return dto.LoginUserResponse{}, errors.New("google response did not contain an id token")
	}

	idToken, err := s.googleOIDC.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("verify google id token: %w", err)
	}

	var claims dto.GoogleUserInfo
	if err := idToken.Claims(&claims); err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("decode google id token claims: %w", err)
	}

	claims.Sub = strings.TrimSpace(claims.Sub)
	claims.Email = strings.TrimSpace(claims.Email)

	if claims.Sub == "" || claims.Email == "" || !claims.EmailVerified {
		return dto.LoginUserResponse{}, ErrInvalidCredentials
	}

	type googleLoginUser struct {
		ID       int64
		Username string
		Fullname string
		Email    string
	}

	dbUser, err := s.userRepo.GetUserByGoogleSub(ctx, &claims.Sub)

	var user googleLoginUser

	if err == nil {
		user = googleLoginUser{
			ID:       dbUser.ID,
			Username: dbUser.Username,
			Fullname: dbUser.Fullname,
			Email:    dbUser.Email,
		}
	} else {
		if !errors.Is(err, repository.ErrUserNotFound) {
			return dto.LoginUserResponse{}, fmt.Errorf("find google user: %w", err)
		}

		existingUser, lookupErr := s.userRepo.GetUserForLogin(ctx, claims.Email)
		switch {
		case lookupErr == nil:
			if _, bindErr := s.userRepo.BindGoogleSub(ctx, existingUser.ID, &claims.Sub); bindErr != nil {
				if errors.Is(bindErr, repository.ErrDuplicateGoogleSub) {
					return dto.LoginUserResponse{}, ErrDuplicateGoogleSub
				}
				return dto.LoginUserResponse{}, fmt.Errorf("bind google sub: %w", bindErr)
			}

			user = googleLoginUser{
				ID:       existingUser.ID,
				Username: existingUser.Username,
				Fullname: existingUser.Fullname,
				Email:    existingUser.Email,
			}

		case errors.Is(lookupErr, repository.ErrUserNotFound):
			fullname := strings.TrimSpace(claims.Name)
			if fullname == "" {
				fullname = strings.TrimSpace(claims.GivenName + " " + claims.FamilyName)
			}

			created := false
			for attempt := 0; attempt < 5; attempt++ {
				username, genErr := s.generateGoogleUsername(ctx, claims.Email)
				if genErr != nil {
					return dto.LoginUserResponse{}, fmt.Errorf("generate google username: %w", genErr)
				}
				if utf8.RuneCountInString(fullname) < 3 {
					fullname = username
				}

				createdUser, createErr := s.userRepo.CreateOIDCUser(ctx, claims.Email, fullname, username, oidcPlaceholderHash, &claims.Sub)
				if createErr == nil {
					user = googleLoginUser{
						ID:       createdUser.ID,
						Username: createdUser.Username,
						Fullname: createdUser.Fullname,
						Email:    createdUser.Email,
					}
					created = true
					break
				}
				if errors.Is(createErr, repository.ErrDuplicateUsername) {
					continue
				}
				if errors.Is(createErr, repository.ErrDuplicateEmail) {
					return dto.LoginUserResponse{}, ErrDuplicateEmail
				}
				if errors.Is(createErr, repository.ErrDuplicateGoogleSub) {
					return dto.LoginUserResponse{}, ErrDuplicateGoogleSub
				}
				return dto.LoginUserResponse{}, fmt.Errorf("create google user: %w", createErr)
			}
			if !created {
				return dto.LoginUserResponse{}, ErrDuplicateUsername
			}

		default:
			return dto.LoginUserResponse{}, fmt.Errorf("check existing email: %w", lookupErr)
		}
	}

	workspaces, err := s.workspaceMemberRepo.ListUserWorkspaces(ctx, user.ID, "")
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("get login workspaces: %w", err)
	}

	accessToken, err := auth.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := auth.GenerateRefToken()
	if err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("generate refresh token: %w", err)
	}

	if err := s.refreshToken.CreateRefreshToken(ctx, refreshToken, user.ID); err != nil {
		return dto.LoginUserResponse{}, fmt.Errorf("store refresh token: %w", err)
	}

	return dto.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.MeResponse{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Fullname:   user.Fullname,
			Workspaces: s.mapWorkspaces(workspaces),
		},
	}, nil
}

func (s *UserService) generateGoogleUsername(ctx context.Context, email string) (string, error) {
	localPart := strings.SplitN(email, "@", 2)[0]
	localPart = strings.ToLower(localPart)

	var builder strings.Builder
	for _, r := range localPart {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '.' {
			builder.WriteRune(r)
		}
	}

	base := strings.Trim(builder.String(), "._")
	if len(base) < 3 {
		base = "googleuser"
	}
	if len(base) > 20 {
		base = base[:20]
	}

	for attempt := 0; attempt < 5; attempt++ {
		randomBytes := make([]byte, 4)
		if _, err := rand.Read(randomBytes); err != nil {
			return "", err
		}
		suffix := hex.EncodeToString(randomBytes)
		username := base + "_" + strings.ToLower(suffix[:5])

		_, err := s.userRepo.GetUserByUsername(ctx, username)
		if errors.Is(err, repository.ErrUserNotFound) {
			return username, nil
		}
		if err != nil {
			return "", err
		}
	}

	return "", errors.New("could not generate a unique username")
}

func (s *UserService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}

	err := s.refreshToken.DeleteRefreshToken(ctx, token)
	if err != nil && !errors.Is(err, refreshtoken.ErrTokenNotFound) {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}
