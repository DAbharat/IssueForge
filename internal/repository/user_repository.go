package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{
		queries: queries,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, email, username, fullName, passwordHash string) (sqlc.CreateOnboardingUserRow, error) {
	params := sqlc.CreateOnboardingUserParams{
		Email:        email,
		Fullname:     fullName,
		Username:     username,
		PasswordHash: passwordHash,
	}

	user, err := r.queries.CreateOnboardingUser(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_email_key":
					return sqlc.CreateOnboardingUserRow{}, ErrDuplicateEmail
				case "users_username_key":
					return sqlc.CreateOnboardingUserRow{}, ErrDuplicateUsername
				}
			}
		}
		return sqlc.CreateOnboardingUserRow{}, fmt.Errorf("create onboarding user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserForLogin(ctx context.Context, email string) (sqlc.GetUserForLoginRow, error) {
	user, err := r.queries.GetUserForLogin(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetUserForLoginRow{}, ErrUserNotFound
		}
		return sqlc.GetUserForLoginRow{}, fmt.Errorf("get user for login: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (sqlc.GetUserByIDRow, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetUserByIDRow{}, ErrUserNotFound
		}
		return sqlc.GetUserByIDRow{}, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (sqlc.GetUserByUsernameRow, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetUserByUsernameRow{}, ErrUserNotFound
		}
		return sqlc.GetUserByUsernameRow{}, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepository) SearchUserByUsername(ctx context.Context, search *string) ([]sqlc.SearchUserByUsernameRow, error) {
	var newSearch pgtype.Text
	if search != nil {
		newSearch.String = *search
		newSearch.Valid = true
	}

	user, err := r.queries.SearchUserByUsername(ctx, newSearch)
	if err != nil {
		return nil, fmt.Errorf("search user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int64) (int64, error) {
	id, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, fmt.Errorf("delete user: %w", err)
	}
	return id, nil
}
