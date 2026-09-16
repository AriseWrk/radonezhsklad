package repository

import (
"context"
"errors"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/order/internal/models"
)

type ContractRepo struct{ db *pgxpool.Pool }

func NewContractRepo(db *pgxpool.Pool) *ContractRepo { return &ContractRepo{db: db} }

const contractCols = `id, number, contract_type, code, doc_date, customer_id, organization_id,
                      amount, currency, paid, fulfilled, comment,
                      printed_at, sent_at, archived, created_at, updated_at`

func scanContract(row pgx.Row) (*models.Contract, error) {
c := &models.Contract{}
err := row.Scan(&c.ID, &c.Number, &c.ContractType, &c.Code, &c.DocDate, &c.CustomerID, &c.OrganizationID,
&c.Amount, &c.Currency, &c.Paid, &c.Fulfilled, &c.Comment,
&c.PrintedAt, &c.SentAt, &c.Archived, &c.CreatedAt, &c.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return c, err
}

func (r *ContractRepo) Create(ctx context.Context, c *models.Contract) error {
if c.Currency == "" { c.Currency = "RUB" }
if c.ContractType == "" { c.ContractType = "Договор купли-продажи" }
if c.DocDate.IsZero() { c.DocDate = time.Now() }
return r.db.QueryRow(ctx,
`INSERT INTO contracts
 (number, contract_type, code, doc_date, customer_id, organization_id,
  amount, currency, paid, fulfilled, comment, archived)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
 RETURNING id, created_at, updated_at`,
c.Number, c.ContractType, c.Code, c.DocDate, c.CustomerID, c.OrganizationID,
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
 number=$2, contract_type=$3, code=$4, doc_date=$5, customer_id=$6, organization_id=$7,
 amount=$8, currency=$9, paid=$10, fulfilled=$11, comment=$12, archived=$13
 WHERE id=$1`,
c.ID, c.Number, c.ContractType, c.Code, c.DocDate, c.CustomerID, c.OrganizationID,
c.Amount, c.Currency, c.Paid, c.Fulfilled, c.Comment, c.Archived)
return err
}

func (r *ContractRepo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `DELETE FROM contracts WHERE id = $1`, id)
if err != nil { return false, err }
return tag.RowsAffected() > 0, nil
}

// NeighborsByID — для кнопок ◀ ▶ (соседние договоры в порядке doc_date DESC).
type ContractNeighbor struct {
ID      uuid.UUID
Number  string
Index   int
Total   int
}

func (r *ContractRepo) Neighbors(ctx context.Context, id uuid.UUID) (*ContractNeighbor, error) {
// Получаем общее число активных договоров
var total int
if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contracts WHERE archived = FALSE`).Scan(&total); err != nil {
return nil, err
}

// Находим позицию текущего
rows, err := r.db.Query(ctx,
`SELECT id FROM contracts WHERE archived = FALSE ORDER BY doc_date DESC, created_at DESC`)
if err != nil { return nil, err }
defer rows.Close()

idx := 0
found := false
for rows.Next() {
var cid uuid.UUID
if err := rows.Scan(&cid); err != nil { return nil, err }
idx++
if cid == id { found = true; break }
}
if !found { idx = 1 }

return &ContractNeighbor{ID: id, Index: idx, Total: total}, nil
}

func (r *ContractRepo) GetNext(ctx context.Context, currentDocDate time.Time, currentID uuid.UUID) (*models.Contract, error) {
// Следующий = более ранний по дате (мы идём от свежих к старым)
c, err := scanContract(r.db.QueryRow(ctx,
`SELECT `+contractCols+` FROM contracts
 WHERE archived = FALSE AND (doc_date, id) < ($1, $2)
 ORDER BY doc_date DESC, id DESC LIMIT 1`,
currentDocDate, currentID))
if err != nil { return nil, err }
return c, nil
}

func (r *ContractRepo) GetPrev(ctx context.Context, currentDocDate time.Time, currentID uuid.UUID) (*models.Contract, error) {
c, err := scanContract(r.db.QueryRow(ctx,
`SELECT `+contractCols+` FROM contracts
 WHERE archived = FALSE AND (doc_date, id) > ($1, $2)
 ORDER BY doc_date ASC, id ASC LIMIT 1`,
currentDocDate, currentID))
if err != nil { return nil, err }
return c, nil
}