package mspull

import (
"context"
"fmt"
"log/slog"
"net/url"
"sync"
"time"

"github.com/google/uuid"
"golang.org/x/sync/errgroup"

"github.com/radonezhsklad/warehouse/internal/repository"
)

// internalOrderRow — строка ответа /entity/internalorder.
type internalOrderRow struct {
ID          uuid.UUID  `json:"id"`
Name        string     `json:"name"`
Moment      *string    `json:"moment"`
Applicable  bool       `json:"applicable"`
Sum         float64    `json:"sum"`
Printed     bool       `json:"printed"`
Description *string    `json:"description"`
Store       *msRefMeta `json:"store"`
Organization *msRefMeta `json:"organization"`
Project     *msRefMeta `json:"project"`
}

// PullInternalOrder — загрузка ВЗ из МС.
func (r *Runner) PullInternalOrder(ctx context.Context, jobID uuid.UUID) error {
kind := "pull-docs-internalorder"
cursor := r.cursorFor(ctx, kind)
slog.Info("mspull: internalorder", "since", cursor)

fetched := 0
total := 0
offset := 0
skippedNoWH := 0
skippedNoProd := 0

for {
q := url.Values{}
q.Set("limit", fmt.Sprint(pageSize))
q.Set("offset", fmt.Sprint(offset))
q.Set("filter", "moment>=2020-01-01 00:00:00")
q.Set("order", "moment,desc")

var page msListResponse[internalOrderRow]
if err := r.ms.Get(ctx, "/entity/internalorder", q, &page); err != nil {
return fmt.Errorf("GET /entity/internalorder: %w", err)
}
if total == 0 {
total = page.Meta.Size
}
if len(page.Rows) == 0 {
break
}

// Параллельно по 8 документов.
g, gctx := errgroup.WithContext(ctx)
g.SetLimit(2)
var mu sync.Mutex
for _, row := range page.Rows {
row := row
g.Go(func() error {
err := r.processInternalOrderRow(gctx, row, &mu, &skippedNoWH, &skippedNoProd)
if err != nil {
slog.Warn("mspull: internalorder row failed", "ms_id", row.ID, "error", err)
}
mu.Lock()
fetched++
if fetched%10 == 0 {
_ = r.jobs.SetLastLine(ctx, jobID,
fmt.Sprintf("[PROGRESS] fetched=%d total=%d", fetched, total))
}
mu.Unlock()
return nil
})
}
_ = g.Wait()
r.setProgress(ctx, jobID, fetched, total)
offset += len(page.Rows)
if offset >= page.Meta.Size {
break
}
}

slog.Info("mspull: internalorder done",
"fetched", fetched, "skipped_no_warehouse", skippedNoWH, "skipped_no_product", skippedNoProd)
return nil
}

func (r *Runner) processInternalOrderRow(
ctx context.Context, row internalOrderRow,
mu *sync.Mutex, skippedNoWH, skippedNoProd *int,
) error {
// Склад обязателен.
if row.Store == nil {
mu.Lock(); *skippedNoWH++; mu.Unlock()
return nil // нет store — пропускаем без ошибки
}
storeExt := hrefUUID(row.Store.Meta.Href)
if storeExt == nil {
mu.Lock(); *skippedNoWH++; mu.Unlock()
return fmt.Errorf("bad store href")
}
whID, err := r.repo.EnsureWarehouseByExternalID(ctx, *storeExt)
if err != nil || whID == nil {
mu.Lock(); *skippedNoWH++; mu.Unlock()
return fmt.Errorf("склад %s: %w", storeExt, err)
}

// Организация (опционально).
var orgID *uuid.UUID
if row.Organization != nil {
if ext := hrefUUID(row.Organization.Meta.Href); ext != nil {
if oid, err := r.repo.LookupOrganizationID(ctx, *ext); err == nil && oid != nil {
orgID = oid
}
}
}

// Позиции.
items, err := r.fetchPositions(ctx, "internalorder", row.ID)
if err != nil {
return fmt.Errorf("positions: %w", err)
}
if len(items) == 0 {
return fmt.Errorf("нет позиций")
}

extIDs := make([]uuid.UUID, 0, len(items))
for _, it := range items {
if it.Assortment == nil {
continue
}
if ext := hrefUUID(it.Assortment.Meta.Href); ext != nil {
extIDs = append(extIDs, *ext)
}
}
prods, err := r.pc.ListByExternalIDs(ctx, r.token, extIDs)
if err != nil {
return fmt.Errorf("product client: %w", err)
}
byExt := make(map[uuid.UUID]uuid.UUID, len(prods))
for _, p := range prods {
if p.ExternalID != nil {
byExt[*p.ExternalID] = p.ID
}
}

docItems := make([]repository.MsIntOrderItem, 0, len(items))
for _, it := range items {
if it.Assortment == nil {
continue
}
prodExt := hrefUUID(it.Assortment.Meta.Href)
if prodExt == nil {
continue
}
prodID, ok := byExt[*prodExt]
if !ok {
mu.Lock(); *skippedNoProd++; mu.Unlock()
continue
}
priceRub := it.Price / 100.0
vatRub := it.Vat / 100.0
sumRub := priceRub * it.Quantity
docItems = append(docItems, repository.MsIntOrderItem{
ExternalID: it.ID,
ProductID:  prodID,
Quantity:   it.Quantity,
Price:      priceRub,
Vat:        vatRub,
Sum:        sumRub,
})
}
if len(docItems) == 0 {
return nil // нет смэппленных позиций — пропускаем
}

status := "draft"
if row.Applicable {
status = "posted"
}
var moment *time.Time
if row.Moment != nil {
moment = parseMSMoment(row.Moment)
}

o := repository.MsInternalOrder{
ExternalID:     row.ID,
Number:         row.Name,
DocDate:        moment,
Status:         status,
WarehouseID:    whID,
OrganizationID: orgID,
Total:          row.Sum / 100.0,
Description:    row.Description,
Printed:        row.Printed,
}
return r.repo.UpsertInternalOrder(ctx, o, docItems)
}