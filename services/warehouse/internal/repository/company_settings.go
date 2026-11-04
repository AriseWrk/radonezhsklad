package repository

import (
"context"
"time"

"github.com/google/uuid"
"github.com/jackc/pgx/v5/pgxpool"
)

type CompanySettings struct {
Numbering         string    `json:"numbering"`
ReplyAddress      string    `json:"reply_address"`
ForbidShipNoStock bool      `json:"forbid_ship_no_stock"`
AutoMinPrice      bool      `json:"auto_min_price"`
MoveDeletedToBin  bool      `json:"move_deleted_to_bin"`
UseLots           bool      `json:"use_lots"`
ShowExtraFields   bool      `json:"show_extra_fields"`
UpdatedBy         *uuid.UUID `json:"updated_by,omitempty"`
UpdatedByName     *string   `json:"updated_by_name,omitempty"`
UpdatedAt         time.Time `json:"updated_at"`
}

type CompanySettingsRepo struct {
db *pgxpool.Pool
}

func NewCompanySettingsRepo(db *pgxpool.Pool) *CompanySettingsRepo {
return &CompanySettingsRepo{db: db}
}

func (r *CompanySettingsRepo) Get(ctx context.Context) (*CompanySettings, error) {
var s CompanySettings
err := r.db.QueryRow(ctx, `
SELECT numbering, reply_address, forbid_ship_no_stock, auto_min_price,
       move_deleted_to_bin, use_lots, show_extra_fields,
       updated_by, updated_by_name, updated_at
FROM company_settings WHERE id = 1
`).Scan(
&s.Numbering, &s.ReplyAddress, &s.ForbidShipNoStock, &s.AutoMinPrice,
&s.MoveDeletedToBin, &s.UseLots, &s.ShowExtraFields,
&s.UpdatedBy, &s.UpdatedByName, &s.UpdatedAt,
)
if err != nil {
return nil, err
}
return &s, nil
}

func (r *CompanySettingsRepo) Update(ctx context.Context, s CompanySettings, userID uuid.UUID, userName string) error {
_, err := r.db.Exec(ctx, `
UPDATE company_settings SET
numbering = $1, reply_address = $2, forbid_ship_no_stock = $3,
auto_min_price = $4, move_deleted_to_bin = $5, use_lots = $6,
show_extra_fields = $7, updated_by = $8, updated_by_name = $9, updated_at = NOW()
WHERE id = 1
`, s.Numbering, s.ReplyAddress, s.ForbidShipNoStock, s.AutoMinPrice,
s.MoveDeletedToBin, s.UseLots, s.ShowExtraFields, userID, userName)
return err
}