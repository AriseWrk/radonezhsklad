package repository

import (
"context"
"errors"
"strings"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/auth/internal/models"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

const selectCols = `id, email, password_hash, full_name, last_name, first_name, middle_name,
                    phone, login, description, role, is_active, created_at, updated_at`

func scanUser(row pgx.Row) (*models.User, error) {
u := &models.User{}
err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName,
&u.LastName, &u.FirstName, &u.MiddleName,
&u.Phone, &u.Login, &u.Description,
&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
return u, err
}

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
return r.db.QueryRow(ctx,
`INSERT INTO users (email, password_hash, full_name, last_name, first_name, middle_name,
                    phone, login, description, role)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
 RETURNING id, is_active, created_at, updated_at`,
u.Email, u.PasswordHash, u.FullName, u.LastName, u.FirstName, u.MiddleName,
u.Phone, u.Login, u.Description, u.Role,
).Scan(&u.ID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
return scanUser(r.db.QueryRow(ctx, `SELECT `+selectCols+` FROM users WHERE email = $1`, email))
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
return scanUser(r.db.QueryRow(ctx, `SELECT `+selectCols+` FROM users WHERE id = $1`, id))
}

func (r *UserRepo) List(ctx context.Context) ([]models.User, error) {
rows, err := r.db.Query(ctx, `SELECT `+selectCols+` FROM users ORDER BY last_name, first_name`)
if err != nil { return nil, err }
defer rows.Close()
out := []models.User{}
for rows.Next() {
u, err := scanUser(rows)
if err != nil { return nil, err }
out = append(out, *u)
}
return out, rows.Err()
}

func (r *UserRepo) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
_, err := r.db.Exec(ctx, `UPDATE users SET role = $2 WHERE id = $1`, id, role)
return err
}

func (r *UserRepo) UpdateActive(ctx context.Context, id uuid.UUID, active bool) error {
_, err := r.db.Exec(ctx, `UPDATE users SET is_active = $2 WHERE id = $1`, id, active)
return err
}

func (r *UserRepo) Update(ctx context.Context, u *models.User) error {
    fullName := strings.TrimSpace(u.LastName + " " + u.FirstName + " " + u.MiddleName)
    _, err := r.db.Exec(ctx,
        `UPDATE users SET email=$2, password_hash=$3,
                  last_name=$4, first_name=$5, middle_name=$6,
                  phone=$7, login=$8, description=$9, role=$10, is_active=$11,
                  full_name=$12, updated_at=NOW()
         WHERE id=$1`,
        u.ID, u.Email, u.PasswordHash,
        u.LastName, u.FirstName, u.MiddleName,
        u.Phone, u.Login, u.Description, u.Role, u.IsActive,
        fullName,
    )
    return err
}
