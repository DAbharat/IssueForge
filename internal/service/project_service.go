package service

import (
	"IssueForge/internal/db/sqlc"
	"IssueForge/internal/dto"
	"IssueForge/internal/redis"
	"IssueForge/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"
)

type ProjectRepo interface {
	CreateProject(ctx context.Context, workspaceID, leadID int64, name, description string) (sqlc.Project, error)
	GetProjectByID(ctx context.Context, projectID int64) (sqlc.GetProjectByIDRow, error)
	UpdateProjectDetails(ctx context.Context, name, description *string, id int64) (sqlc.Project, error)
	UpdateProjectLead(ctx context.Context, leadID *int64, projectID int64) (sqlc.Project, error)
	DeleteProject(ctx context.Context, id int64) (sqlc.DeleteProjectRow, error)
	ListProjectsByLead(ctx context.Context, workspaceID int64, leadID *int64) ([]sqlc.ListProjectsByLeadRow, error)
}

type ProjectService struct {
	projectRepo  ProjectRepo
	projectCache redis.ProjectCache
	authz        AuthzService
}

func NewProjectService(repo ProjectRepo, projectCache redis.ProjectCache, authz AuthzService) *ProjectService {
	return &ProjectService{
		projectRepo:  repo,
		projectCache: projectCache,
		authz:        authz,
	}
}

func (s *ProjectService) invalidateProjectCache(ctx context.Context, projectID int64) {
	if err := s.projectCache.DeleteProject(ctx, projectID); err != nil {
		log.Printf("failed to invalidate project cache: %v", err)
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, leadID int64, req dto.CreateProjectRequest) (dto.CreateProjectResponse, error) {
	if err := s.authz.RequireWorkspaceMember(ctx, req.WorkspaceID, leadID); err != nil {
		return dto.CreateProjectResponse{}, err
	}

	projectName := strings.TrimSpace(req.Name)
	projectDesc := strings.TrimSpace(req.Description)

	if utf8.RuneCountInString(projectName) < 3 || utf8.RuneCountInString(projectName) > 100 {
		return dto.CreateProjectResponse{}, ErrInvalidProjectName
	}

	if utf8.RuneCountInString(projectDesc) < 10 || utf8.RuneCountInString(projectDesc) > 300 {
		return dto.CreateProjectResponse{}, ErrInvalidDescription
	}

	project, err := s.projectRepo.CreateProject(ctx, req.WorkspaceID, leadID, projectName, projectDesc)
	if err != nil {
		if errors.Is(err, ErrProjectNameTaken) {
			return dto.CreateProjectResponse{}, err
		}
		return dto.CreateProjectResponse{}, fmt.Errorf("create project service failure: %w", err)
	}
	log.Printf("service: workspaceID=%d leadID=%d", req.WorkspaceID, leadID)

	return dto.CreateProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		LeadID:      project.LeadID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	}, nil
}

func (s *ProjectService) GetProjectByID(ctx context.Context, requesterID, projectID int64) (dto.ProjectResponse, error) {
	if projectID <= 0 {
		return dto.ProjectResponse{}, ErrInvalidProjectID
	}

	cachedProject, found, err := s.projectCache.GetProject(ctx, projectID)
	if err != nil {
		log.Printf("redis get failed: %v", err)
	}
	if found {
		log.Println("CACHE HIT")
		return cachedProject, nil
	}
	log.Println("CACHE MISS")

	project, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectResponse{}, ErrProjectNotFound
		}
		return dto.ProjectResponse{}, fmt.Errorf("get project by id: %w", err)
	}

	if err := s.authz.RequireProjectMember(ctx, project.ID, requesterID); err != nil {
		return dto.ProjectResponse{}, err
	}

	response := dto.ProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		LeadID:      project.LeadID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	}
	if err := s.projectCache.SetProject(ctx, response); err != nil {
		log.Printf("redis set fail: %v", err)
	}
	log.Println("CACHE SET")

	return response, nil
}

