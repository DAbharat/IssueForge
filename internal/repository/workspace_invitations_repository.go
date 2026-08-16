package repository

import (
	"IssueForge/internal/db/sqlc"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkspaceInvitationsRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewWorkspaceInvitationRepository(db *pgxpool.Pool, queries *sqlc.Queries) *WorkspaceInvitationsRepository {
	return &WorkspaceInvitationsRepository{
		db:      db,
		queries: queries,
	}
}

func (r *WorkspaceInvitationsRepository) CreateWorkspaceInvitation(ctx context.Context, workspaceID, invitedUserID, invitedBy int64) (sqlc.WorkspaceInvitation, error) {
	params := sqlc.CreateWorkspaceInvitationParams{
		WorkspaceID:   workspaceID,
		InvitedUserID: invitedUserID,
		InvitedBy:     invitedBy,
	}

	invitation, err := r.queries.CreateWorkspaceInvitation(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "workspace_invitations_pending_unique" {
				return sqlc.WorkspaceInvitation{}, ErrPendingInvitationExists
			}
		}
		return sqlc.WorkspaceInvitation{}, fmt.Errorf("create workspace invitation: %w", err)
	}
	return invitation, nil
}

func (r *WorkspaceInvitationsRepository) GetWorkspaceInvitation(ctx context.Context, id int64) (sqlc.WorkspaceInvitation, error) {
	invitation, err := r.queries.GetWorkspaceInvitation(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.WorkspaceInvitation{}, ErrInvitationNotFound
		}
		return sqlc.WorkspaceInvitation{}, fmt.Errorf("get workspace invitation: %w", err)
	}
	return invitation, nil
}

func (r *WorkspaceInvitationsRepository) ListPendingWorkspaceInvitations(ctx context.Context, invitedUserID int64) ([]sqlc.ListPendingWorkspaceInvitationsRow, error) {
	invitations, err := r.queries.ListPendingWorkspaceInvitations(ctx, invitedUserID)
	if err != nil {
		return nil, fmt.Errorf("list pending invitations: %w", err)
	}
	return invitations, nil
}

func (r *WorkspaceInvitationsRepository) ListPendingWorkspaceInvitationsForWorkspace(ctx context.Context, workspaceID int64) ([]sqlc.ListPendingWorkspaceInvitationsForWorkspaceRow, error) {
	invitations, err := r.queries.ListPendingWorkspaceInvitationsForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list pending workspace for workspace: %w", err)
	}
	return invitations, nil
}

func (r *WorkspaceInvitationsRepository) AcceptInvitation(ctx context.Context, id, invitedUserID int64) (sqlc.AcceptInvitationRow, error) {
	params := sqlc.AcceptInvitationParams{
		ID:            id,
		InvitedUserID: invitedUserID,
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.AcceptInvitationRow{}, fmt.Errorf("accept invitation transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := r.queries.WithTx(tx)

	invitation, err := qtx.AcceptInvitation(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.AcceptInvitationRow{}, ErrInvitationNotFound
		}
		return sqlc.AcceptInvitationRow{}, fmt.Errorf("accept invitation: %w", err)
	}

	_, err = qtx.AddWorkspaceMember(ctx, sqlc.AddWorkspaceMemberParams{
		WorkspaceID: invitation.WorkspaceID,
		UserID:      invitation.InvitedUserID,
		Role:        "MEMBER",
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			return sqlc.AcceptInvitationRow{}, ErrWorkspaceMemberAlreadyExists
		}
		return sqlc.AcceptInvitationRow{}, fmt.Errorf("add workspace member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.AcceptInvitationRow{}, fmt.Errorf("commit invitation acceptance: %w", err)
	}
	return invitation, nil
}

func (r *WorkspaceInvitationsRepository) DeclineInvitation(ctx context.Context, id, invitedUserID int64) (sqlc.DeclineInvitationRow, error) {
	params := sqlc.DeclineInvitationParams{
		ID:            id,
		InvitedUserID: invitedUserID,
	}

	invitation, err := r.queries.DeclineInvitation(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.DeclineInvitationRow{}, ErrInvitationNotFound
		}
		return sqlc.DeclineInvitationRow{}, fmt.Errorf("decline invitation: %w", err)
	}
	return invitation, nil
}

func (r *WorkspaceInvitationsRepository) CancelInvitation(ctx context.Context, id, invitedBy int64) (sqlc.CancelInvitationRow, error) {
	params := sqlc.CancelInvitationParams{
		ID:        id,
		InvitedBy: invitedBy,
	}

	invitation, err := r.queries.CancelInvitation(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.CancelInvitationRow{}, ErrInvitationNotFound
		}
		return sqlc.CancelInvitationRow{}, fmt.Errorf("cancel invitation: %w", err)
	}
	return invitation, nil
}
