package service

import (
    "context"

    "github.com/google/uuid"

    "github.com/radonezhsklad/warehouse/internal/models"
    "github.com/radonezhsklad/warehouse/internal/repository"
    apperr "github.com/radonezhsklad/shared/errors"
)

type InventoryService struct {
    repo *repository.InventoryRepo
}

func NewInventoryService(repo *repository.InventoryRepo) *InventoryService {
    return &InventoryService{repo: repo}
}

func (s *InventoryService) List(ctx context.Context, f repository.InventoryFilters) ([]models.Inventory, error) {
    return s.repo.List(ctx, f)
}

func (s *InventoryService) Get(ctx context.Context, id uuid.UUID) (*models.Inventory, error) {
    inv, err := s.repo.Get(ctx, id)
    if err != nil { return nil, apperr.Internal("get inventory", err) }
    if inv == nil { return nil, apperr.NotFound("inventory not found") }
    return inv, nil
}

func (s *InventoryService) Neighbors(ctx context.Context, id uuid.UUID) (*repository.InventoryNeighbors, error) {
    n, err := s.repo.Neighbors(ctx, id)
    if err != nil { return nil, apperr.Internal("get inventory neighbors", err) }
    if n == nil { return nil, apperr.NotFound("inventory not found") }
    return n, nil
}
