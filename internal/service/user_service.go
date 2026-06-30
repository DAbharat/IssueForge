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
	ErrInvalidPassword = errors.New("password must be atleast 8 characters")
	ErrInvalidUsername = errors.New("username must be more than 3 characters")
	ErrInvalidEmail    = errors.New("email is invalid")
)

const passwordHashCost = bcrypt.DefaultCost

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

func (s *UserService) CreateUser(
	ctx context.Context,
	req dto.CreateUserRequest,
) (dto.CreateUserResponse, error) {

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
		return dto.CreateUserResponse{}, fmt.Errorf("create user(service): %w", err)
	}

	return dto.CreateUserResponse{
		ID:       user.ID,
		Username: user.Username,
		Fullname: user.Fullname.String,
		Email:    user.Email,
	}, nil
}
