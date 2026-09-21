package service

import (
"context"

"github.com/google/uuid"

"github.com/radonezhsklad/warehouse/internal/models"
"github.com/radonezhsklad/warehouse/internal/repository"
apperr "github.com/radonezhsklad/shared/errors"
)

type InternalOrderService struct {
repo *repository.InternalOrderRepo
}

func NewInternalOrderService(repo *repository.InternalOrderRepo) *InternalOrderService {
return &InternalOrderService{repo: repo}
}

func (s *InternalOrderService) NextNumber(ctx context.Context) (string, error) {
return s.repo.NextNumber(ctx)
}

func (s *InternalOrderService) Create(ctx context.Context, in repository.InternalOrderInput) (*models.InternalOrder, error) {
if len(in.Items) == 0 {
return nil, apperr.BadRequest("order must have at least one item")
}
if in.Number == "" {
n, err := s.repo.NextNumber(ctx)
if err != nil { return nil, apperr.Internal("next number", err) }
in.Number = n
}
o, err := s.repo.Create(ctx, in)
if err != nil { return nil, apperr.Internal("create order", err) }
return o, nil
}

func (s *InternalOrderService) Get(ctx context.Context, id uuid.UUID) (*models.InternalOrder, error) {
o, err := s.repo.Get(ctx, id)
if err != nil { return nil, apperr.Internal("get order", err) }
if o == nil { return nil, apperr.NotFound("order not found") }
return o, nil
}

func (s *InternalOrderService) List(ctx context.Context, f repository.InternalOrderFilters) ([]models.InternalOrder, error) {
return s.repo.List(ctx, f)
}

func (s *InternalOrderService) Update(ctx context.Context, id uuid.UUID, in repository.InternalOrderInput) (*models.InternalOrder, error) {
o, err := s.Get(ctx, id)
if err != nil { return nil, err }
if o.Status != "draft" { return nil, apperr.Conflict("only draft orders can be edited") }
if len(in.Items) == 0 { return nil, apperr.BadRequest("order must have at least one item") }
res, err := s.repo.Update(ctx, id, in)
if err != nil { return nil, apperr.Internal("update order", err) }
return res, nil
}

func (s *InternalOrderService) Post(ctx context.Context, id uuid.UUID) (*models.InternalOrder, error) {
o, err := s.Get(ctx, id)
if err != nil { return nil, err }
if o.Status != "draft" { return nil, apperr.Conflict("only draft orders can be posted") }
if len(o.Items) == 0 { return nil, apperr.BadRequest("order has no items") }
if err := s.repo.UpdateStatus(ctx, id, "posted", true, false); err != nil {
return nil, apperr.Internal("post order", err)
}
return s.Get(ctx, id)
}

func (s *InternalOrderService) Cancel(ctx context.Context, id uuid.UUID) (*models.InternalOrder, error) {
o, err := s.Get(ctx, id)
if err != nil { return nil, err }
if o.Status != "posted" { return nil, apperr.Conflict("only posted orders can be cancelled") }
if err := s.repo.UpdateStatus(ctx, id, "cancelled", false, true); err != nil {
return nil, apperr.Internal("cancel order", err)
}
return s.Get(ctx, id)
}

func (s *InternalOrderService) Delete(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.Delete(ctx, id)
if err != nil { return apperr.Internal("delete order", err) }
if !ok { return apperr.Conflict("удалить можно только черновик; для проведённого используйте «Отменить»") }
return nil
}
func (s *InternalOrderService) MarkPrinted(ctx context.Context, id uuid.UUID) error {
return s.repo.MarkPrinted(ctx, id)
}

func (s *InternalOrderService) MarkSent(ctx context.Context, id uuid.UUID) error {
return s.repo.MarkSent(ctx, id)
}