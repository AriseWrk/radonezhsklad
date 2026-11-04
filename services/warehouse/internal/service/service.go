package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	apperr "github.com/radonezhsklad/shared/errors"
	"github.com/radonezhsklad/warehouse/internal/models"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

type Service struct{ repo *repository.Repo }

func New(repo *repository.Repo) *Service { return &Service{repo: repo} }

// ---------- warehouses ----------

func (s *Service) CreateWarehouse(ctx context.Context, name string, address *string) (*models.Warehouse, error) {
	if name == "" {
		return nil, apperr.BadRequest("name is required")
	}
	return s.repo.CreateWarehouse(ctx, name, address)
}

func (s *Service) ListWarehouses(ctx context.Context) ([]models.Warehouse, error) {
	return s.repo.ListWarehouses(ctx)
}

func (s *Service) GetWarehouse(ctx context.Context, id uuid.UUID) (*models.Warehouse, error) {
	w, err := s.repo.GetWarehouse(ctx, id)
	if err != nil {
		return nil, apperr.Internal("get warehouse", err)
	}
	if w == nil {
		return nil, apperr.NotFound("warehouse not found")
	}
	return w, nil
}

func (s *Service) UpdateWarehouse(ctx context.Context, id uuid.UUID, name string, address *string, isActive bool) (*models.Warehouse, error) {
	if name == "" {
		return nil, apperr.BadRequest("name is required")
	}
	w, err := s.repo.UpdateWarehouse(ctx, id, name, address, isActive)
	if err != nil {
		return nil, apperr.Internal("update warehouse", err)
	}
	if w == nil {
		return nil, apperr.NotFound("warehouse not found")
	}
	return w, nil
}

func (s *Service) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
	ok, err := s.repo.DeleteWarehouse(ctx, id)
	if err != nil {
		return apperr.Internal("delete warehouse", err)
	}
	if !ok {
		return apperr.NotFound("warehouse not found")
	}
	return nil
}

// ---------- stock ----------

func (s *Service) ListStock(ctx context.Context, warehouseID, productID *uuid.UUID) ([]models.StockBalance, error) {
	return s.repo.ListStock(ctx, warehouseID, productID)
}

func (s *Service) StockExtended(ctx context.Context, f repository.StockExtendedFilter) ([]repository.ExtendedRow, error) {
	return s.repo.StockExtended(ctx, f)
}

func (s *Service) IncomingByProduct(ctx context.Context) (map[uuid.UUID]float64, error) {
	return s.repo.IncomingByProduct(ctx)
}

func (s *Service) BookStockForInventory(ctx context.Context, warehouseID uuid.UUID) ([]repository.InventoryRow, error) {
	return s.repo.BookStockForInventory(ctx, warehouseID)
}

// ---------- documents ----------

var validTypes = map[string]bool{
	"receipt": true, "shipment": true, "transfer": true, "writeoff": true, "inventory": true,
	"purchase_order": true, "invoice_in": true, "purchase_return": true, "facture_in": true,
}

// requiresWarehouse — типы, для которых warehouse_id обязателен.
var requiresWarehouse = map[string]bool{
	"receipt": true, "shipment": true, "transfer": true, "writeoff": true, "inventory": true,
	"purchase_return": true,
}

// requiresSupplier — типы, для которых нужен контрагент (поставщик).
var requiresSupplier = map[string]bool{
	"purchase_order": true, "invoice_in": true, "purchase_return": true,
}

func (s *Service) CreateDocument(ctx context.Context, in repository.DocumentInput) (*models.Document, error) {
	if !validTypes[in.Type] {
		return nil, apperr.BadRequest("invalid document type")
	}
	if len(in.Items) == 0 {
		return nil, apperr.BadRequest("document must have at least one item")
	}
	if requiresWarehouse[in.Type] && in.WarehouseID == nil {
		return nil, apperr.BadRequest(in.Type + " requires warehouse_id")
	}
	if requiresSupplier[in.Type] && in.SupplierID == nil {
		return nil, apperr.BadRequest(in.Type + " requires supplier_id")
	}
	if in.Type == "transfer" {
		if in.TargetWarehouseID == nil {
			return nil, apperr.BadRequest("transfer requires target_warehouse_id")
		}
		if in.WarehouseID == nil || *in.TargetWarehouseID == *in.WarehouseID {
			return nil, apperr.BadRequest("source and target warehouses must differ")
		}
	}
	if in.Number == "" {
		in.Number = fmt.Sprintf("%s-%d", in.Type[:3], time.Now().Unix())
	}
	d, err := s.repo.CreateDocument(ctx, in)
	if err != nil {
		return nil, apperr.Internal("create document", err)
	}
	return d, nil
}

func (s *Service) GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	d, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		return nil, apperr.Internal("get document", err)
	}
	if d == nil {
		return nil, apperr.NotFound("document not found")
	}
	return d, nil
}

func (s *Service) ListDocuments(ctx context.Context, f repository.DocumentFilters) ([]models.Document, error) {
	return s.repo.ListDocuments(ctx, f)
}

