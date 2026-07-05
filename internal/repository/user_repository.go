package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateUsername = errors.New("username already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{
		queries: queries,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.User, error) {
	user, err := r.queries.CreateUser(ctx, params)
	if err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_email_key":
					return sqlc.User{}, ErrDuplicateEmail

				case "users_username_key":
					return sqlc.User{}, ErrDuplicateUsername
				}
			}
		}
		return sqlc.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserForLogin(ctx context.Context, identifier string) (sqlc.GetUserForLoginRow, error) {
	user, err := r.queries.GetUserForLogin(ctx, identifier)
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
