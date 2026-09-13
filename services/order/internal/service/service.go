package service

import (
"context"
"fmt"
"time"

"github.com/google/uuid"

"github.com/radonezhsklad/order/internal/models"
"github.com/radonezhsklad/order/internal/repository"
"github.com/radonezhsklad/order/internal/warehouse"
apperr "github.com/radonezhsklad/shared/errors"
)

type Service struct {
repo      *repository.Repo
warehouse *warehouse.Client
}

func New(repo *repository.Repo, wh *warehouse.Client) *Service {
return &Service{repo: repo, warehouse: wh}
}

// ---------- customers ----------

func (s *Service) CreateCustomer(ctx context.Context, name string, phone, email, address *string) (*models.Customer, error) {
if name == "" { return nil, apperr.BadRequest("name is required") }
return s.repo.CreateCustomer(ctx, name, phone, email, address)
}

func (s *Service) ListCustomers(ctx context.Context) ([]models.Customer, error) {
return s.repo.ListCustomers(ctx)
}

func (s *Service) GetCustomer(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
c, err := s.repo.GetCustomer(ctx, id)
if err != nil { return nil, apperr.Internal("get customer", err) }
if c == nil { return nil, apperr.NotFound("customer not found") }
return c, nil
}

func (s *Service) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
ok, err := s.repo.DeleteCustomer(ctx, id)
if err != nil { return apperr.Internal("delete customer", err) }
if !ok { return apperr.NotFound("customer not found") }
return nil
}

// ---------- orders ----------

func (s *Service) CreateOrder(ctx context.Context, in repository.OrderInput) (*models.Order, error) {
if len(in.Items) == 0 {
return nil, apperr.BadRequest("order must have at least one item")
}
if _, err := s.GetCustomer(ctx, in.CustomerID); err != nil {
return nil, err
}
if in.Number == "" {
in.Number = fmt.Sprintf("ORD-%d", time.Now().Unix())
}
o, err := s.repo.CreateOrder(ctx, in)
if err != nil { return nil, apperr.Internal("create order", err) }
return o, nil
}

func (s *Service) GetOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
o, err := s.repo.GetOrder(ctx, id)
if err != nil { return nil, apperr.Internal("get order", err) }
if o == nil { return nil, apperr.NotFound("order not found") }
return o, nil
}

func (s *Service) ListOrders(ctx context.Context, statusF *string, customerID *uuid.UUID) ([]models.Order, error) {
return s.repo.ListOrders(ctx, statusF, customerID)
}

func (s *Service) ConfirmOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
o, err := s.GetOrder(ctx, id)
if err != nil { return nil, err }
if o.Status != "draft" {
return nil, apperr.Conflict("only draft orders can be confirmed")
}
if err := s.repo.UpdateStatus(ctx, id, "confirmed", nil); err != nil {
return nil, apperr.Internal("confirm order", err)
}
return s.GetOrder(ctx, id)
}

// ShipOrder — интегрируется с warehouse: создаёт и проводит shipment.
// authToken — JWT текущего пользователя, прокидывается вниз.
func (s *Service) ShipOrder(ctx context.Context, id uuid.UUID, authToken string) (*models.Order, error) {
o, err := s.GetOrder(ctx, id)
if err != nil { return nil, err }
if o.Status != "confirmed" {
return nil, apperr.Conflict("only confirmed orders can be shipped")
}

items := make([]warehouse.ShipmentItem, 0, len(o.Items))
for _, it := range o.Items {
items = append(items, warehouse.ShipmentItem{
ProductID: it.ProductID,
Quantity:  it.Quantity,
Price:     it.Price,
})
}

docID, err := s.warehouse.CreateAndPostShipment(ctx, authToken, warehouse.CreateShipmentReq{
Type:        "shipment",
Number:      "SHIP-" + o.Number,
WarehouseID: o.WarehouseID,
Comment:     "Отгрузка по заказу " + o.Number,
Items:       items,
})
if err != nil {
// Ошибка warehouse — возвращаем 409, чтобы клиент понял: заказ не отгружен
return nil, apperr.Conflict("warehouse: " + err.Error())
}

if err := s.repo.UpdateStatus(ctx, id, "shipped", &docID); err != nil {
return nil, apperr.Internal("mark shipped", err)
}
return s.GetOrder(ctx, id)
}

func (s *Service) CancelOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
o, err := s.GetOrder(ctx, id)
if err != nil { return nil, err }
if o.Status == "shipped" {
return nil, apperr.Conflict("shipped orders cannot be cancelled")
}
if o.Status == "cancelled" {
return nil, apperr.Conflict("order already cancelled")
}
if err := s.repo.UpdateStatus(ctx, id, "cancelled", nil); err != nil {
return nil, apperr.Internal("cancel order", err)
}
return s.GetOrder(ctx, id)
}
// ---------- analytics ----------

func (s *Service) SalesAnalytics(ctx context.Context, days int) ([]repository.SalesRow, error) {
if days <= 0 { days = 14 }
if days > 365 { days = 365 }
return s.repo.SalesAnalytics(ctx, days)
}