package repository

import (
"context"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgconn"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/product/internal/models"
)

type Repo struct {
db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

func IsUniqueViolation(err error) bool {
var pgErr *pgconn.PgError
return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// ---------- categories ----------

func (r *Repo) CreateCategory(ctx context.Context, name string, parentID *uuid.UUID) (*models.Category, error) {
c := &models.Category{}
err := r.db.QueryRow(ctx,
`INSERT INTO categories (name, parent_id) VALUES ($1, $2)
 RETURNING id, name, parent_id, created_at, updated_at`,
name, parentID,
).Scan(&c.ID, &c.Name, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
return c, err
}

func (r *Repo) ListCategories(ctx context.Context) ([]models.Category, error) {
rows, err := r.db.Query(ctx,
`SELECT id, name, parent_id, created_at, updated_at
 FROM categories ORDER BY name`)
if err != nil {
return nil, err
}
defer rows.Close()

out := []models.Category{}
for rows.Next() {
var c models.Category
if err := rows.Scan(&c.ID, &c.Name, &c.ParentID, &c.CreatedAt, &c.UpdatedAt); err != nil {
return nil, err
}
out = append(out, c)
}
return out, rows.Err()
}

func (r *Repo) GetCategory(ctx context.Context, id uuid.UUID) (*models.Category, error) {
c := &models.Category{}
err := r.db.QueryRow(ctx,
`SELECT id, name, parent_id, created_at, updated_at FROM categories WHERE id = $1`,
id,
).Scan(&c.ID, &c.Name, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return c, err
}

func (r *Repo) UpdateCategory(ctx context.Context, id uuid.UUID, name string, parentID *uuid.UUID) (*models.Category, error) {
c := &models.Category{}
err := r.db.QueryRow(ctx,
`UPDATE categories SET name = $2, parent_id = $3 WHERE id = $1
 RETURNING id, name, parent_id, created_at, updated_at`,
id, name, parentID,
).Scan(&c.ID, &c.Name, &c.ParentID, &c.CreatedAt, &c.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return c, err
}

func (r *Repo) DeleteCategory(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
if err != nil {
return false, err
}
return tag.RowsAffected() > 0, nil
}

// ---------- units ----------

func (r *Repo) ListUnits(ctx context.Context) ([]models.Unit, error) {
rows, err := r.db.Query(ctx,
`SELECT id, code, name, short_name, created_at FROM units ORDER BY name`)
if err != nil {
return nil, err
}
defer rows.Close()

out := []models.Unit{}
for rows.Next() {
var u models.Unit
if err := rows.Scan(&u.ID, &u.Code, &u.Name, &u.ShortName, &u.CreatedAt); err != nil {
return nil, err
}
out = append(out, u)
}
return out, rows.Err()
}

// ---------- products ----------

type ProductInput struct {
Name        string
SKU         *string
Barcode     *string
CategoryID  *uuid.UUID
UnitID      *uuid.UUID
Description *string
Price       float64
Currency    string
}

func (r *Repo) CreateProduct(ctx context.Context, in ProductInput) (*models.Product, error) {
p := &models.Product{}
err := r.db.QueryRow(ctx,
`INSERT INTO products (name, sku, barcode, category_id, unit_id, description, price, currency)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
 RETURNING id, name, sku, barcode, category_id, unit_id, description, price, currency, is_archived, created_at, updated_at`,
in.Name, in.SKU, in.Barcode, in.CategoryID, in.UnitID, in.Description, in.Price, in.Currency,
).Scan(&p.ID, &p.Name, &p.SKU, &p.Barcode, &p.CategoryID, &p.UnitID, &p.Description,
&p.Price, &p.Currency, &p.IsArchived, &p.CreatedAt, &p.UpdatedAt)
return p, err
}

func (r *Repo) GetProduct(ctx context.Context, id uuid.UUID) (*models.Product, error) {
p := &models.Product{}
err := r.db.QueryRow(ctx,
`SELECT id, name, sku, barcode, category_id, unit_id, description, price, currency, is_archived, created_at, updated_at
 FROM products WHERE id = $1`, id,
).Scan(&p.ID, &p.Name, &p.SKU, &p.Barcode, &p.CategoryID, &p.UnitID, &p.Description,
&p.Price, &p.Currency, &p.IsArchived, &p.CreatedAt, &p.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return p, err
}

func (r *Repo) ListProducts(ctx context.Context, includeArchived bool, categoryID *uuid.UUID) ([]models.Product, error) {
q := `SELECT id, name, sku, barcode, category_id, unit_id, description, price, currency, is_archived, created_at, updated_at
      FROM products WHERE 1=1`
args := []any{}
i := 1
if !includeArchived {
q += ` AND is_archived = FALSE`
}
if categoryID != nil {
q += ` AND category_id = $` + itoa(i)
args = append(args, *categoryID)
i++
}
q += ` ORDER BY name`

rows, err := r.db.Query(ctx, q, args...)
if err != nil {
return nil, err
}
defer rows.Close()

out := []models.Product{}
for rows.Next() {
var p models.Product
if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.Barcode, &p.CategoryID, &p.UnitID,
&p.Description, &p.Price, &p.Currency, &p.IsArchived, &p.CreatedAt, &p.UpdatedAt); err != nil {
return nil, err
}
out = append(out, p)
}
return out, rows.Err()
}

func (r *Repo) UpdateProduct(ctx context.Context, id uuid.UUID, in ProductInput) (*models.Product, error) {
p := &models.Product{}
err := r.db.QueryRow(ctx,
`UPDATE products SET name=$2, sku=$3, barcode=$4, category_id=$5, unit_id=$6,
                     description=$7, price=$8, currency=$9
 WHERE id=$1
 RETURNING id, name, sku, barcode, category_id, unit_id, description, price, currency, is_archived, created_at, updated_at`,
id, in.Name, in.SKU, in.Barcode, in.CategoryID, in.UnitID, in.Description, in.Price, in.Currency,
).Scan(&p.ID, &p.Name, &p.SKU, &p.Barcode, &p.CategoryID, &p.UnitID, &p.Description,
&p.Price, &p.Currency, &p.IsArchived, &p.CreatedAt, &p.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
return p, err
}

func (r *Repo) ArchiveProduct(ctx context.Context, id uuid.UUID) (bool, error) {
tag, err := r.db.Exec(ctx, `UPDATE products SET is_archived = TRUE WHERE id = $1`, id)
if err != nil {
return false, err
}
return tag.RowsAffected() > 0, nil
}

func itoa(n int) string {
if n == 0 {
return "0"
}
var buf [20]byte
i := len(buf)
for n > 0 {
i--
buf[i] = byte('0' + n%10)
n /= 10
}
return string(buf[i:])
}