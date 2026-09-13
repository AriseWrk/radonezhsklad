package repository

import (
	"time"
"context"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/order/internal/models"
)

type ContractRepo struct{ db *pgxpool.Pool }

func NewContractRepo(db *pgxpool.Pool) *ContractRepo { return &ContractRepo{db: db} }

const contractCols = `id, number, code, doc_date, customer_id, organization_id,
                      amount, currency, paid, fulfilled, comment,
                      printed_at, sent_at, archived, created_at, updated_at`

func scanContract(row pgx.Row) (*models.Contract, error) {
c := &models.Contract{}
err := row.Scan(&c.ID, &c.Number, &c.Code, &c.DocDate, &c.CustomerID, &c.OrganizationID,
&c.Amount, &c.Currency, &c.Paid, &c.Fulfilled, &c.Comment,
&c.PrintedAt, &c.SentAt, &c.Archived, &c.CreatedAt, &c.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return c, err
}

func (r *ContractRepo) Create(ctx context.Context, c *models.Contract) error {
if c.Currency == "" { c.Currency = "RUB" }
if c.DocDate.IsZero() { c.DocDate = time.Now() }
return r.db.QueryRow(ctx,
`INSERT INTO contracts
 (number, code, doc_date, customer_id, organization_id,
  amount, currency, paid, fulfilled, comment, archived)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
 RETURNING id, created_at, updated_at`,
c.Number, c.Code, c.DocDate, c.CustomerID, c.OrganizationID,
c.Amount, c.Currency, c.Paid, c.Fulfilled, c.Comment, c.Archived,
).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *ContractRepo) List(ctx context.Context, includeArchived bool) ([]models.Contract, error) {
q := `SELECT ` + contractCols + ` FROM contracts`
if !includeArchived { q += ` WHERE archived = FALSE` }
q += ` ORDER BY doc_date DESC, created_at DESC LIMIT 5000`

rows, err := r.db.Query(ctx, q)
if err != nil { return nil, err }
defer rows.Close()

out := []models.Contract{}
for rows.Next() {
c, err := scanContract(rows)
if err != nil { return nil, err }
out = append(out, *c)
}
return out, rows.Err()
}

func (r *ContractRepo) Get(ctx context.Context, id uuid.UUID) (*models.Contract, error) {
return scanContract(r.db.QueryRow(ctx, `SELECT `+contractCols+` FROM contracts WHERE id = $1`, id))
}

func (r *ContractRepo) Update(ctx context.Context, c *models.Contract) error {
_, err := r.db.Exec(ctx,
`UPDATE contracts SET
 number=$2, code=$3, doc_date=$4, customer_id=$5, organization_id=$6,
 amount=$7, currency=$8, paid=$9, fulfilled=$10, comment=$11, archived=$12
 WHERE id=$1`,
c.ID, c.Number, c.Code, c.DocDate, c.CustomerID, c.OrganizationID,
c.Amount, c.Currency, c.Paid, c.Fulfilled, c.Comment, c.Archived)
return err
}

func (r *ContractRepo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `DELETE FROM contracts WHERE id = $1`, id)
if err != nil { return false, err }
return tag.RowsAffected() > 0, nil
}