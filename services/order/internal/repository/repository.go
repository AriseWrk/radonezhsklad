package repository

import (
"context"
	"time"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/order/internal/models"
)

type Repo struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

func (r *Repo) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
tx, err := r.db.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx)
if err := fn(tx); err != nil { return err }
return tx.Commit(ctx)
}

// ---------- customers / counterparties ----------

const custCols = `id, name, full_name, last_name, first_name, middle_name,
                  phone, fax, email, address, legal_address, actual_address,
                  inn, kpp, ogrn, okpo, external_code, counterparty_type,
                  status, group_name, comment, archived, created_at, updated_at`

func scanCustomer(row pgx.Row) (*models.Customer, error) {
c := &models.Customer{}
err := row.Scan(&c.ID, &c.Name, &c.FullName, &c.LastName, &c.FirstName, &c.MiddleName,
&c.Phone, &c.Fax, &c.Email, &c.Address, &c.LegalAddress, &c.ActualAddress,
&c.INN, &c.KPP, &c.OGRN, &c.OKPO, &c.ExternalCode, &c.CounterpartyType,
&c.Status, &c.GroupName, &c.Comment, &c.Archived, &c.CreatedAt, &c.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return c, err
}

func (r *Repo) CreateCustomer(ctx context.Context, c *models.Customer) error {
if c.Status == "" { c.Status = "Новый" }
return r.db.QueryRow(ctx,
`INSERT INTO customers
 (name, full_name, last_name, first_name, middle_name,
  phone, fax, email, address, legal_address, actual_address,
  inn, kpp, ogrn, okpo, external_code, counterparty_type,
  status, group_name, comment, archived)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
 RETURNING id, created_at, updated_at`,
c.Name, c.FullName, c.LastName, c.FirstName, c.MiddleName,
c.Phone, c.Fax, c.Email, c.Address, c.LegalAddress, c.ActualAddress,
c.INN, c.KPP, c.OGRN, c.OKPO, c.ExternalCode, c.CounterpartyType,
c.Status, c.GroupName, c.Comment, c.Archived,
).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *Repo) ListCustomers(ctx context.Context, includeArchived bool) ([]models.Customer, error) {
q := `SELECT ` + custCols + ` FROM customers`
if !includeArchived { q += ` WHERE archived = FALSE` }
q += ` ORDER BY name LIMIT 2000`
rows, err := r.db.Query(ctx, q)
if err != nil { return nil, err }
defer rows.Close()
out := []models.Customer{}
for rows.Next() {
c, err := scanCustomer(rows)
if err != nil { return nil, err }
out = append(out, *c)
}
return out, rows.Err()
}

func (r *Repo) GetCustomer(ctx context.Context, id uuid.UUID) (*models.Customer, error) {
return scanCustomer(r.db.QueryRow(ctx, `SELECT `+custCols+` FROM customers WHERE id = $1`, id))
}

func (r *Repo) UpdateCustomer(ctx context.Context, c *models.Customer) error {
_, err := r.db.Exec(ctx,
`UPDATE customers SET
 name=$2, full_name=$3, last_name=$4, first_name=$5, middle_name=$6,
 phone=$7, fax=$8, email=$9, address=$10, legal_address=$11, actual_address=$12,
 inn=$13, kpp=$14, ogrn=$15, okpo=$16, external_code=$17, counterparty_type=$18,
 status=$19, group_name=$20, comment=$21, archived=$22
 WHERE id=$1`,
c.ID, c.Name, c.FullName, c.LastName, c.FirstName, c.MiddleName,
c.Phone, c.Fax, c.Email, c.Address, c.LegalAddress, c.ActualAddress,
c.INN, c.KPP, c.OGRN, c.OKPO, c.ExternalCode, c.CounterpartyType,
c.Status, c.GroupName, c.Comment, c.Archived)
return err
}

func (r *Repo) DeleteCustomer(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `DELETE FROM customers WHERE id = $1`, id)
if err != nil { return false, err }
return tag.RowsAffected() > 0, nil
}

// ---------- orders ----------

type OrderInput struct {
Number      string
CustomerID  uuid.UUID
WarehouseID uuid.UUID
Comment     *string
CreatedBy   *uuid.UUID
Items       []OrderItemInput
}

