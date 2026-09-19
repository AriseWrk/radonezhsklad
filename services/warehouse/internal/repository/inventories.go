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

type InventoryRepo struct{ db *pgxpool.Pool }

func NewInventoryRepo(db *pgxpool.Pool) *InventoryRepo { return &InventoryRepo{db: db} }

type InventoryFilters struct {
    WarehouseID *uuid.UUID
    From        *time.Time
    To          *time.Time
    Limit       int
}

const inventorySelect = `
SELECT i.id, i.external_id, i.number, i.doc_date, i.warehouse_id,
       COALESCE(w.name, ''), i.organization_id, i.comment, i.total,
       i.created_at, i.updated_at,
       (SELECT COUNT(*) FROM inventory_items it WHERE it.inventory_id = i.id) AS items_count
FROM inventories i
LEFT JOIN warehouses w ON w.id = i.warehouse_id`

func scanInventory(row pgx.Row) (*models.Inventory, error) {
    inv := &models.Inventory{}
    err := row.Scan(&inv.ID, &inv.ExternalID, &inv.Number, &inv.DocDate, &inv.WarehouseID,
        &inv.WarehouseName, &inv.OrganizationID, &inv.Comment, &inv.Total,
        &inv.CreatedAt, &inv.UpdatedAt, &inv.ItemsCount)
    if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
    return inv, err
}

func (r *InventoryRepo) List(ctx context.Context, f InventoryFilters) ([]models.Inventory, error) {
    q := inventorySelect + ` WHERE 1=1`
    args := []any{}
    i := 1
    if f.WarehouseID != nil {
        q += ` AND i.warehouse_id = $` + itoa(i)
        args = append(args, *f.WarehouseID)
        i++
    }
    if f.From != nil {
        q += ` AND i.doc_date >= $` + itoa(i)
        args = append(args, *f.From)
        i++
    }
    if f.To != nil {
        q += ` AND i.doc_date <= $` + itoa(i)
        args = append(args, *f.To)
        i++
    }
    lim := f.Limit
    if lim <= 0 { lim = 500 }
    q += ` ORDER BY i.doc_date DESC LIMIT ` + itoa(lim)

    rows, err := r.db.Query(ctx, q, args...)
    if err != nil { return nil, err }
    defer rows.Close()
    out := []models.Inventory{}
    for rows.Next() {
        inv, err := scanInventory(rows)
        if err != nil { return nil, err }
        if inv != nil { out = append(out, *inv) }
    }
    return out, rows.Err()
}

func (r *InventoryRepo) Get(ctx context.Context, id uuid.UUID) (*models.Inventory, error) {
    inv, err := scanInventory(r.db.QueryRow(ctx, inventorySelect+` WHERE i.id = $1`, id))
    if err != nil || inv == nil { return inv, err }

    rows, err := r.db.Query(ctx,
        `SELECT id, inventory_id, product_id, quantity, calculated_quantity,
                correction_amount, price, correction_sum, created_at
         FROM inventory_items WHERE inventory_id = $1`, id)
    if err != nil { return nil, err }
    defer rows.Close()
    items := []models.InventoryItem{}
    for rows.Next() {
        var it models.InventoryItem
        if err := rows.Scan(&it.ID, &it.InventoryID, &it.ProductID,
            &it.Quantity, &it.CalculatedQuantity, &it.CorrectionAmount,
            &it.Price, &it.CorrectionSum, &it.CreatedAt); err != nil {
            return nil, err
        }
        items = append(items, it)
    }
    inv.Items = items
    return inv, rows.Err()
}
