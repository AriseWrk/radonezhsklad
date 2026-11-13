package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderBrief — минимальная карточка ВЗ для списка внутри проекта.
type OrderBrief struct {
	ID            uuid.UUID `json:"id"`
	Number        string    `json:"number"`
	DocDate       time.Time `json:"doc_date"`
	Status        string    `json:"status"`
	Total         float64   `json:"total"`
	ShippedAmount float64   `json:"shipped_amount"`
	ItemCount     int       `json:"item_count"`
	Comment       *string   `json:"comment,omitempty"`
}

// SystemItem — агрегированная позиция по системе внутри проекта.
type SystemItem struct {
	SystemCode  string     `json:"system_code"`
	ProductID   uuid.UUID  `json:"product_id"`
	Quantity    float64    `json:"quantity"`
	OrderCount  int        `json:"order_count"`
	LastDocDate *time.Time `json:"last_doc_date,omitempty"`
}

type ProjectStatsRepo struct {
	db *pgxpool.Pool
}

func NewProjectStatsRepo(db *pgxpool.Pool) *ProjectStatsRepo {
	return &ProjectStatsRepo{db: db}
}

// ListOrdersByProject — все ВЗ проекта (кроме удалённых), сортировка по дате ВЗ desc.
func (r *ProjectStatsRepo) ListOrdersByProject(ctx context.Context, projectID uuid.UUID) ([]OrderBrief, error) {
	const q = "SELECT o.id, o.number, o.doc_date, o.status, o.total, o.shipped_amount, " +
		"COALESCE((SELECT COUNT(*) FROM internal_order_items i WHERE i.order_id = o.id), 0) AS item_count, " +
		"o.comment " +
		"FROM internal_orders o " +
		"WHERE o.project = $1 " +
		"AND o.deleted_at IS NULL " +
		"ORDER BY o.doc_date DESC, o.number DESC"
	rows, err := r.db.Query(ctx, q, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]OrderBrief, 0)
	for rows.Next() {
		var o OrderBrief
		if err := rows.Scan(&o.ID, &o.Number, &o.DocDate, &o.Status, &o.Total, &o.ShippedAmount, &o.ItemCount, &o.Comment); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// SystemsSummaryByProject — агрегация позиций по системам внутри проекта.
// system_code может быть NULL (позиции без системы) — они попадают в бакет system_code=”.
func (r *ProjectStatsRepo) SystemsSummaryByProject(ctx context.Context, projectID uuid.UUID) ([]SystemItem, error) {
	const q = "SELECT COALESCE(i.system_code, '') AS system_code, " +
		"i.product_id, " +
		"SUM(i.quantity)::float8 AS quantity, " +
		"COUNT(DISTINCT i.order_id)::int AS order_count, " +
		"MAX(o.doc_date) AS last_doc_date " +
		"FROM internal_order_items i " +
		"JOIN internal_orders o ON o.id = i.order_id " +
		"WHERE o.project = $1 " +
		"AND o.deleted_at IS NULL " +
		"GROUP BY COALESCE(i.system_code, ''), i.product_id " +
		"ORDER BY system_code, quantity DESC"
	rows, err := r.db.Query(ctx, q, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SystemItem, 0)
	for rows.Next() {
		var it SystemItem
		if err := rows.Scan(&it.SystemCode, &it.ProductID, &it.Quantity, &it.OrderCount, &it.LastDocDate); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}
