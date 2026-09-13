package service

import (
"context"
"fmt"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"

"github.com/radonezhsklad/warehouse/internal/models"
"github.com/radonezhsklad/warehouse/internal/repository"
apperr "github.com/radonezhsklad/shared/errors"
)

type Service struct{ repo *repository.Repo }

func New(repo *repository.Repo) *Service { return &Service{repo: repo} }

// ---------- warehouses ----------

func (s *Service) CreateWarehouse(ctx context.Context, name string, address *string) (*models.Warehouse, error) {
if name == "" { return nil, apperr.BadRequest("name is required") }
return s.repo.CreateWarehouse(ctx, name, address)
}

func (s *Service) ListWarehouses(ctx context.Context) ([]models.Warehouse, error) {
return s.repo.ListWarehouses(ctx)
}

func (s *Service) GetWarehouse(ctx context.Context, id uuid.UUID) (*models.Warehouse, error) {
w, err := s.repo.GetWarehouse(ctx, id)
if err != nil { return nil, apperr.Internal("get warehouse", err) }
if w == nil { return nil, apperr.NotFound("warehouse not found") }
return w, nil
}

func (s *Service) UpdateWarehouse(ctx context.Context, id uuid.UUID, name string, address *string, isActive bool) (*models.Warehouse, error) {
if name == "" { return nil, apperr.BadRequest("name is required") }
w, err := s.repo.UpdateWarehouse(ctx, id, name, address, isActive)
if err != nil { return nil, apperr.Internal("update warehouse", err) }
if w == nil { return nil, apperr.NotFound("warehouse not found") }
return w, nil
}

func (s *Service) DeleteWarehouse(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.DeleteWarehouse(ctx, id)
if err != nil { return apperr.Internal("delete warehouse", err) }
if !ok { return apperr.NotFound("warehouse not found") }
return nil
}

// ---------- stock ----------

func (s *Service) ListStock(ctx context.Context, warehouseID, productID *uuid.UUID) ([]models.StockBalance, error) {
return s.repo.ListStock(ctx, warehouseID, productID)
}

func (s *Service) StockExtended(ctx context.Context, warehouseID *uuid.UUID) ([]repository.ExtendedRow, error) {
return s.repo.StockExtended(ctx, warehouseID)
}

func (s *Service) IncomingByProduct(ctx context.Context) (map[uuid.UUID]float64, error) {
return s.repo.IncomingByProduct(ctx)
}

func (s *Service) BookStockForInventory(ctx context.Context, warehouseID uuid.UUID) ([]repository.InventoryRow, error) {
return s.repo.BookStockForInventory(ctx, warehouseID)
}

// ---------- documents ----------

var validTypes = map[string]bool{"receipt": true, "shipment": true, "transfer": true, "inventory": true}

func (s *Service) CreateDocument(ctx context.Context, in repository.DocumentInput) (*models.Document, error) {
if !validTypes[in.Type] { return nil, apperr.BadRequest("invalid document type") }
if len(in.Items) == 0 { return nil, apperr.BadRequest("document must have at least one item") }
if in.Type == "transfer" {
if in.TargetWarehouseID == nil { return nil, apperr.BadRequest("transfer requires target_warehouse_id") }
if *in.TargetWarehouseID == in.WarehouseID { return nil, apperr.BadRequest("source and target warehouses must differ") }
}
if in.Number == "" {
in.Number = fmt.Sprintf("%s-%d", in.Type[:3], time.Now().Unix())
}
d, err := s.repo.CreateDocument(ctx, in)
if err != nil { return nil, apperr.Internal("create document", err) }
return d, nil
}

func (s *Service) GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
d, err := s.repo.GetDocument(ctx, id)
if err != nil { return nil, apperr.Internal("get document", err) }
if d == nil { return nil, apperr.NotFound("document not found") }
return d, nil
}

func (s *Service) ListDocuments(ctx context.Context, f repository.DocumentFilters) ([]models.Document, error) {
return s.repo.ListDocuments(ctx, f)
}

