package service

import (
"context"

"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/models"
"github.com/radonezhsklad/warehouse/internal/repository"
apperr "github.com/radonezhsklad/shared/errors"
)

type SupplierService struct {
repo *repository.SupplierRepo
orgs *repository.OrganizationRepo
}

func NewSupplierService(repo *repository.SupplierRepo, orgs *repository.OrganizationRepo) *SupplierService {
return &SupplierService{repo: repo, orgs: orgs}
}

type SupplierInput struct {
Name    string
INN     *string
Phone   *string
Email   *string
Address *string
}

func (s *SupplierService) Create(ctx context.Context, in SupplierInput) (*models.Supplier, error) {
if in.Name == "" { return nil, apperr.BadRequest("name is required") }
sup := &models.Supplier{
Name:    in.Name,
INN:     in.INN,
Phone:   in.Phone,
Email:   in.Email,
Address: in.Address,
}
if err := s.repo.Create(ctx, sup); err != nil {
return nil, apperr.Internal("create supplier", err)
}
return sup, nil
}

func (s *SupplierService) Get(ctx context.Context, id uuid.UUID) (*models.Supplier, error) {
sup, err := s.repo.Get(ctx, id)
if err != nil { return nil, apperr.Internal("get supplier", err) }
if sup == nil { return nil, apperr.NotFound("supplier not found") }
return sup, nil
}

func (s *SupplierService) List(ctx context.Context) ([]models.Supplier, error) {
return s.repo.List(ctx)
}

func (s *SupplierService) Update(ctx context.Context, id uuid.UUID, in SupplierInput) (*models.Supplier, error) {
if in.Name == "" { return nil, apperr.BadRequest("name is required") }
sup, err := s.repo.Get(ctx, id)
if err != nil { return nil, apperr.Internal("get supplier", err) }
if sup == nil { return nil, apperr.NotFound("supplier not found") }

sup.Name = in.Name
sup.INN = in.INN
sup.Phone = in.Phone
sup.Email = in.Email
sup.Address = in.Address

if err := s.repo.Update(ctx, sup); err != nil {
return nil, apperr.Internal("update supplier", err)
}
return sup, nil
}

func (s *SupplierService) Delete(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.Delete(ctx, id)
if err != nil { return apperr.Internal("delete supplier", err) }
if !ok { return apperr.NotFound("supplier not found") }
return nil
}

func (s *SupplierService) ListOrganizations(ctx context.Context) ([]models.Organization, error) {
return s.orgs.List(ctx)
}

func (s *SupplierService) DefaultOrganization(ctx context.Context) (*models.Organization, error) {
org, err := s.orgs.Default(ctx)
if err != nil { return nil, apperr.Internal("default org", err) }
return org, nil
}