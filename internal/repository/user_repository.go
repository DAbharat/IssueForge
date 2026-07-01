package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateUsername = errors.New("username already exists")
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{
		queries: queries,
	}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	params sqlc.CreateUserParams,
) (sqlc.User, error) {

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