func (s *Service) PostDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
d, err := s.GetDocument(ctx, id)
if err != nil { return nil, err }
if d.Status != "draft" { return nil, apperr.Conflict("only draft documents can be posted") }
if len(d.Items) == 0 { return nil, apperr.BadRequest("document has no items") }

if d.Type == "shipment" || d.Type == "transfer" {
err := s.repo.WithTx(ctx, func(tx pgx.Tx) error {
for _, it := range d.Items {
avail, err := s.repo.GetBalance(ctx, tx, d.WarehouseID, it.ProductID)
if err != nil { return apperr.Internal("check balance", err) }
if avail < it.Quantity {
return apperr.Conflict(fmt.Sprintf("недостаточно товара %s: доступно %.3f, требуется %.3f",
it.ProductID, avail, it.Quantity))
}
}
return nil
})
if err != nil { return nil, err }
}

err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
for _, it := range d.Items {
switch d.Type {
case "receipt":
if err := s.repo.ApplyMovement(ctx, tx, d.WarehouseID, it.ProductID, it.Quantity, d.ID); err != nil {
return apperr.Internal("apply receipt", err)
}
case "shipment":
if err := s.repo.ApplyMovement(ctx, tx, d.WarehouseID, it.ProductID, -it.Quantity, d.ID); err != nil {
return apperr.Internal("apply shipment", err)
}
case "transfer":
if err := s.repo.ApplyMovement(ctx, tx, d.WarehouseID, it.ProductID, -it.Quantity, d.ID); err != nil {
return apperr.Internal("apply transfer out", err)
}
if err := s.repo.ApplyMovement(ctx, tx, *d.TargetWarehouseID, it.ProductID, it.Quantity, d.ID); err != nil {
return apperr.Internal("apply transfer in", err)
}
case "inventory":
cur, err := s.repo.GetBalance(ctx, tx, d.WarehouseID, it.ProductID)
if err != nil { return apperr.Internal("inventory balance", err) }
delta := it.Quantity - cur
if err := s.repo.ApplyMovement(ctx, tx, d.WarehouseID, it.ProductID, delta, d.ID); err != nil {
return apperr.Internal("apply inventory", err)
}
}
}
return nil
})
if err != nil { return nil, err }

if err := s.repo.UpdateDocStatus(ctx, id, "posted", true, false); err != nil {
return nil, apperr.Internal("update status", err)
}
return s.GetDocument(ctx, id)
}

func (s *Service) CancelDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
d, err := s.GetDocument(ctx, id)
if err != nil { return nil, err }
if d.Status != "posted" { return nil, apperr.Conflict("only posted documents can be cancelled") }

err = s.repo.WithTx(ctx, func(tx pgx.Tx) error {
for _, it := range d.Items {
switch d.Type {
case "receipt":
if err := s.repo.ApplyMovement(ctx, tx, d.WarehouseID, it.ProductID, -it.Quantity, d.ID); err != nil {
return apperr.Internal("revert receipt", err)
}
case "shipment":
if err := s.repo.ApplyMovement(ctx, tx, d.WarehouseID, it.ProductID, it.Quantity, d.ID); err != nil {
return apperr.Internal("revert shipment", err)
}
case "transfer":
if err := s.repo.ApplyMovement(ctx, tx, d.WarehouseID, it.ProductID, it.Quantity, d.ID); err != nil {
return apperr.Internal("revert transfer out", err)
}
if err := s.repo.ApplyMovement(ctx, tx, *d.TargetWarehouseID, it.ProductID, -it.Quantity, d.ID); err != nil {
return apperr.Internal("revert transfer in", err)
}
case "inventory":
return apperr.Conflict("inventory documents cannot be cancelled")
}
}
return nil
})
if err != nil { return nil, err }

if err := s.repo.UpdateDocStatus(ctx, id, "cancelled", false, true); err != nil {
return nil, apperr.Internal("update status", err)
}
return s.GetDocument(ctx, id)
}