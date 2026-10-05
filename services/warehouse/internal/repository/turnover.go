package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TurnoverRow struct {
	ProductID uuid.UUID
	Opening   float64
	Income    float64
	Outcome   float64
	Closing   float64
}

func (r *Repo) TurnoverReport(ctx context.Context, warehouseID *uuid.UUID, from, to time.Time) ([]TurnoverRow, error) {
	rows, err := r.db.Query(ctx, `
WITH agg AS (
    SELECT
        product_id,
        COALESCE(SUM(CASE WHEN created_at < $2::timestamptz THEN quantity_delta ELSE 0 END), 0) AS opening,
        COALESCE(SUM(CASE WHEN created_at >= $2::timestamptz AND created_at <= $3::timestamptz AND quantity_delta > 0 THEN quantity_delta ELSE 0 END), 0) AS income,
        COALESCE(SUM(CASE WHEN created_at >= $2::timestamptz AND created_at <= $3::timestamptz AND quantity_delta < 0 THEN -quantity_delta ELSE 0 END), 0) AS outcome
    FROM stock_movements
    WHERE ($1::uuid IS NULL OR warehouse_id = $1::uuid)
      AND created_at <= $3::timestamptz
    GROUP BY product_id
)
SELECT product_id, opening, income, outcome, opening + income - outcome AS closing
FROM agg
WHERE opening <> 0 OR income <> 0 OR outcome <> 0
ORDER BY product_id`, warehouseID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []TurnoverRow{}
	for rows.Next() {
		var row TurnoverRow
		if err := rows.Scan(&row.ProductID, &row.Opening, &row.Income, &row.Outcome, &row.Closing); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
