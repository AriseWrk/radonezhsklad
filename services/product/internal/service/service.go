package service

import (
"context"

"github.com/google/uuid"

"github.com/radonezhsklad/product/internal/models"
"github.com/radonezhsklad/product/internal/repository"
apperr "github.com/radonezhsklad/shared/errors"
)

type Service struct{ repo *repository.Repo }

func New(repo *repository.Repo) *Service { return &Service{repo: repo} }

func (s *Service) CreateCategory(ctx context.Context, name string, parentID *uuid.UUID) (*models.Category, error) {
if name == "" { return nil, apperr.BadRequest("name is required") }
return s.repo.CreateCategory(ctx, name, parentID)
}
func (s *Service) ListCategories(ctx context.Context) ([]models.Category, error) { return s.repo.ListCategories(ctx) }
func (s *Service) GetCategory(ctx context.Context, id uuid.UUID) (*models.Category, error) {
c, err := s.repo.GetCategory(ctx, id)
if err != nil { return nil, apperr.Internal("get category", err) }
if c == nil { return nil, apperr.NotFound("category not found") }
return c, nil
}
func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, name string, parentID *uuid.UUID) (*models.Category, error) {
if name == "" { return nil, apperr.BadRequest("name is required") }
c, err := s.repo.UpdateCategory(ctx, id, name, parentID)
if err != nil { return nil, apperr.Internal("update category", err) }
if c == nil { return nil, apperr.NotFound("category not found") }
return c, nil
}
func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.DeleteCategory(ctx, id)
if err != nil { return apperr.Internal("delete category", err) }
if !ok { return apperr.NotFound("category not found") }
return nil
}

func (s *Service) ListUnits(ctx context.Context) ([]models.Unit, error) { return s.repo.ListUnits(ctx) }

func (s *Service) CreateProduct(ctx context.Context, in repository.ProductInput) (*models.Product, error) {
if in.Name == "" { return nil, apperr.BadRequest("name is required") }
if in.Currency == "" { in.Currency = "RUB" }
p, err := s.repo.CreateProduct(ctx, in)
if err != nil {
if repository.IsUniqueViolation(err) { return nil, apperr.Conflict("product with this SKU already exists") }
return nil, apperr.Internal("create product", err)
}
return p, nil
}

func (s *Service) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
p, err := s.repo.GetProduct(ctx, id)
if err != nil { return nil, apperr.Internal("get product", err) }
if p == nil { return nil, apperr.NotFound("product not found") }
return p, nil
}

func (s *Service) ListProducts(ctx context.Context, includeArchived bool, categoryID *uuid.UUID) ([]models.Product, error) {
return s.repo.ListProducts(ctx, includeArchived, categoryID)
}

func (s *Service) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]models.Product, error) {
return s.repo.ListByIDs(ctx, ids)
}

func (s *Service) UpdateProduct(ctx context.Context, id uuid.UUID, in repository.ProductInput) (*models.Product, error) {
if in.Name == "" { return nil, apperr.BadRequest("name is required") }
if in.Currency == "" { in.Currency = "RUB" }
p, err := s.repo.UpdateProduct(ctx, id, in)
if err != nil {
if repository.IsUniqueViolation(err) { return nil, apperr.Conflict("product with this SKU already exists") }
return nil, apperr.Internal("update product", err)
}
if p == nil { return nil, apperr.NotFound("product not found") }
return p, nil
}

func (s *Service) ArchiveProduct(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.ArchiveProduct(ctx, id)
if err != nil { return apperr.Internal("archive product", err) }
if !ok { return apperr.NotFound("product not found") }
return nil
}