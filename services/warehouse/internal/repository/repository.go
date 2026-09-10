package repository

import (
"context"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/warehouse/internal/models"
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

// ---------- warehouses ----------

func (r *Repo) CreateWarehouse(ctx context.Context, name string, address *string) (*models.Warehouse, error) {
w := &models.Warehouse{}
err := r.db.QueryRow(ctx,
`INSERT INTO warehouses (name, address) VALUES ($1, $2)
 RETURNING id, name, address, is_active, created_at, updated_at`,
name, address,
).Scan(&w.ID, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
return w, err
}

func (r *Repo) ListWarehouses(ctx context.Context) ([]models.Warehouse, error) {
rows, err := r.db.Query(ctx,
`SELECT id, name, address, is_active, created_at, updated_at
 FROM warehouses ORDER BY name`)
if err != nil { return nil, err }
defer rows.Close()

out := []models.Warehouse{}
for rows.Next() {
var w models.Warehouse
if err := rows.Scan(&w.ID, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
return nil, err
}
out = append(out, w)
}
return out, rows.Err()
}

func (r *Repo) GetWarehouse(ctx context.Context, id uuid.UUID) (*models.Warehouse, error) {
w := &models.Warehouse{}
err := r.db.QueryRow(ctx,
`SELECT id, name, address, is_active, created_at, updated_at FROM warehouses WHERE id = $1`,
id,
).Scan(&w.ID, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return w, err
}

func (r *Repo) UpdateWarehouse(ctx context.Context, id uuid.UUID, name string, address *string, isActive bool) (*models.Warehouse, error) {
w := &models.Warehouse{}
err := r.db.QueryRow(ctx,
`UPDATE warehouses SET name = $2, address = $3, is_active = $4 WHERE id = $1
 RETURNING id, name, address, is_active, created_at, updated_at`,
id, name, address, isActive,
).Scan(&w.ID, &w.Name, &w.Address, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return w, err
}

func (r *Repo) DeleteWarehouse(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `DELETE FROM warehouses WHERE id = $1`, id)
if err != nil { return false, err }
return tag.RowsAffected() > 0, nil
}

// ---------- stock ----------

func (r *Repo) ListStock(ctx context.Context, warehouseID, productID *uuid.UUID) ([]models.StockBalance, error) {
q := `SELECT id, warehouse_id, product_id, quantity, updated_at FROM stock_balances WHERE 1=1`
args := []any{}
i := 1
if warehouseID != nil {
q += ` AND warehouse_id = $` + itoa(i); args = append(args, *warehouseID); i++
}
if productID != nil {
q += ` AND product_id = $` + itoa(i); args = append(args, *productID); i++
}
q += ` ORDER BY updated_at DESC`

rows, err := r.db.Query(ctx, q, args...)
if err != nil { return nil, err }
defer rows.Close()

out := []models.StockBalance{}
for rows.Next() {
var b models.StockBalance
if err := rows.Scan(&b.ID, &b.WarehouseID, &b.ProductID, &b.Quantity, &b.UpdatedAt); err != nil {
return nil, err
}
out = append(out, b)
}
return out, rows.Err()
}

// ---------- documents ----------

type DocumentInput struct {
Type              string
Number            string
WarehouseID       uuid.UUID
TargetWarehouseID *uuid.UUID
Comment           *string
CreatedBy         *uuid.UUID
Items             []DocItemInput
}

type DocItemInput struct {
ProductID uuid.UUID
Quantity  float64
Price     float64
}

func (r *Repo) CreateDocument(ctx context.Context, in DocumentInput) (*models.Document, error) {
var out *models.Document
err := r.WithTx(ctx, func(tx pgx.Tx) error {
d := &models.Document{}
err := tx.QueryRow(ctx,
`INSERT INTO documents (type, number, warehouse_id, target_warehouse_id, comment, created_by)
 VALUES ($1, $2, $3, $4, $5, $6)
 RETURNING id, type, number, status, warehouse_id, target_warehouse_id, comment, created_by, created_at, updated_at`,
in.Type, in.Number, in.WarehouseID, in.TargetWarehouseID, in.Comment, in.CreatedBy,
).Scan(&d.ID, &d.Type, &d.Number, &d.Status, &d.WarehouseID, &d.TargetWarehouseID,
&d.Comment, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt)
if err != nil { return err }

for _, it := range in.Items {
if _, err := tx.Exec(ctx,
`INSERT INTO document_items (document_id, product_id, quantity, price)
 VALUES ($1, $2, $3, $4)`,
d.ID, it.ProductID, it.Quantity, it.Price,
); err != nil { return err }
}

items, err := loadItemsTx(ctx, tx, d.ID)
if err != nil { return err }
d.Items = items
out = d
return nil
})
return out, err
}

func loadItemsTx(ctx context.Context, tx pgx.Tx, docID uuid.UUID) ([]models.DocItem, error) {
rows, err := tx.Query(ctx,
`SELECT id, document_id, product_id, quantity, price, created_at
 FROM document_items WHERE document_id = $1`, docID)
if err != nil { return nil, err }
defer rows.Close()

out := []models.DocItem{}
for rows.Next() {
var it models.DocItem
if err := rows.Scan(&it.ID, &it.DocumentID, &it.ProductID, &it.Quantity, &it.Price, &it.CreatedAt); err != nil {
return nil, err
}
out = append(out, it)
}
return out, rows.Err()
}

func (r *Repo) GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
d := &models.Document{}
err := r.db.QueryRow(ctx,
`SELECT id, type, number, status, warehouse_id, target_warehouse_id, comment, created_by,
        created_at, updated_at, posted_at, cancelled_at
 FROM documents WHERE id = $1`, id,
).Scan(&d.ID, &d.Type, &d.Number, &d.Status, &d.WarehouseID, &d.TargetWarehouseID,
&d.Comment, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt, &d.PostedAt, &d.CancelledAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
if err != nil { return nil, err }

items, err := r.itemsByDoc(ctx, id)
if err != nil { return nil, err }
d.Items = items
return d, nil
}

func (r *Repo) itemsByDoc(ctx context.Context, docID uuid.UUID) ([]models.DocItem, error) {
rows, err := r.db.Query(ctx,
`SELECT id, document_id, product_id, quantity, price, created_at
 FROM document_items WHERE document_id = $1`, docID)
if err != nil { return nil, err }
defer rows.Close()

out := []models.DocItem{}
for rows.Next() {
var it models.DocItem
if err := rows.Scan(&it.ID, &it.DocumentID, &it.ProductID, &it.Quantity, &it.Price, &it.CreatedAt); err != nil {
return nil, err
}
out = append(out, it)
}
return out, rows.Err()
}

func (r *Repo) ListDocuments(ctx context.Context, typeFilter, statusFilter *string, warehouseID *uuid.UUID) ([]models.Document, error) {
q := `SELECT id, type, number, status, warehouse_id, target_warehouse_id, comment, created_by,
             created_at, updated_at, posted_at, cancelled_at
      FROM documents WHERE 1=1`
args := []any{}
i := 1
if typeFilter != nil {
q += ` AND type = $` + itoa(i); args = append(args, *typeFilter); i++
}
if statusFilter != nil {
q += ` AND status = $` + itoa(i); args = append(args, *statusFilter); i++
}
if warehouseID != nil {
q += ` AND warehouse_id = $` + itoa(i); args = append(args, *warehouseID); i++
}
q += ` ORDER BY created_at DESC LIMIT 200`

rows, err := r.db.Query(ctx, q, args...)
if err != nil { return nil, err }
defer rows.Close()

out := []models.Document{}
for rows.Next() {
var d models.Document
if err := rows.Scan(&d.ID, &d.Type, &d.Number, &d.Status, &d.WarehouseID, &d.TargetWarehouseID,
&d.Comment, &d.CreatedBy, &d.CreatedAt, &d.UpdatedAt, &d.PostedAt, &d.CancelledAt); err != nil {
return nil, err
}
out = append(out, d)
}
return out, rows.Err()
}

func (r *Repo) UpdateDocStatus(ctx context.Context, id uuid.UUID, status string, setPosted, setCancelled bool) error {
var q string
switch {
case setPosted:
q = `UPDATE documents SET status = $2, posted_at = NOW() WHERE id = $1`
case setCancelled:
q = `UPDATE documents SET status = $2, cancelled_at = NOW() WHERE id = $1`
default:
q = `UPDATE documents SET status = $2 WHERE id = $1`
}
_, err := r.db.Exec(ctx, q, id, status)
return err
}

// ---------- stock movements / apply ----------

// ApplyMovementUPSERT — атомарно изменяет остаток на складе.
func (r *Repo) ApplyMovement(ctx context.Context, tx pgx.Tx, warehouseID, productID uuid.UUID, delta float64, docID uuid.UUID) error {
_, err := tx.Exec(ctx,
`INSERT INTO stock_balances (warehouse_id, product_id, quantity)
 VALUES ($1, $2, $3)
 ON CONFLICT (warehouse_id, product_id)
 DO UPDATE SET quantity = stock_balances.quantity + EXCLUDED.quantity,
               updated_at = NOW()`,
warehouseID, productID, delta,
)
if err != nil { return err }

_, err = tx.Exec(ctx,
`INSERT INTO stock_movements (warehouse_id, product_id, document_id, quantity_delta)
 VALUES ($1, $2, $3, $4)`,
warehouseID, productID, docID, delta,
)
return err
}

// GetBalanceForUpdate — получить текущий остаток для проверки.
func (r *Repo) GetBalance(ctx context.Context, tx pgx.Tx, warehouseID, productID uuid.UUID) (float64, error) {
var q float64
err := tx.QueryRow(ctx,
`SELECT quantity FROM stock_balances WHERE warehouse_id = $1 AND product_id = $2`,
warehouseID, productID,
).Scan(&q)
if errors.Is(err, pgx.ErrNoRows) { return 0, nil }
return q, err
}

func itoa(n int) string {
if n == 0 { return "0" }
var buf [20]byte
i := len(buf)
for n > 0 { i--; buf[i] = byte('0' + n%10); n /= 10 }
return string(buf[i:])
}