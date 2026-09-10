package repository

import (
"context"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepo struct {
db *pgxpool.Pool
}

func NewTokenRepo(db *pgxpool.Pool) *TokenRepo {
return &TokenRepo{db: db}
}

func (r *TokenRepo) Save(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
_, err := r.db.Exec(ctx,
`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
 VALUES ($1, $2, $3)`,
userID, tokenHash, expiresAt,
)
return err
}

func (r *TokenRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
_, err := r.db.Exec(ctx,
`DELETE FROM refresh_tokens WHERE user_id = $1`, userID,
)
return err
}