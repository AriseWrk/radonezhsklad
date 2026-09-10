package repository

import (
"context"
"errors"

"github.com/google/uuid"
"github.com/jackc/pgx/v5"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/auth/internal/models"
)

type UserRepo struct {
db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
return r.db.QueryRow(ctx,
`INSERT INTO users (email, password_hash, full_name, role)
 VALUES ($1, $2, $3, $4)
 RETURNING id, is_active, created_at, updated_at`,
u.Email, u.PasswordHash, u.FullName, u.Role,
).Scan(&u.ID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
u := &models.User{}
err := r.db.QueryRow(ctx,
`SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
 FROM users WHERE email = $1`,
email,
).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
if err != nil {
return nil, err
}
return u, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
u := &models.User{}
err := r.db.QueryRow(ctx,
`SELECT id, email, password_hash, full_name, role, is_active, created_at, updated_at
 FROM users WHERE id = $1`,
id,
).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
if errors.Is(err, pgx.ErrNoRows) {
return nil, nil
}
if err != nil {
return nil, err
}
return u, nil
}