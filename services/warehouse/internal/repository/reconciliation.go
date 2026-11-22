package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReconciliationRun struct {
	ID                   uuid.UUID  `json:"id"`
	Status               string     `json:"status"`
	StartedAt            time.Time  `json:"started_at"`
	FinishedAt           *time.Time `json:"finished_at,omitempty"`
	TriggeredBy          *uuid.UUID `json:"triggered_by,omitempty"`
	TriggeredByName      string     `json:"triggered_by_name"`
	PriceTypeName        string     `json:"price_type_name"`
	Error                *string    `json:"error,omitempty"`
	TotalProducts        int        `json:"total_products"`
	MatchedProducts      int        `json:"matched_products"`
	PriceDifferences     int        `json:"price_differences"`
	StockDifferences     int        `json:"stock_differences"`
	MissingLocalProducts int        `json:"missing_local_products"`
	MissingMSProducts    int        `json:"missing_ms_products"`
	AmbiguousProducts    int        `json:"ambiguous_products"`
	UnmatchedWarehouses  int        `json:"unmatched_warehouses"`
	ZeroStockItems       int        `json:"zero_stock_items"`
}

type ReconciliationItem struct {
	ID                 uuid.UUID       `json:"id"`
	RunID              uuid.UUID       `json:"run_id"`
	ItemType           string          `json:"item_type"`
	ResultStatus       string          `json:"result_status"`
	MatchMethod        *string         `json:"match_method,omitempty"`
	LocalProductID     *uuid.UUID      `json:"local_product_id,omitempty"`
	MSProductID        *string         `json:"ms_product_id,omitempty"`
	SKU                *string         `json:"sku,omitempty"`
	LocalProductName   *string         `json:"local_product_name,omitempty"`
	MSProductName      *string         `json:"ms_product_name,omitempty"`
	LocalWarehouseID   *uuid.UUID      `json:"local_warehouse_id,omitempty"`
	MSStoreID          *string         `json:"ms_store_id,omitempty"`
	LocalWarehouseName *string         `json:"local_warehouse_name,omitempty"`
	MSStoreName        *string         `json:"ms_store_name,omitempty"`
	LocalPrice         *float64        `json:"local_price,omitempty"`
	MSPrice            *float64        `json:"ms_price,omitempty"`
	PriceDifference    *float64        `json:"price_difference,omitempty"`
	LocalStock         *float64        `json:"local_stock,omitempty"`
	MSStock            *float64        `json:"ms_stock,omitempty"`
	StockDifference    *float64        `json:"stock_difference,omitempty"`
	IsZeroStock        bool            `json:"is_zero_stock"`
	Details            json.RawMessage `json:"details"`
	CreatedAt          time.Time       `json:"created_at"`
}

type ReconciliationCounters struct {
	TotalProducts        int
	MatchedProducts      int
	PriceDifferences     int
	StockDifferences     int
	MissingLocalProducts int
	MissingMSProducts    int
	AmbiguousProducts    int
	UnmatchedWarehouses  int
	ZeroStockItems       int
}

type ReconciliationRepo struct {
	db *pgxpool.Pool
}

func NewReconciliationRepo(db *pgxpool.Pool) *ReconciliationRepo {
	return &ReconciliationRepo{db: db}
}

const reconciliationRunSelect = `
SELECT id, status, started_at, finished_at, triggered_by,
       COALESCE(triggered_by_name, ''), price_type_name, error,
       total_products, matched_products, price_differences,
       stock_differences, missing_local_products, missing_ms_products,
       ambiguous_products, unmatched_warehouses, zero_stock_items
FROM reconciliation_runs`

