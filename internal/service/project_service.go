package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"context"
	"errors"
	"fmt"
	"strings"
)

type ProjectRepo interface {
	CreateProject(ctx context.Context, workspaceID, leadID int64, name, description string) (sqlc.CreateProjectRow, error)
	ListProjectsByLead(ctx context.Context, leadID int64) ([]sqlc.Project, error)
	ListProjectsByWorkspace(ctx context.Context, workspaceID int64) ([]sqlc.Project, error)
}

type ProjectService struct {
	projectRepo ProjectRepo
}

func NewProjectService(repo ProjectRepo) *ProjectService {
	return &ProjectService{
		projectRepo: repo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, workspaceID int64, leadID int64, req dto.CreateProjectRequest) (dto.CreateProjectResponse, error) {
	projectName := strings.TrimSpace(req.Name)
	projectDesc := strings.TrimSpace(req.Description)

	if len(projectName) < 3 || len(projectName) > 100 {
		return dto.CreateProjectResponse{}, ErrInvalidProjectName
	}

	if len(projectDesc) < 10 || len(projectDesc) > 300 {
		return dto.CreateProjectResponse{}, ErrInvalidDescription
	}

	project, err := s.projectRepo.CreateProject(ctx, workspaceID, leadID, projectName, projectDesc)
	if err != nil {
		if errors.Is(err, ErrProjectNameTaken) {
			return dto.CreateProjectResponse{}, err
		}
		return dto.CreateProjectResponse{}, fmt.Errorf("create project service failure: %w", err)
	}

	return dto.CreateProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		LeadID:      project.LeadID,
		Name:        project.Name,
		Description: project.Description,
	}, nil
}

func (s *ProjectService) ListProjectsByLead(ctx context.Context, leadID int64) ([]dto.ProjectResponse, error) {
	dbProjects, err := s.projectRepo.ListProjectsByLead(ctx, leadID)
	if err != nil {
		return nil, fmt.Errorf("list projects by lead service failure: %w", err)
	}

	projects := make([]dto.ProjectResponse, 0, len(dbProjects))

	for _, p := range dbProjects {
		projects = append(projects, dto.ProjectResponse{
			ID:          p.ID,
			WorkspaceID: p.WorkspaceID,
			LeadID:      p.LeadID,
			Name:        p.Name,
			Description: p.Description,
		})
	}
	return projects, nil
}

func (s *ProjectService) ListProjectsByWorkspace(ctx context.Context, workspaceID int64) ([]dto.ProjectResponse, error) {
	workspaceProjects, err := s.projectRepo.ListProjectsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list projects by workspaces: %w", err)
	}

	projects := make([]dto.ProjectResponse, 0, len(workspaceProjects))

	for _, p := range workspaceProjects {
		projects = append(projects, dto.ProjectResponse{
			ID:          p.ID,
			WorkspaceID: p.WorkspaceID,
			LeadID:      p.LeadID,
			Name:        p.Name,
			Description: p.Description,
		})
	}
	return projects, nil
}
