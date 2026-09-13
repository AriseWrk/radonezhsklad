package repository

import (
"context"
"strings"

"github.com/google/uuid"
"github.com/jackc/pgx/v5/pgxpool"

"github.com/radonezhsklad/audit/internal/models"
)

type Repo struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

func (r *Repo) Create(ctx context.Context, l *models.AuditLog) error {
return r.db.QueryRow(ctx,
`INSERT INTO audit_logs
 (user_id, user_email, method, path, resource, resource_id, status, request_body, client_ip, request_id)
 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
 RETURNING id, created_at`,
l.UserID, l.UserEmail, l.Method, l.Path, l.Resource, l.ResourceID,
l.Status, l.RequestBody, l.ClientIP, l.RequestID,
).Scan(&l.ID, &l.CreatedAt)
}

type ListFilters struct {
UserID   *uuid.UUID
Method   *string
Resource *string
DateFrom *string
DateTo   *string
Limit    int
Offset   int
}

func (r *Repo) List(ctx context.Context, f ListFilters) ([]models.AuditLog, int, error) {
// Собираем условия WHERE
where := []string{"1=1"}
args := []any{}
i := 1

if f.UserID != nil {
where = append(where, "user_id = $"+itoa(i))
args = append(args, *f.UserID)
i++
}
if f.Method != nil {
where = append(where, "method = $"+itoa(i))
args = append(args, *f.Method)
i++
}
if f.Resource != nil {
where = append(where, "resource = $"+itoa(i))
args = append(args, *f.Resource)
i++
}
if f.DateFrom != nil {
where = append(where, "created_at >= $"+itoa(i))
args = append(args, *f.DateFrom)
i++
}
if f.DateTo != nil {
where = append(where, "created_at <= $"+itoa(i))
args = append(args, *f.DateTo)
i++
}

whereSQL := strings.Join(where, " AND ")

// Считаем total
var total int
countSQL := "SELECT COUNT(*) FROM audit_logs WHERE " + whereSQL
if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
return nil, 0, err
}

// Собираем список
listSQL := `SELECT id, user_id, user_email, method, path, resource, resource_id,
                   status, request_body, client_ip, request_id, created_at
            FROM audit_logs WHERE ` + whereSQL + `
            ORDER BY created_at DESC
            LIMIT $` + itoa(i) + ` OFFSET $` + itoa(i+1)

listArgs := append(args, f.Limit, f.Offset)

rows, err := r.db.Query(ctx, listSQL, listArgs...)
if err != nil {
return nil, 0, err
}
defer rows.Close()

out := []models.AuditLog{}
for rows.Next() {
var l models.AuditLog
if err := rows.Scan(&l.ID, &l.UserID, &l.UserEmail, &l.Method, &l.Path,
&l.Resource, &l.ResourceID, &l.Status, &l.RequestBody, &l.ClientIP,
&l.RequestID, &l.CreatedAt); err != nil {
return nil, 0, err
}
out = append(out, l)
}
return out, total, rows.Err()
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