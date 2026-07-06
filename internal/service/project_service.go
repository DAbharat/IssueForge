package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidProjectName = errors.New("project name must be between 3 and 100 characters")
	ErrInvalidDescription = errors.New("project description must be between 10 and 300 characters")
	ErrProjectNameTaken   = errors.New("a project with this name already exists for your account")
)

type ProjectService struct {
	projectRepo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepo: repo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, ownerID int64, req dto.CreateProjectRequest) (dto.CreateProjectResponse, error) {
	projectName := strings.TrimSpace(req.Name)
	projectDesc := strings.TrimSpace(req.Description)

	if len(projectName) < 3 || len(projectName) > 100 {
		return dto.CreateProjectResponse{}, ErrInvalidProjectName
	}

	if len(projectDesc) < 10 || len(projectDesc) > 300 {
		return dto.CreateProjectResponse{}, ErrInvalidDescription
	}

	params := sqlc.CreateProjectParams{
		OwnerID:     ownerID,
		Name:        projectName,
		Description: projectDesc,
	}

	project, err := s.projectRepo.CreateProject(ctx, params)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateProjectName) {
			return dto.CreateProjectResponse{}, ErrProjectNameTaken
		}
		return dto.CreateProjectResponse{}, fmt.Errorf("create project service failure: %w", err)
	}

	return dto.CreateProjectResponse{
		ID:          project.ID,
		OwnerID:     project.OwnerID,
		Name:        project.Name,
		Description: project.Description,
	}, nil
}
