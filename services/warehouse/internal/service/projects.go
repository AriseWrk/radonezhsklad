package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	apperr "github.com/radonezhsklad/shared/errors"
	"github.com/radonezhsklad/warehouse/internal/models"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

type ProjectService struct {
	repo *repository.ProjectRepo
}

func NewProjectService(repo *repository.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

type ProjectInput struct {
	Name     string
	Archived bool
}

func (s *ProjectService) Create(ctx context.Context, in ProjectInput) (*models.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperr.BadRequest("name is required")
	}
	p := &models.Project{Name: name}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperr.Internal("create project", err)
	}
	return p, nil
}

func (s *ProjectService) Get(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, apperr.Internal("get project", err)
	}
	if p == nil {
		return nil, apperr.NotFound("project not found")
	}
	return p, nil
}

func (s *ProjectService) List(ctx context.Context, includeArchived bool) ([]models.Project, error) {
	return s.repo.List(ctx, includeArchived)
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, in ProjectInput) (*models.Project, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperr.BadRequest("name is required")
	}
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, apperr.Internal("get project", err)
	}
	if p == nil {
		return nil, apperr.NotFound("project not found")
	}
	p.Name = name
	p.Archived = in.Archived
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperr.Internal("update project", err)
	}
	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	ok, err := s.repo.Delete(ctx, id)
	if err != nil {
		return apperr.Internal("delete project", err)
	}
	if !ok {
		return apperr.NotFound("project not found")
	}
	return nil
}