func (s *ProjectService) UpdateProjectDetails(ctx context.Context, requesterID int64, req dto.UpdateProjectDetailsRequest, projectID int64) (dto.ProjectResponse, error) {
	if projectID <= 0 {
		return dto.ProjectResponse{}, ErrInvalidProjectID
	}

	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
		if utf8.RuneCountInString(*req.Name) < 3 || utf8.RuneCountInString(*req.Name) > 100 {
			return dto.ProjectResponse{}, ErrInvalidProjectName
		}
	}

	if req.Description != nil {
		*req.Description = strings.TrimSpace(*req.Description)
		if utf8.RuneCountInString(*req.Description) < 10 || utf8.RuneCountInString(*req.Description) > 300 {
			return dto.ProjectResponse{}, ErrInvalidDescription
		}
	}

	dbProject, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectResponse{}, ErrProjectNotFound
		}
		return dto.ProjectResponse{}, fmt.Errorf("get project by id: %w", err)
	}

	if err := s.authz.RequireProjectLead(ctx, dbProject.ID, requesterID); err != nil {
		return dto.ProjectResponse{}, err
	}

	project, err := s.projectRepo.UpdateProjectDetails(ctx, req.Name, req.Description, dbProject.ID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectResponse{}, ErrProjectNotFound
		}
		if errors.Is(err, repository.ErrProjectAlreadyExists) {
			return dto.ProjectResponse{}, ErrProjectAlreadyExists
		}
		return dto.ProjectResponse{}, fmt.Errorf("update project details: %w", err)
	}

	s.invalidateProjectCache(ctx, project.ID)

	return dto.ProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		LeadID:      project.LeadID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	}, nil
}

func (s *ProjectService) UpdateProjectLead(ctx context.Context, requesterID int64, req dto.UpdateProjectLeadRequest, projectID int64) (dto.ProjectResponse, error) {
	if projectID <= 0 {
		return dto.ProjectResponse{}, ErrInvalidProjectID
	}

	if req.LeadID != nil {
		if *req.LeadID <= 0 {
			return dto.ProjectResponse{}, ErrInvalidLeadID
		}
	}

	dbProject, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectResponse{}, ErrProjectNotFound
		}
		return dto.ProjectResponse{}, fmt.Errorf("get project by id: %w", err)
	}

	if err := s.authz.RequireProjectLead(ctx, dbProject.ID, requesterID); err != nil {
		return dto.ProjectResponse{}, err
	}
	if err := s.authz.RequireProjectMember(ctx, dbProject.ID, *req.LeadID); err != nil {
		return dto.ProjectResponse{}, ErrUserNotProjectMember
	}

	if *req.LeadID == dbProject.LeadID {
		return dto.ProjectResponse{}, ErrLeadUnchanged
	}

	project, err := s.projectRepo.UpdateProjectLead(ctx, req.LeadID, dbProject.ID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return dto.ProjectResponse{}, ErrUserNotFound
		}
		return dto.ProjectResponse{}, fmt.Errorf("update project lead :%w", err)
	}

	s.invalidateProjectCache(ctx, project.ID)

	return dto.ProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		LeadID:      project.LeadID,
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	}, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, requesterID, projectID int64) (dto.ProjectResponse, error) {
	if projectID <= 0 {
		return dto.ProjectResponse{}, ErrInvalidProjectID
	}

	dbProject, err := s.projectRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectResponse{}, ErrProjectNotFound
		}
		return dto.ProjectResponse{}, fmt.Errorf("get project by id: %w", err)
	}

	isProjectLead := s.authz.RequireProjectLead(ctx, dbProject.ID, requesterID) == nil
	isWorkspaceAdmin := s.authz.RequireWorkspaceAdmin(ctx, dbProject.WorkspaceID, requesterID) == nil

	if !isProjectLead && !isWorkspaceAdmin {
		return dto.ProjectResponse{}, ErrForbidden
	}

	project, err := s.projectRepo.DeleteProject(ctx, dbProject.ID)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return dto.ProjectResponse{}, ErrProjectNotFound
		}
		return dto.ProjectResponse{}, fmt.Errorf("delete project: %w", err)
	}

	s.invalidateProjectCache(ctx, project.ID)

	return dto.ProjectResponse{
		ID:          project.ID,
		WorkspaceID: project.WorkspaceID,
		LeadID:      project.LeadID,
	}, nil
}

func (s *ProjectService) ListProjectsByLead(ctx context.Context, requesterID, workspaceID int64, leadID *int64) ([]dto.ProjectResponse, error) {
	if err := s.authz.RequireWorkspaceMember(ctx, workspaceID, requesterID); err != nil {
		return nil, err
	}

	if leadID != nil {
		if *leadID <= 0 {
			return nil, ErrInvalidLeadID
		}
	}

	dbProjects, err := s.projectRepo.ListProjectsByLead(ctx, workspaceID, leadID)
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
