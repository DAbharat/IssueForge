package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"net/mail"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword    = errors.New("password must be atleast 8 characters")
	ErrInvalidUsername    = errors.New("username must be more than 3 characters")
	ErrInvalidEmail       = errors.New("email is invalid")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

const passwordHashCost = bcrypt.DefaultCost
const dummyHash = "$2a$10$Az9g3YmX8.L7Z8gH4q2uOu9yVj6X7W4E8k3d1c9b8a7f6e5d4c3b2"

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: repo,
	}
}

func (s *UserService) validateCreateUser(req dto.CreateUserRequest) error {

	if len(req.Password) < 8 {
		return ErrInvalidPassword
	}
	if len(req.Username) < 3 {
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

	//DTO->sqlc params
	params := sqlc.CreateUserParams{
		Username: req.Username,
		Fullname: pgtype.Text{
			String: req.Fullname,
			Valid:  req.Fullname != "",
		},
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	user, err := s.userRepo.CreateUser(ctx, params)
	if err != nil {
		return dto.CreateUserResponse{}, fmt.Errorf("create user: %w", err)
	}

	return dto.CreateUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Fullname: user.Fullname.String,
		Email:    user.Email,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req dto.LoginUserRequest) (dto.LoginUserResponse, error) {
	if req.Identifier == "" || req.Password == "" {
		return dto.LoginUserResponse{}, ErrInvalidCredentials
	}

	var hashToCompare string
	userExists := true

	user, err := s.userRepo.GetUserForLogin(ctx, req.Identifier)
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

	return dto.LoginUserResponse{
		ID: user.ID,
	}, nil
}
