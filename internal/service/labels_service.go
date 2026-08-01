package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type LabelsRepo interface {
	CreateLabel(ctx context.Context, projectID int64, name, color string, createdBy int64) (sqlc.Label, error)
	GetLabelByID(ctx context.Context, labelID int64) (sqlc.Label, error)
	ListProjectLabels(ctx context.Context, projectID int64) ([]sqlc.Label, error)
	UpdateLabel(ctx context.Context, name, color *string, labelID int64) (sqlc.Label, error)
	DeleteLabel(ctx context.Context, labelID int64) (sqlc.DeleteLabelRow, error)
	GetLabelProjectID(ctx context.Context, labelID int64) (int64, error)
	AttachLabelsToIssue(ctx context.Context, issueID int64, labelsID []int64) error
	RemoveLabelFromIssue(ctx context.Context, issueID, labelID int64) (int64, error)
	ListIssueLabels(ctx context.Context, issueID int64) ([]sqlc.Label, error)
	CountProjectLabels(ctx context.Context, projectID int64, labelsID []int64) (int64, error)
}

type LabelsService struct {
	labelsRepo LabelsRepo
	issueRepo  IssueRepo
	authz      AuthzService
}

func NewLabelsService(labelsRepo LabelsRepo, issueRepo IssueRepo, authz AuthzService) *LabelsService {
	return &LabelsService{
		labelsRepo: labelsRepo,
		issueRepo:  issueRepo,
		authz:      authz,
	}
}

var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func (s *LabelsService) CreateLabel(ctx context.Context, requesterID, projectID int64, req dto.CreateLabelRequest) (dto.LabelResponse, error) {
	if projectID <= 0 {
		return dto.LabelResponse{}, ErrInvalidProjectID
	}

	req.Name = strings.TrimSpace(req.Name)

	if utf8.RuneCountInString(req.Name) < 3 || utf8.RuneCountInString(req.Name) > 20 {
		return dto.LabelResponse{}, ErrInvalidLabelName
	}

	if req.Color == "" {
		req.Color = "#525252"
	} else if !hexColorRegex.MatchString(req.Color) {
		return dto.LabelResponse{}, ErrInvalidColor
	}

	if err := s.authz.RequireProjectLead(ctx, projectID, requesterID); err != nil {
		return dto.LabelResponse{}, err
	}

	label, err := s.labelsRepo.CreateLabel(ctx, projectID, req.Name, req.Color, requesterID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.LabelResponse{}, ErrProjectNotFound
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.LabelResponse{}, ErrUserNotFound
		}
		if errors.Is(err, repository.ErrLabelAlreadyExists) {
			return dto.LabelResponse{}, ErrLabelAlreadyExists
		}
		return dto.LabelResponse{}, fmt.Errorf("create label: %w", err)
	}
	return dto.LabelResponse{
		ID:        label.ID,
		ProjectID: label.ProjectID,
		Name:      label.Name,
		Color:     label.Color,
		CreatedBy: label.CreatedBy,
		CreatedAt: label.CreatedAt.Time,
		UpdatedAt: label.UpdatedAt.Time,
	}, nil
}

func (s *LabelsService) GetLabelByID(ctx context.Context, requesterID, labelID int64) (dto.LabelResponse, error) {
	if labelID <= 0 {
		return dto.LabelResponse{}, ErrInvalidLabelID
	}

	label, err := s.labelsRepo.GetLabelByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return dto.LabelResponse{}, ErrLabelNotFound
		}
		return dto.LabelResponse{}, fmt.Errorf("get label by id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, label.ProjectID, requesterID); err != nil {
		return dto.LabelResponse{}, err
	}

	return dto.LabelResponse{
		ID:        labelID,
		ProjectID: label.ProjectID,
		Name:      label.Name,
		Color:     label.Color,
		CreatedBy: label.CreatedBy,
		CreatedAt: label.CreatedAt.Time,
		UpdatedAt: label.UpdatedAt.Time,
	}, nil
}

func (s *LabelsService) ListProjectLabels(ctx context.Context, requesterID, projectID int64) ([]dto.LabelResponse, error) {
	if projectID <= 0 {
		return nil, ErrInvalidProjectID
	}

	if err := s.authz.RequireProjectMember(ctx, projectID, requesterID); err != nil {
		return nil, err
	}

	dbLabels, err := s.labelsRepo.ListProjectLabels(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project labels: %w", err)
	}

	labels := make([]dto.LabelResponse, 0, len(dbLabels))

	for _, l := range dbLabels {
		labels = append(labels, dto.LabelResponse{
			ID:        l.ID,
			ProjectID: l.ProjectID,
			Name:      l.Name,
			Color:     l.Color,
			CreatedBy: l.CreatedBy,
			CreatedAt: l.CreatedAt.Time,
			UpdatedAt: l.UpdatedAt.Time,
		})
	}
	return labels, nil
}