func scanReconciliationRun(row pgx.Row) (*ReconciliationRun, error) {
	var r ReconciliationRun
	err := row.Scan(
		&r.ID, &r.Status, &r.StartedAt, &r.FinishedAt, &r.TriggeredBy,
		&r.TriggeredByName, &r.PriceTypeName, &r.Error,
		&r.TotalProducts, &r.MatchedProducts, &r.PriceDifferences,
		&r.StockDifferences, &r.MissingLocalProducts, &r.MissingMSProducts,
		&r.AmbiguousProducts, &r.UnmatchedWarehouses, &r.ZeroStockItems,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (r *ReconciliationRepo) CreateRun(ctx context.Context, userID uuid.UUID, userName string) (*ReconciliationRun, error) {
	id := uuid.New()
	_, err := r.db.Exec(ctx, `
INSERT INTO reconciliation_runs (id, status, triggered_by, triggered_by_name, price_type_name)
VALUES ($1, 'running', NULLIF($2::uuid, '00000000-0000-0000-0000-000000000000'), $3, 'Цена продажи')`,
		id, userID, userName)
	if err != nil {
		return nil, err
	}
	return r.GetRun(ctx, id)
}

func (r *ReconciliationRepo) GetRun(ctx context.Context, id uuid.UUID) (*ReconciliationRun, error) {
	run, err := scanReconciliationRun(r.db.QueryRow(ctx, reconciliationRunSelect+` WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return run, err
}

func (r *ReconciliationRepo) GetActiveRun(ctx context.Context) (*ReconciliationRun, error) {
	run, err := scanReconciliationRun(r.db.QueryRow(ctx,
		reconciliationRunSelect+` WHERE status = 'running' ORDER BY started_at DESC LIMIT 1`))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return run, err
}

func (r *ReconciliationRepo) ListRuns(ctx context.Context, limit int) ([]*ReconciliationRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.Query(ctx,
		reconciliationRunSelect+` ORDER BY started_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*ReconciliationRun, 0, limit)
	for rows.Next() {
		run, err := scanReconciliationRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

func (r *ReconciliationRepo) InsertItems(ctx context.Context, runID uuid.UUID, items []ReconciliationItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range items {
		item := &items[i]
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.RunID = runID
		if len(item.Details) == 0 {
			item.Details = json.RawMessage(`{}`)
		}
		if !json.Valid(item.Details) {
			return fmt.Errorf("invalid details JSON for reconciliation item %s", item.ID)
		}

		_, err := tx.Exec(ctx, `
INSERT INTO reconciliation_items (
id, run_id, item_type, result_status, match_method,
local_product_id, ms_product_id, sku, local_product_name, ms_product_name,
local_warehouse_id, ms_store_id, local_warehouse_name, ms_store_name,
local_price, ms_price, price_difference,
local_stock, ms_stock, stock_difference, is_zero_stock, details
) VALUES (
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22::jsonb
)`,
			item.ID, item.RunID, item.ItemType, item.ResultStatus, item.MatchMethod,
			item.LocalProductID, item.MSProductID, item.SKU, item.LocalProductName, item.MSProductName,
			item.LocalWarehouseID, item.MSStoreID, item.LocalWarehouseName, item.MSStoreName,
			item.LocalPrice, item.MSPrice, item.PriceDifference,
			item.LocalStock, item.MSStock, item.StockDifference, item.IsZeroStock, string(item.Details),
		)
		if err != nil {
			return fmt.Errorf("insert reconciliation item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *ReconciliationRepo) FinishRun(ctx context.Context, id uuid.UUID, status, errorMessage string, counters ReconciliationCounters) error {
	if status != "completed" && status != "failed" {
		return fmt.Errorf("invalid reconciliation status %q", status)
	}

	var errValue *string
	if strings.TrimSpace(errorMessage) != "" {
		errValue = &errorMessage
	}

	_, err := r.db.Exec(ctx, `
UPDATE reconciliation_runs
SET status = $2,
    error = $3,
    finished_at = NOW(),
    total_products = $4,
    matched_products = $5,
    price_differences = $6,
    stock_differences = $7,
    missing_local_products = $8,
    missing_ms_products = $9,
    ambiguous_products = $10,
    unmatched_warehouses = $11,
    zero_stock_items = $12
WHERE id = $1`,
		id, status, errValue,
		counters.TotalProducts, counters.MatchedProducts,
		counters.PriceDifferences, counters.StockDifferences,
		counters.MissingLocalProducts, counters.MissingMSProducts,
		counters.AmbiguousProducts, counters.UnmatchedWarehouses,
		counters.ZeroStockItems,
	)
	return err
}

func (r *ReconciliationRepo) ListItems(ctx context.Context, runID uuid.UUID, itemType, resultStatus, search string, limit, offset int) ([]ReconciliationItem, int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	where := []string{"run_id = $1"}
	args := []any{runID}
	add := func(condition string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(condition, len(args)))
	}

	if itemType != "" {
		add("item_type = $%d", itemType)
	}
	if resultStatus != "" {
		add("result_status = $%d", resultStatus)
	}
	if search != "" {
		add("(COALESCE(sku,'') ILIKE $%d OR COALESCE(local_product_name,'') ILIKE $%d OR COALESCE(ms_product_name,'') ILIKE $%d OR COALESCE(local_warehouse_name,'') ILIKE $%d OR COALESCE(ms_store_name,'') ILIKE $%d)",
			"%"+search+"%")
		// Для поискового условия один параметр используется во всех пяти позициях.
		p := len(args)
		where[len(where)-1] = fmt.Sprintf(
			"(COALESCE(sku,'') ILIKE $%d OR COALESCE(local_product_name,'') ILIKE $%d OR COALESCE(ms_product_name,'') ILIKE $%d OR COALESCE(local_warehouse_name,'') ILIKE $%d OR COALESCE(ms_store_name,'') ILIKE $%d)",
			p, p, p, p, p,
		)
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM reconciliation_items WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	query := `
SELECT id, run_id, item_type, result_status, match_method,
       local_product_id, ms_product_id, sku, local_product_name, ms_product_name,
       local_warehouse_id, ms_store_id, local_warehouse_name, ms_store_name,
       local_price, ms_price, price_difference,
       local_stock, ms_stock, stock_difference, is_zero_stock, details, created_at
FROM reconciliation_items
WHERE ` + whereSQL + fmt.Sprintf(" ORDER BY created_at, id LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]ReconciliationItem, 0, limit)
	for rows.Next() {
		var item ReconciliationItem
		if err := rows.Scan(
			&item.ID, &item.RunID, &item.ItemType, &item.ResultStatus, &item.MatchMethod,
			&item.LocalProductID, &item.MSProductID, &item.SKU, &item.LocalProductName, &item.MSProductName,
			&item.LocalWarehouseID, &item.MSStoreID, &item.LocalWarehouseName, &item.MSStoreName,
			&item.LocalPrice, &item.MSPrice, &item.PriceDifference,
			&item.LocalStock, &item.MSStock, &item.StockDifference, &item.IsZeroStock,
			&item.Details, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
