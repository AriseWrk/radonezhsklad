package repository

import (
"context"
"errors"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/warehouse/internal/models"
)

type InternalOrderRepo struct{ db *pgxpool.Pool }

func NewInternalOrderRepo(db *pgxpool.Pool) *InternalOrderRepo {
return &InternalOrderRepo{db: db}
}

type InternalOrderInput struct {
Number         string
DocDate        *time.Time
OrganizationID *uuid.UUID
WarehouseID    *uuid.UUID
PlanDate       *time.Time
Project        *string
Comment        *string
VatEnabled     bool
VatIncluded    bool
CreatedBy      *uuid.UUID
Items          []InternalOrderItemInput
}

type InternalOrderItemInput struct {
ProductID uuid.UUID
Quantity  float64
Price     float64
VatRate   float64
}

const intOrderSelect = `
SELECT o.id, o.number, o.doc_date, o.status, o.organization_id, o.warehouse_id,
       o.plan_date, o.project, COALESCE(p.name, '') AS project_name, o.comment, o.total, o.shipped_amount,
       o.sent_at, o.printed_at, o.owner_id, o.owner_dept,
       o.vat_enabled, o.vat_included,
       o.posted_at, o.cancelled_at, o.created_by, o.created_at, o.updated_at, o.external_id,
       (SELECT COUNT(*) FROM internal_order_items i WHERE i.order_id = o.id)
FROM internal_orders o
LEFT JOIN projects p ON p.external_id::text = o.project`

func scanIntOrder(row pgx.Row) (*models.InternalOrder, error) {
o := &models.InternalOrder{}
err := row.Scan(&o.ID, &o.Number, &o.DocDate, &o.Status, &o.OrganizationID, &o.WarehouseID,
&o.PlanDate, &o.Project, &o.ProjectName, &o.Comment, &o.Total, &o.ShippedAmount,
&o.SentAt, &o.PrintedAt, &o.OwnerID, &o.OwnerDept,
&o.VatEnabled, &o.VatIncluded,
&o.PostedAt, &o.CancelledAt, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt, &o.ExternalID, &o.ItemsCount)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return o, err
}

func (r *InternalOrderRepo) Create(ctx context.Context, in InternalOrderInput) (*models.InternalOrder, error) {
var out *models.InternalOrder
err := r.WithTx(ctx, func(tx pgx.Tx) error {
docDate := time.Now()
if in.DocDate != nil { docDate = *in.DocDate }

var total float64
for _, it := range in.Items {
// цена считается с учётом НДС (цена уже включает или нет — упрощённо всегда сумма = qty*price)
total += it.Quantity * it.Price
}

var id uuid.UUID
err := tx.QueryRow(ctx, `
INSERT INTO internal_orders
 (number, doc_date, organization_id, warehouse_id, plan_date, project, comment,
  total, vat_enabled, vat_included, created_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
RETURNING id`,
in.Number, docDate, in.OrganizationID, in.WarehouseID, in.PlanDate, in.Project, in.Comment,
total, in.VatEnabled, in.VatIncluded, in.CreatedBy,
).Scan(&id)
if err != nil { return err }

for _, it := range in.Items {
sum := it.Quantity * it.Price
if _, err := tx.Exec(ctx, `
INSERT INTO internal_order_items (order_id, product_id, quantity, price, vat_rate, sum)
VALUES ($1,$2,$3,$4,$5,$6)`,
id, it.ProductID, it.Quantity, it.Price, it.VatRate, sum,
); err != nil { return err }
}

o, err := scanIntOrder(tx.QueryRow(ctx, intOrderSelect+` WHERE o.id = $1`, id))
if err != nil { return err }
items, err := loadIntItemsTx(ctx, tx, id)
if err != nil { return err }
o.Items = items
out = o
return nil
})
return out, err
}

func loadIntItemsTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) ([]models.InternalOrderItem, error) {
rows, err := tx.Query(ctx, `
SELECT id, order_id, product_id, quantity, price, vat_rate, sum, created_at
FROM internal_order_items WHERE order_id = $1 ORDER BY created_at`, id)
if err != nil { return nil, err }
defer rows.Close()
out := []models.InternalOrderItem{}
for rows.Next() {
var it models.InternalOrderItem
if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Quantity,
&it.Price, &it.VatRate, &it.Sum, &it.CreatedAt); err != nil { return nil, err }
out = append(out, it)
}
return out, rows.Err()
}

func (r *InternalOrderRepo) Get(ctx context.Context, id uuid.UUID) (*models.InternalOrder, error) {
o, err := scanIntOrder(r.db.QueryRow(ctx, intOrderSelect+` WHERE o.id = $1`, id))
if err != nil || o == nil { return o, err }
rows, err := r.db.Query(ctx, `
SELECT id, order_id, product_id, quantity, price, vat_rate, sum, created_at
FROM internal_order_items WHERE order_id = $1 ORDER BY created_at`, id)
if err != nil { return nil, err }
defer rows.Close()
out := []models.InternalOrderItem{}
for rows.Next() {
var it models.InternalOrderItem
if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Quantity,
&it.Price, &it.VatRate, &it.Sum, &it.CreatedAt); err != nil { return nil, err }
out = append(out, it)
}
o.Items = out
return o, rows.Err()
}