type OrderItemInput struct {
ProductID uuid.UUID
Quantity  float64
Price     float64
}

func (r *Repo) CreateOrder(ctx context.Context, in OrderInput) (*models.Order, error) {
var out *models.Order
err := r.WithTx(ctx, func(tx pgx.Tx) error {
var total float64
for _, it := range in.Items { total += it.Quantity * it.Price }

o := &models.Order{}
err := tx.QueryRow(ctx,
`INSERT INTO orders (number, customer_id, warehouse_id, comment, created_by, total)
 VALUES ($1,$2,$3,$4,$5,$6)
 RETURNING id, number, customer_id, warehouse_id, status, total, currency, comment,
           warehouse_doc_id, created_by, created_at, updated_at,
           confirmed_at, shipped_at, cancelled_at`,
in.Number, in.CustomerID, in.WarehouseID, in.Comment, in.CreatedBy, total,
).Scan(&o.ID, &o.Number, &o.CustomerID, &o.WarehouseID, &o.Status, &o.Total, &o.Currency,
&o.Comment, &o.WarehouseDocID, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
&o.ConfirmedAt, &o.ShippedAt, &o.CancelledAt)
if err != nil { return err }

for _, it := range in.Items {
if _, err := tx.Exec(ctx,
`INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1,$2,$3,$4)`,
o.ID, it.ProductID, it.Quantity, it.Price,
); err != nil { return err }
}

items, err := loadItemsTx(ctx, tx, o.ID)
if err != nil { return err }
o.Items = items
out = o
return nil
})
return out, err
}

func loadItemsTx(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) ([]models.OrderItem, error) {
rows, err := tx.Query(ctx,
`SELECT id, order_id, product_id, quantity, price, created_at
 FROM order_items WHERE order_id = $1`, orderID)
if err != nil { return nil, err }
defer rows.Close()
out := []models.OrderItem{}
for rows.Next() {
var it models.OrderItem
if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Quantity, &it.Price, &it.CreatedAt); err != nil {
return nil, err
}
out = append(out, it)
}
return out, rows.Err()
}

func (r *Repo) GetOrder(ctx context.Context, id uuid.UUID) (*models.Order, error) {
o := &models.Order{}
err := r.db.QueryRow(ctx,
`SELECT id, number, customer_id, warehouse_id, status, total, currency, comment,
        warehouse_doc_id, created_by, created_at, updated_at,
        confirmed_at, shipped_at, cancelled_at
 FROM orders WHERE id = $1`, id,
).Scan(&o.ID, &o.Number, &o.CustomerID, &o.WarehouseID, &o.Status, &o.Total, &o.Currency,
&o.Comment, &o.WarehouseDocID, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
&o.ConfirmedAt, &o.ShippedAt, &o.CancelledAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
if err != nil { return nil, err }

items, err := r.itemsByOrder(ctx, id)
if err != nil { return nil, err }
o.Items = items
return o, nil
}

func (r *Repo) itemsByOrder(ctx context.Context, orderID uuid.UUID) ([]models.OrderItem, error) {
rows, err := r.db.Query(ctx,
`SELECT id, order_id, product_id, quantity, price, created_at
 FROM order_items WHERE order_id = $1`, orderID)
if err != nil { return nil, err }
defer rows.Close()
out := []models.OrderItem{}
for rows.Next() {
var it models.OrderItem
if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Quantity, &it.Price, &it.CreatedAt); err != nil {
return nil, err
}
out = append(out, it)
}
return out, rows.Err()
}

func (r *Repo) ListOrders(ctx context.Context, statusFilter *string, customerID *uuid.UUID) ([]models.Order, error) {
q := `SELECT id, number, customer_id, warehouse_id, status, total, currency, comment,
             warehouse_doc_id, created_by, created_at, updated_at,
             confirmed_at, shipped_at, cancelled_at
      FROM orders WHERE 1=1`
args := []any{}
i := 1
if statusFilter != nil { q += ` AND status = $` + itoa(i); args = append(args, *statusFilter); i++ }
if customerID != nil   { q += ` AND customer_id = $` + itoa(i); args = append(args, *customerID); i++ }
q += ` ORDER BY created_at DESC LIMIT 200`

rows, err := r.db.Query(ctx, q, args...)
if err != nil { return nil, err }
defer rows.Close()
out := []models.Order{}
for rows.Next() {
var o models.Order
if err := rows.Scan(&o.ID, &o.Number, &o.CustomerID, &o.WarehouseID, &o.Status, &o.Total, &o.Currency,
&o.Comment, &o.WarehouseDocID, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt,
&o.ConfirmedAt, &o.ShippedAt, &o.CancelledAt); err != nil {
return nil, err
}
out = append(out, o)
}
return out, rows.Err()
}

