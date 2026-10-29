package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	apperr "github.com/radonezhsklad/shared/errors"
	"github.com/radonezhsklad/warehouse/internal/models"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

type InternalOrderService struct {
	repo      *repository.InternalOrderRepo
	issueRepo *repository.InternalOrderIssueRepo
	stockRepo *repository.Repo
}

func NewInternalOrderService(repo *repository.InternalOrderRepo, issueRepo *repository.InternalOrderIssueRepo, stockRepo *repository.Repo) *InternalOrderService {
	return &InternalOrderService{repo: repo, issueRepo: issueRepo, stockRepo: stockRepo}
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
		if err != nil {
			return nil, apperr.Internal("next number", err)
		}
		in.Number = n
	}
	o, err := s.repo.Create(ctx, in)
	if err != nil {
		return nil, apperr.Internal("create order", err)
	}
	return o, nil
}

func (s *InternalOrderService) Get(ctx context.Context, id uuid.UUID) (*models.InternalOrder, error) {
	o, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, apperr.Internal("get order", err)
	}
	if o == nil {
		return nil, apperr.NotFound("order not found")
	}
	return o, nil
}

func (s *InternalOrderService) List(ctx context.Context, f repository.InternalOrderFilters) ([]models.InternalOrder, error) {
	return s.repo.List(ctx, f)
}

func (s *InternalOrderService) Update(ctx context.Context, id uuid.UUID, in repository.InternalOrderInput) (*models.InternalOrder, error) {
	o, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.Status != "draft" {
		return nil, apperr.Conflict("only draft orders can be edited")
	}
	if len(in.Items) == 0 {
		return nil, apperr.BadRequest("order must have at least one item")
	}
	res, err := s.repo.Update(ctx, id, in)
	if err != nil {
		return nil, apperr.Internal("update order", err)
	}
	return res, nil
}

func (s *InternalOrderService) Post(ctx context.Context, id uuid.UUID) (*models.InternalOrder, error) {
	o, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.Status != "draft" {
		return nil, apperr.Conflict("only draft orders can be posted")
	}
	if len(o.Items) == 0 {
		return nil, apperr.BadRequest("order has no items")
	}
	if o.WarehouseID == nil {
		return nil, apperr.BadRequest("у заказа не указан склад-источник")
	}

	err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		for _, it := range o.Items {
			avail, err := s.stockRepo.GetBalance(ctx, tx, *o.WarehouseID, it.ProductID)
			if err != nil {
				return apperr.Internal("get balance", err)
			}
			reserved, err := s.stockRepo.GetReserved(ctx, tx, *o.WarehouseID, it.ProductID)
			if err != nil {
				return apperr.Internal("get reserved", err)
			}
			if avail-reserved < it.Quantity {
				return apperr.Conflict("недостаточно свободного остатка для резерва по позиции " + it.ProductID.String())
			}
			if err := s.stockRepo.AddReserve(ctx, tx, *o.WarehouseID, it.ProductID, o.ID, it.Quantity); err != nil {
				return apperr.Internal("add reserve", err)
			}
		}
		return s.repo.SetStatusTx(ctx, tx, id, "posted")
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

func (s *InternalOrderService) Cancel(ctx context.Context, id uuid.UUID) (*models.InternalOrder, error) {
	o, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.Status != "posted" && o.Status != "shipped" {
		return nil, apperr.Conflict("only posted or shipped orders can be cancelled")
	}
	err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.stockRepo.ReleaseReserve(ctx, tx, id); err != nil {
			return apperr.Internal("release reserve", err)
		}
		return s.repo.SetStatusTx(ctx, tx, id, "cancelled")
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}
func (s *InternalOrderService) Delete(ctx context.Context, id uuid.UUID, force bool) error {
	ok, err := s.repo.Delete(ctx, id, force)
	if err != nil {
		return apperr.Internal("delete order", err)
	}
	if !ok {
		return apperr.Conflict("удалить можно только черновик; для проведённого используйте «Отменить»")
	}
	return nil
}

func (s *InternalOrderService) MarkPrinted(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkPrinted(ctx, id)
}

func (s *InternalOrderService) MarkSent(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkSent(ctx, id)
}

// Issue фиксирует фактическую выдачу по позиции заказа и списывает остаток со склада.
// Разрешено только для проведённого/отгруженного заказа.
func (s *InternalOrderService) Issue(ctx context.Context, orderID uuid.UUID, in repository.InternalOrderIssueInput) (*models.InternalOrderIssue, error) {
	o, err := s.Get(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != "posted" && o.Status != "shipped" {
		return nil, apperr.Conflict("выдача доступна только для проведённого заказа")
	}
	if o.WarehouseID == nil {
		return nil, apperr.BadRequest("у заказа не указан склад")
	}

	var iss *models.InternalOrderIssue
	err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		var err error
		iss, err = s.issueRepo.AddIssueTx(ctx, tx, orderID, in)
		if err != nil {
			if err == repository.ErrIssueOverflow {
				return apperr.BadRequest("выдача превышает остаток по позиции")
			}
			return apperr.Internal("add issue", err)
		}
		var productID uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT product_id FROM internal_order_items WHERE id = $1`, in.ItemID).Scan(&productID); err != nil {
			return apperr.Internal("get product_id", err)
		}
		if err := s.stockRepo.ApplyMovement(ctx, tx, *o.WarehouseID, productID, -in.Qty, orderID, "internal_order_issue", in.IssuedBy, o.WarehouseID, nil); err != nil {
			return apperr.Internal("apply issue movement", err)
		}
		if err := s.stockRepo.ReleaseReserveItem(ctx, tx, orderID, productID, in.Qty); err != nil {
			return apperr.Internal("release reserve item", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Автопереход статуса: shipped при первой выдаче, closed при полной.
	o2, err := s.Get(ctx, orderID)
	if err == nil && o2 != nil {
		newStatus := o2.Status
		switch {
		case o2.IssueStatus == "full" && o2.Status != "closed":
			newStatus = "closed"
		case (o2.IssueStatus == "partial" || o2.IssueStatus == "full") && o2.Status == "posted":
			newStatus = "shipped"
		}
		if newStatus != o2.Status {
			_ = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
				return s.repo.SetStatusTx(ctx, tx, orderID, newStatus)
			})
		}
	}
	return iss, nil
}
func (s *InternalOrderService) ListIssues(ctx context.Context, orderID uuid.UUID) ([]models.InternalOrderIssue, error) {
	if _, err := s.Get(ctx, orderID); err != nil {
		return nil, err
	}
	return s.issueRepo.ListByOrder(ctx, orderID)
}
