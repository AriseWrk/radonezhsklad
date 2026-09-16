package service

import (
"context"
	"time"

"github.com/google/uuid"

"github.com/radonezhsklad/order/internal/models"
"github.com/radonezhsklad/order/internal/repository"
apperr "github.com/radonezhsklad/shared/errors"
)

type ContractService struct{ repo *repository.ContractRepo }

func NewContractService(repo *repository.ContractRepo) *ContractService {
return &ContractService{repo: repo}
}

func (s *ContractService) Create(ctx context.Context, c *models.Contract) (*models.Contract, error) {
if c.Number == "" { return nil, apperr.BadRequest("number is required") }
if err := s.repo.Create(ctx, c); err != nil {
return nil, apperr.Internal("create contract", err)
}
return c, nil
}

func (s *ContractService) List(ctx context.Context, includeArchived bool) ([]models.Contract, error) {
return s.repo.List(ctx, includeArchived)
}

func (s *ContractService) Get(ctx context.Context, id uuid.UUID) (*models.Contract, error) {
c, err := s.repo.Get(ctx, id)
if err != nil { return nil, apperr.Internal("get contract", err) }
if c == nil { return nil, apperr.NotFound("contract not found") }
return c, nil
}

func (s *ContractService) Update(ctx context.Context, id uuid.UUID, c *models.Contract) (*models.Contract, error) {
existing, err := s.repo.Get(ctx, id)
if err != nil { return nil, apperr.Internal("get contract", err) }
if existing == nil { return nil, apperr.NotFound("contract not found") }
c.ID = id
if err := s.repo.Update(ctx, c); err != nil {
return nil, apperr.Internal("update contract", err)
}
return s.repo.Get(ctx, id)
}

func (s *ContractService) Delete(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.Delete(ctx, id)
if err != nil { return apperr.Internal("delete contract", err) }
if !ok { return apperr.NotFound("contract not found") }
return nil
}
func (s *ContractService) Neighbors(ctx context.Context, id uuid.UUID) (*repository.ContractNeighbor, error) {
nb, err := s.repo.Neighbors(ctx, id)
if err != nil { return nil, apperr.Internal("neighbors", err) }
return nb, nil
}

func (s *ContractService) GetNext(ctx context.Context, docDate time.Time, id uuid.UUID) (*models.Contract, error) {
c, err := s.repo.GetNext(ctx, docDate, id)
if err != nil { return nil, apperr.Internal("next", err) }
return c, nil
}

func (s *ContractService) GetPrev(ctx context.Context, docDate time.Time, id uuid.UUID) (*models.Contract, error) {
c, err := s.repo.GetPrev(ctx, docDate, id)
if err != nil { return nil, apperr.Internal("prev", err) }
return c, nil
}