func (s *LabelsService) UpdateLabel(ctx context.Context, requesterID int64, req dto.UpdateLabelRequest, labelID int64) (dto.LabelResponse, error) {
	if labelID <= 0 {
		return dto.LabelResponse{}, ErrInvalidLabelID
	}

	label, err := s.labelsRepo.GetLabelByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return dto.LabelResponse{}, ErrLabelNotFound
		}
		return dto.LabelResponse{}, fmt.Errorf("get label by id: %w", err)
	}

	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
		if utf8.RuneCountInString(*req.Name) < 3 || utf8.RuneCountInString(*req.Name) > 20 {
			return dto.LabelResponse{}, ErrInvalidLabelName
		}
	}

	if req.Color != nil {
		if *req.Color == "" {
			*req.Color = "#525252"
		} else if !hexColorRegex.MatchString(*req.Color) {
			return dto.LabelResponse{}, ErrInvalidColor
		}
	}

	if err := s.authz.RequireProjectLead(ctx, label.ProjectID, requesterID); err != nil {
		return dto.LabelResponse{}, err
	}

	labels, err := s.labelsRepo.UpdateLabel(ctx, req.Name, req.Color, label.ID)
	if err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return dto.LabelResponse{}, ErrLabelNotFound
		}
		if errors.Is(err, repository.ErrLabelAlreadyExists) {
			return dto.LabelResponse{}, ErrLabelAlreadyExists
		}
		return dto.LabelResponse{}, fmt.Errorf("update label: %w", err)
	}

	return dto.LabelResponse{
		ID:        labels.ID,
		ProjectID: labels.ProjectID,
		Name:      labels.Name,
		Color:     labels.Color,
		CreatedBy: labels.CreatedBy,
		CreatedAt: labels.CreatedAt.Time,
		UpdatedAt: labels.UpdatedAt.Time,
	}, nil
}

func (s *LabelsService) DeleteLabel(ctx context.Context, requesterID, labelID int64) (dto.LabelResponse, error) {
	if labelID <= 0 {
		return dto.LabelResponse{}, ErrInvalidLabelID
	}

	label, err := s.labelsRepo.GetLabelByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return dto.LabelResponse{}, ErrLabelNotFound
		}
		return dto.LabelResponse{}, fmt.Errorf("get label by id: %w", err)
	}

	if err := s.authz.RequireProjectLead(ctx, label.ProjectID, requesterID); err != nil {
		return dto.LabelResponse{}, err
	}

	delLabel, err := s.labelsRepo.DeleteLabel(ctx, label.ID)
	if err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return dto.LabelResponse{}, ErrLabelNotFound
		}
		return dto.LabelResponse{}, fmt.Errorf("delete label: %w", err)
	}
	return dto.LabelResponse{
		ID:        delLabel.ID,
		ProjectID: delLabel.ProjectID,
	}, nil
}

func (s *LabelsService) AttachLabelsToIssue(ctx context.Context, requesterID, issueID int64, req dto.AttachLabelsRequest) error {
	if issueID <= 0 {
		return ErrInvalidIssueID
	}

	if len(req.LabelIDs) <= 0 {
		return ErrInvalidLabelID
	}

	dbIssue, err := s.issueRepo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return ErrIssueNotFound
		}
		return fmt.Errorf("get issue by id: %w", err)
	}

	isLead := s.authz.RequireProjectLead(ctx, dbIssue.ProjectID, requesterID) == nil
	isCreator := dbIssue.CreatedBy == requesterID

	if !isCreator && !isLead {
		return ErrForbidden
	}

	count, err := s.labelsRepo.CountProjectLabels(ctx, dbIssue.ProjectID, req.LabelIDs)
	if err != nil {
		return fmt.Errorf("count project labels: %w", err)
	}
	if count != int64(len(req.LabelIDs)) {
		return ErrLabelDoesNotBelongToProject
	}

	err = s.labelsRepo.AttachLabelsToIssue(ctx, issueID, req.LabelIDs)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return ErrIssueNotFound
		}
		if errors.Is(err, repository.ErrLabelNotFound) {
			return ErrLabelNotFound
		}
		return fmt.Errorf("attach labels to issue: %w", err)
	}
	return nil
}

func (s *LabelsService) RemoveLabelFromIssue(ctx context.Context, requesterID, issueID, labelID int64) (int64, error) {
	if issueID <= 0 {
		return 0, ErrInvalidIssueID
	}
	if labelID <= 0 {
		return 0, ErrInvalidLabelID
	}

	label, err := s.labelsRepo.GetLabelByID(ctx, labelID)
	if err != nil {
		if errors.Is(err, repository.ErrLabelNotFound) {
			return 0, ErrLabelNotFound
		}
		return 0, fmt.Errorf("get labe by id: %w", err)
	}

	dbIssue, err := s.issueRepo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return 0, ErrIssueNotFound
		}
		return 0, fmt.Errorf("get issue by id: %w", err)
	}

	isLead := s.authz.RequireProjectLead(ctx, label.ProjectID, requesterID) == nil
	isCreator := dbIssue.CreatedBy == requesterID

	if !isCreator && !isLead {
		return 0, ErrForbidden
	}

	id, err := s.labelsRepo.RemoveLabelFromIssue(ctx, issueID, label.ID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return 0, ErrIssueNotFound
		}
		if errors.Is(err, repository.ErrLabelNotAttached) {
			return 0, ErrLabelNotAttached
		}
		return 0, fmt.Errorf("remove label from issue: %w", err)
	}
	return id, nil
}

func (s *LabelsService) ListIssueLabels(ctx context.Context, requesterID, issueID int64) ([]dto.LabelResponse, error) {
	if issueID <= 0 {
		return nil, ErrInvalidIssueID
	}

	dbIssue, err := s.issueRepo.GetIssueByID(ctx, issueID)
	if err != nil {
		if errors.Is(err, repository.ErrIssueNotFound) {
			return nil, ErrIssueNotFound
		}
		return nil, fmt.Errorf("get issue by id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, dbIssue.ProjectID, requesterID); err != nil {
		return nil, err
	}

	dbLabel, err := s.labelsRepo.ListIssueLabels(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("list issue labels: %w", err)
	}

	labels := make([]dto.LabelResponse, 0, len(dbLabel))

	for _, l := range dbLabel {
		labels = append(labels, dto.LabelResponse{
			ID:        l.ID,
			ProjectID: l.ProjectID,
			Name:      l.Name,
			Color:     l.Color,
			CreatedBy: l.CreatedBy,
			CreatedAt: l.CreatedAt.Time,
			UpdatedAt: l.UpdatedAt.Time,
		})
	}
	return labels, nil
}
