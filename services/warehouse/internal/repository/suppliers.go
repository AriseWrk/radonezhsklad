package repository

import (
"context"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/warehouse/internal/models"
)

type SupplierRepo struct{ db *pgxpool.Pool }

func NewSupplierRepo(db *pgxpool.Pool) *SupplierRepo { return &SupplierRepo{db: db} }

func (r *SupplierRepo) Create(ctx context.Context, s *models.Supplier) error {
return r.db.QueryRow(ctx,
`INSERT INTO suppliers (name, inn, phone, email, address)
 VALUES ($1,$2,$3,$4,$5)
 RETURNING id, created_at, updated_at`,
s.Name, s.INN, s.Phone, s.Email, s.Address,
).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *SupplierRepo) Get(ctx context.Context, id uuid.UUID) (*models.Supplier, error) {
s := &models.Supplier{}
err := r.db.QueryRow(ctx,
`SELECT id, name, inn, phone, email, address, created_at, updated_at
 FROM suppliers WHERE id = $1`, id,
).Scan(&s.ID, &s.Name, &s.INN, &s.Phone, &s.Email, &s.Address, &s.CreatedAt, &s.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return s, err
}

func (r *SupplierRepo) List(ctx context.Context) ([]models.Supplier, error) {
rows, err := r.db.Query(ctx,
`SELECT id, name, inn, phone, email, address, created_at, updated_at
 FROM suppliers ORDER BY name`)
if err != nil { return nil, err }
defer rows.Close()
out := []models.Supplier{}
for rows.Next() {
var s models.Supplier
if err := rows.Scan(&s.ID, &s.Name, &s.INN, &s.Phone, &s.Email, &s.Address, &s.CreatedAt, &s.UpdatedAt); err != nil {
return nil, err
}
out = append(out, s)
}
return out, rows.Err()
}

func (r *SupplierRepo) Update(ctx context.Context, s *models.Supplier) error {
_, err := r.db.Exec(ctx,
`UPDATE suppliers SET name=$2, inn=$3, phone=$4, email=$5, address=$6 WHERE id=$1`,
s.ID, s.Name, s.INN, s.Phone, s.Email, s.Address,
)
return err
}

func (r *SupplierRepo) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `DELETE FROM suppliers WHERE id = $1`, id)
if err != nil { return false, err }
return tag.RowsAffected() > 0, nil
}

// ---------- organizations ----------

type OrganizationRepo struct{ db *pgxpool.Pool }

func NewOrganizationRepo(db *pgxpool.Pool) *OrganizationRepo { return &OrganizationRepo{db: db} }

func (r *OrganizationRepo) List(ctx context.Context) ([]models.Organization, error) {
rows, err := r.db.Query(ctx,
`SELECT id, name, inn, is_default, created_at FROM organizations ORDER BY name`)
if err != nil { return nil, err }
defer rows.Close()
out := []models.Organization{}
for rows.Next() {
var o models.Organization
if err := rows.Scan(&o.ID, &o.Name, &o.INN, &o.IsDefault, &o.CreatedAt); err != nil { return nil, err }
out = append(out, o)
}
return out, rows.Err()
}

func (r *OrganizationRepo) Default(ctx context.Context) (*models.Organization, error) {
o := &models.Organization{}
err := r.db.QueryRow(ctx,
`SELECT id, name, inn, is_default, created_at FROM organizations
 WHERE is_default = TRUE ORDER BY created_at LIMIT 1`,
).Scan(&o.ID, &o.Name, &o.INN, &o.IsDefault, &o.CreatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return o, err
}
func (r *OrganizationRepo) Get(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
o := &models.Organization{}
err := r.db.QueryRow(ctx,
`SELECT id, name, inn, is_default, created_at FROM organizations WHERE id = $1`, id,
).Scan(&o.ID, &o.Name, &o.INN, &o.IsDefault, &o.CreatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return o, err
}