func (r *Repo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, warehouseDocID *uuid.UUID) error {
var q string
switch status {
case "confirmed":
q = `UPDATE orders SET status = $2, confirmed_at = NOW() WHERE id = $1`
case "shipped":
q = `UPDATE orders SET status = $2, shipped_at = NOW(), warehouse_doc_id = $3 WHERE id = $1`
case "cancelled":
q = `UPDATE orders SET status = $2, cancelled_at = NOW() WHERE id = $1`
default:
q = `UPDATE orders SET status = $2 WHERE id = $1`
}
if status == "shipped" {
_, err := r.db.Exec(ctx, q, id, status, warehouseDocID)
return err
}
_, err := r.db.Exec(ctx, q, id, status)
return err
}


// ---------- analytics ----------

type SalesRow struct {
ProductID    uuid.UUID
SoldQty      float64
SoldSum      float64
OrdersCount  int
FirstSoldAt  *time.Time
LastSoldAt   *time.Time
}

// SalesAnalytics — агрегация проданного за N дней по shipped-заказам.
func (r *Repo) SalesAnalytics(ctx context.Context, days int) ([]SalesRow, error) {
rows, err := r.db.Query(ctx, `
SELECT oi.product_id,
       COALESCE(SUM(oi.quantity), 0)               AS sold_qty,
       COALESCE(SUM(oi.quantity * oi.price), 0)    AS sold_sum,
       COUNT(DISTINCT oi.order_id)                 AS orders_count,
       MIN(o.shipped_at)                           AS first_sold_at,
       MAX(o.shipped_at)                           AS last_sold_at
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE o.status = 'shipped'
  AND o.shipped_at IS NOT NULL
  AND o.shipped_at >= NOW() - make_interval(days => $1)
GROUP BY oi.product_id
ORDER BY sold_sum DESC`, days)
if err != nil { return nil, err }
defer rows.Close()

out := []SalesRow{}
for rows.Next() {
var s SalesRow
if err := rows.Scan(&s.ProductID, &s.SoldQty, &s.SoldSum, &s.OrdersCount,
&s.FirstSoldAt, &s.LastSoldAt); err != nil {
return nil, err
}
out = append(out, s)
}
return out, rows.Err()
}


type SalesDailyRow struct {
Day         time.Time
SoldQty     float64
SoldSum     float64
OrdersCount int
}

func (r *Repo) SalesDaily(ctx context.Context, days int) ([]SalesDailyRow, error) {
rows, err := r.db.Query(ctx, `
SELECT DATE(o.shipped_at) AS day,
       COALESCE(SUM(oi.quantity), 0)             AS sold_qty,
       COALESCE(SUM(oi.quantity * oi.price), 0)  AS sold_sum,
       COUNT(DISTINCT oi.order_id)               AS orders_count
FROM order_items oi
JOIN orders o ON o.id = oi.order_id
WHERE o.status = 'shipped'
  AND o.shipped_at IS NOT NULL
  AND o.shipped_at >= NOW() - make_interval(days => $1)
GROUP BY DATE(o.shipped_at)
ORDER BY day`, days)
if err != nil { return nil, err }
defer rows.Close()

out := []SalesDailyRow{}
for rows.Next() {
var s SalesDailyRow
if err := rows.Scan(&s.Day, &s.SoldQty, &s.SoldSum, &s.OrdersCount); err != nil { return nil, err }
out = append(out, s)
}
return out, rows.Err()
}

func itoa(n int) string {
if n == 0 { return "0" }
var buf [20]byte
i := len(buf)
for n > 0 { i--; buf[i] = byte('0' + n%10); n /= 10 }
return string(buf[i:])
}