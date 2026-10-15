package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/radonezhsklad/warehouse/internal/models"
)

type InternalOrderIssueRepo struct{ db *pgxpool.Pool }

func NewInternalOrderIssueRepo(db *pgxpool.Pool) *InternalOrderIssueRepo {
	return &InternalOrderIssueRepo{db: db}
}

type InternalOrderIssueInput struct {
	ItemID    uuid.UUID
	Qty       float64
	IssueDate *string // YYYY-MM-DD, nil = today
	IssuedBy  *uuid.UUID
	Comment   *string
}

var ErrIssueOverflow = errors.New("issue exceeds remaining quantity")

// AddIssue вставляет одну выдачу в транзакции:
//   - проверяет остаток по позиции (с блокировкой строки)
//   - пишет internal_order_issues
//   - увеличивает issued_qty у позиции
//   - пересчитывает issue_status у заказа
func (r *InternalOrderIssueRepo) AddIssue(ctx context.Context, orderID uuid.UUID, in InternalOrderIssueInput) (*models.InternalOrderIssue, error) {
	var out *models.InternalOrderIssue
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var orderedQty, issuedQty float64
		err := tx.QueryRow(ctx, `
SELECT quantity, issued_qty FROM internal_order_items
WHERE id = $1 AND order_id = $2
FOR UPDATE`, in.ItemID, orderID).Scan(&orderedQty, &issuedQty)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("item not found in order")
		}
		if err != nil {
			return err
		}
		if in.Qty <= 0 {
			return fmt.Errorf("qty must be > 0")
		}
		if issuedQty+in.Qty > orderedQty+1e-9 {
			return ErrIssueOverflow
		}

		issueDate := time.Now()
		if in.IssueDate != nil && *in.IssueDate != "" {
			if t, perr := time.Parse("2006-01-02", *in.IssueDate); perr == nil {
				issueDate = t
			}
		}

		iss := &models.InternalOrderIssue{}
		if err := tx.QueryRow(ctx, `
INSERT INTO internal_order_issues (order_id, item_id, qty, issue_date, issued_by, comment)
VALUES ($1,$2,$3,$4,$5,$6)
RETURNING id, order_id, item_id, qty, issue_date, issued_by, comment, created_at`,
			orderID, in.ItemID, in.Qty, issueDate, in.IssuedBy, in.Comment,
		).Scan(
			&iss.ID, &iss.OrderID, &iss.ItemID, &iss.Qty, &iss.IssueDate,
			&iss.IssuedBy, &iss.Comment, &iss.CreatedAt,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
UPDATE internal_order_items SET issued_qty = issued_qty + $2 WHERE id = $1`, in.ItemID, in.Qty); err != nil {
			return err
		}
		if err := recalcIssueStatus(ctx, tx, orderID); err != nil {
			return err
		}
		out = iss
		return nil
	})
	return out, err
}

func (r *InternalOrderIssueRepo) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]models.InternalOrderIssue, error) {
	rows, err := r.db.Query(ctx, `
SELECT id, order_id, item_id, qty, issue_date, issued_by, comment, created_at
FROM internal_order_issues WHERE order_id = $1
ORDER BY created_at DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.InternalOrderIssue{}
	for rows.Next() {
		var iss models.InternalOrderIssue
		if err := rows.Scan(&iss.ID, &iss.OrderID, &iss.ItemID, &iss.Qty, &iss.IssueDate,
			&iss.IssuedBy, &iss.Comment, &iss.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, iss)
	}
	return out, rows.Err()
}

// recalcIssueStatus выставляет none | partial | full
func recalcIssueStatus(ctx context.Context, tx pgx.Tx, orderID uuid.UUID) error {
	var total, fullCount, partialCount int
	err := tx.QueryRow(ctx, `
SELECT
  COUNT(*),
  COUNT(*) FILTER (WHERE issued_qty >= quantity - 1e-9),
  COUNT(*) FILTER (WHERE issued_qty > 0)
FROM internal_order_items WHERE order_id = $1`, orderID).Scan(&total, &fullCount, &partialCount)
	if err != nil {
		return err
	}
	status := "none"
	switch {
	case total > 0 && fullCount == total:
		status = "full"
	case partialCount > 0:
		status = "partial"
	}
	_, err = tx.Exec(ctx, `UPDATE internal_orders SET issue_status = $2 WHERE id = $1`, orderID, status)
	return err
}
