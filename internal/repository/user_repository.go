package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