type InternalOrderFilters struct {
Status      *string
WarehouseID *uuid.UUID
}

func (r *InternalOrderRepo) List(ctx context.Context, f InternalOrderFilters) ([]models.InternalOrder, error) {
q := intOrderSelect + ` WHERE 1=1`
args := []any{}
i := 1
if f.Status != nil      { q += ` AND o.status = $` + itoa(i);       args = append(args, *f.Status); i++ }
if f.WarehouseID != nil { q += ` AND o.warehouse_id = $` + itoa(i); args = append(args, *f.WarehouseID); i++ }
q += ` ORDER BY o.created_at DESC LIMIT 500`

rows, err := r.db.Query(ctx, q, args...)
if err != nil { return nil, err }
defer rows.Close()
out := []models.InternalOrder{}
for rows.Next() {
o := &models.InternalOrder{}
if err := rows.Scan(&o.ID, &o.Number, &o.DocDate, &o.Status, &o.OrganizationID, &o.WarehouseID,
&o.PlanDate, &o.Project, &o.ProjectName, &o.Comment, &o.Total, &o.ShippedAmount,
&o.SentAt, &o.PrintedAt, &o.OwnerID, &o.OwnerDept,
&o.VatEnabled, &o.VatIncluded,
&o.PostedAt, &o.CancelledAt, &o.CreatedBy, &o.CreatedAt, &o.UpdatedAt, &o.ExternalID, &o.ItemsCount); err != nil {
return nil, err
}
out = append(out, *o)
}
return out, rows.Err()
}

func (r *InternalOrderRepo) Update(ctx context.Context, id uuid.UUID, in InternalOrderInput) (*models.InternalOrder, error) {
var out *models.InternalOrder
err := r.WithTx(ctx, func(tx pgx.Tx) error {
var total float64
for _, it := range in.Items { total += it.Quantity * it.Price }

_, err := tx.Exec(ctx, `
UPDATE internal_orders SET
 number=$2, organization_id=$3, warehouse_id=$4, plan_date=$5,
 project=$6, comment=$7, total=$8, vat_enabled=$9, vat_included=$10
WHERE id=$1`,
id, in.Number, in.OrganizationID, in.WarehouseID, in.PlanDate,
in.Project, in.Comment, total, in.VatEnabled, in.VatIncluded)
if err != nil { return err }

if _, err := tx.Exec(ctx, `DELETE FROM internal_order_items WHERE order_id=$1`, id); err != nil { return err }
for _, it := range in.Items {
sum := it.Quantity * it.Price
if _, err := tx.Exec(ctx, `
INSERT INTO internal_order_items (order_id, product_id, quantity, price, vat_rate, sum)
VALUES ($1,$2,$3,$4,$5,$6)`,
id, it.ProductID, it.Quantity, it.Price, it.VatRate, sum); err != nil { return err }
}

o, err := scanIntOrder(tx.QueryRow(ctx, intOrderSelect+` WHERE o.id = $1`, id))
if err != nil { return err }
items, err := loadIntItemsTx(ctx, tx, id)
if err != nil { return err }
o.Items = items
out = o
return nil
})
return out, err
}

func (r *InternalOrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, posted, cancelled bool) error {
var q string
switch {
case posted:     q = `UPDATE internal_orders SET status=$2, posted_at=NOW() WHERE id=$1`
case cancelled:  q = `UPDATE internal_orders SET status=$2, cancelled_at=NOW() WHERE id=$1`
default:         q = `UPDATE internal_orders SET status=$2 WHERE id=$1`
}
_, err := r.db.Exec(ctx, q, id, status)
return err
}

func (r *InternalOrderRepo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `DELETE FROM internal_orders WHERE id=$1 AND status='draft'`, id)
if err != nil { return false, err }
return tag.RowsAffected() > 0, nil
}

func (r *InternalOrderRepo) NextNumber(ctx context.Context) (string, error) {
var n int
err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(CAST(SUBSTRING(number FROM '[0-9]+$') AS INTEGER)), 0) + 1 FROM internal_orders`).Scan(&n)
if err != nil { return "", err }
return "ВЗ-" + itoa(n), nil
}

// Add WithTx helper to InternalOrderRepo
func (r *InternalOrderRepo) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
tx, err := r.db.Begin(ctx)
if err != nil { return err }
defer tx.Rollback(ctx)
if err := fn(tx); err != nil { return err }
return tx.Commit(ctx)
}
func (r *InternalOrderRepo) MarkPrinted(ctx context.Context, id uuid.UUID) error {
_, err := r.db.Exec(ctx, `UPDATE internal_orders SET printed_at = NOW() WHERE id = $1`, id)
return err
}

func (r *InternalOrderRepo) MarkSent(ctx context.Context, id uuid.UUID) error {
_, err := r.db.Exec(ctx, `UPDATE internal_orders SET sent_at = NOW() WHERE id = $1`, id)
return err
}