func (s *Service) PostDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	d, err := s.GetDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.Status != "draft" {
		return nil, apperr.Conflict("only draft documents can be posted")
	}
	if len(d.Items) == 0 {
		return nil, apperr.BadRequest("document has no items")
	}
	if requiresWarehouse[d.Type] && d.WarehouseID == nil {
		return nil, apperr.BadRequest(d.Type + ": warehouse_id не задан")
	}

	// Проверка доступности для расходных операций (учитывая резервы)
	if d.Type == "shipment" || d.Type == "transfer" || d.Type == "purchase_return" {
		err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
			for _, it := range d.Items {
				avail, err := s.repo.GetBalance(ctx, tx, *d.WarehouseID, it.ProductID)
				if err != nil {
					return apperr.Internal("check balance", err)
				}
				reserved, err := s.repo.GetReserved(ctx, tx, *d.WarehouseID, it.ProductID)
				if err != nil {
					return apperr.Internal("check reserved", err)
				}
				if avail-reserved < it.Quantity {
					return apperr.Conflict(fmt.Sprintf("недостаточно товара %s: доступно %.3f (с учётом резерва), требуется %.3f",
						it.ProductID, avail-reserved, it.Quantity))
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		for _, it := range d.Items {
			switch d.Type {
			case "receipt":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, nil, d.WarehouseID); err != nil {
					return apperr.Internal("apply receipt", err)
				}
			case "shipment", "purchase_return":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, -it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, nil); err != nil {
					return apperr.Internal("apply shipment/return", err)
				}
			case "writeoff":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, -it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, nil); err != nil {
					return apperr.Internal("apply writeoff", err)
				}
			case "transfer":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, -it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, d.TargetWarehouseID); err != nil {
					return apperr.Internal("apply transfer out", err)
				}
				if err := s.repo.ApplyMovement(ctx, tx, *d.TargetWarehouseID, it.ProductID, it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, d.TargetWarehouseID); err != nil {
					return apperr.Internal("apply transfer in", err)
				}
			case "inventory":
				cur, err := s.repo.GetBalance(ctx, tx, *d.WarehouseID, it.ProductID)
				if err != nil {
					return apperr.Internal("inventory balance", err)
				}
				delta := it.Quantity - cur
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, delta, &d.ID, nil, d.Type, d.CreatedBy, nil, d.WarehouseID); err != nil {
					return apperr.Internal("apply inventory", err)
				}
			case "purchase_order", "invoice_in", "facture_in":
				// не двигает склад
			}
		}
		if err := s.repo.UpdateDocStatusTx(ctx, tx, d.ID, "posted", true, false); err != nil {
			return apperr.Internal("update status", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetDocument(ctx, id)
}

func (s *Service) CancelDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	d, err := s.GetDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.Status != "posted" {
		return nil, apperr.Conflict("only posted documents can be cancelled")
	}
	if requiresWarehouse[d.Type] && d.WarehouseID == nil {
		return nil, apperr.BadRequest(d.Type + ": warehouse_id не задан")
	}

	err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
		for _, it := range d.Items {
			switch d.Type {
			case "receipt":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, -it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, nil, d.WarehouseID); err != nil {
					return apperr.Internal("revert receipt", err)
				}
			case "shipment", "purchase_return":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, nil); err != nil {
					return apperr.Internal("revert shipment/return", err)
				}
			case "writeoff":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, nil); err != nil {
					return apperr.Internal("revert writeoff", err)
				}
			case "transfer":
				if err := s.repo.ApplyMovement(ctx, tx, *d.WarehouseID, it.ProductID, it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, d.TargetWarehouseID); err != nil {
					return apperr.Internal("revert transfer out", err)
				}
				if err := s.repo.ApplyMovement(ctx, tx, *d.TargetWarehouseID, it.ProductID, -it.Quantity, &d.ID, nil, d.Type, d.CreatedBy, d.WarehouseID, d.TargetWarehouseID); err != nil {
					return apperr.Internal("revert transfer in", err)
				}
			case "inventory":
				return apperr.Conflict("inventory documents cannot be cancelled")
			case "purchase_order", "invoice_in", "facture_in":
				// не двигает склад
			}
		}
		if err := s.repo.UpdateDocStatusTx(ctx, tx, d.ID, "cancelled", false, true); err != nil {
			return apperr.Internal("update status", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetDocument(ctx, id)
}

// ---------- product stock detail ----------

func (s *Service) MovementsByProduct(ctx context.Context, productID uuid.UUID) ([]repository.MovementDetail, error) {
	return s.repo.MovementsByProduct(ctx, productID)
}

func (s *Service) ReservedByProduct(ctx context.Context, productID uuid.UUID) (map[uuid.UUID]float64, error) {
	return s.repo.ReservedByProduct(ctx, productID)
}

func (s *Service) BalancesByProduct(ctx context.Context, productID uuid.UUID) ([]repository.BalanceByWarehouse, error) {
	return s.repo.BalancesByProduct(ctx, productID)
}

// ---------- trash ----------

func (s *Service) SoftDeleteDocument(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.SoftDeleteDocument(ctx, id)
if err != nil {
return apperr.Internal("soft delete document", err)
}
if !ok {
return apperr.Conflict("в корзину можно отправить только черновик")
}
return nil
}

func (s *Service) RestoreDocument(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.RestoreDocument(ctx, id)
if err != nil {
return apperr.Internal("restore document", err)
}
if !ok {
return apperr.NotFound("документ не в корзине")
}
return nil
}

func (s *Service) HardDeleteDocument(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.HardDeleteDocument(ctx, id)
if err != nil {
return apperr.Internal("hard delete document", err)
}
if !ok {
return apperr.NotFound("документ не в корзине")
}
return nil
}

func (s *Service) ClearTrashDocuments(ctx context.Context) (int64, error) {
n, err := s.repo.HardDeleteAllTrashDocuments(ctx)
if err != nil {
return 0, apperr.Internal("clear trash documents", err)
}
return n, nil
}