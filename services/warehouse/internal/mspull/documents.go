package mspull

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/radonezhsklad/warehouse/internal/repository"
)

// msDocType — соответствие kind → (МС-сущность, наш тип документа).
// Вынесено сюда, чтобы удобно расширять.
var msDocType = map[string]struct {
	entity string
	our    string
}{
	"enter":  {entity: "enter", our: "receipt"},
	"demand": {entity: "demand", our: "shipment"},
	"loss":   {entity: "loss", our: "writeoff"},
	"move":   {entity: "move", our: "transfer"},
}

// ---------- MS-строки ----------

type msRefMeta struct {
	Meta struct {
		Href string `json:"href"`
	} `json:"meta"`
}

type msRowDoc struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Applicable     bool       `json:"applicable"`
	Sum            *float64   `json:"sum"`
	VatEnabled     bool       `json:"vatEnabled"`
	VatIncluded    bool       `json:"vatIncluded"`
	Description    *string    `json:"description"`
	Moment         *string    `json:"moment"`
	Updated        *string    `json:"updated"`
	IncomingNumber *string    `json:"incomingNumber"`
	IncomingDate   *string    `json:"incomingDate"`
	Store          *msRefMeta `json:"store"`
	SourceStore    *msRefMeta `json:"sourceStore"`
	TargetStore    *msRefMeta `json:"targetStore"`
	Agent          *msRefMeta `json:"agent"`
	Organization   *msRefMeta `json:"organization"`
}

type msRowPosition struct {
	ID         uuid.UUID  `json:"id"`
	Quantity   float64    `json:"quantity"`
	Price      float64    `json:"price"`
	Vat        float64    `json:"vat"`
	Discount   float64    `json:"discount"`
	Assortment *msRefMeta `json:"assortment"`
}

// ---------- helpers ----------

// hrefUUID достаёт UUID из meta.href вида .../entity/product/<uuid>.
func hrefUUID(href string) *uuid.UUID {
	if href == "" {
		return nil
	}
	i := strings.LastIndex(href, "/")
	if i < 0 || i == len(href)-1 {
		return nil
	}
	id, err := uuid.Parse(href[i+1:])
	if err != nil {
		return nil
	}
	return &id
}

// pullDocType — общий движок для одного типа (enter/demand/loss/move).
func (r *Runner) pullDocType(ctx context.Context, kind, entity, ourType string, jobID uuid.UUID) error {
	cursor := r.cursorFor(ctx, kind)
	slog.Info("mspull: docs", "kind", kind, "entity", entity, "since", cursor)

	fetched := 0
	total := 0
	offset := 0
	skippedNoWH := 0
	skippedNoProd := 0

	for {
		q := url.Values{}
		q.Set("limit", fmt.Sprint(pageSize))
		q.Set("offset", fmt.Sprint(offset))
		q.Set("filter", "updated>="+cursor)

		var page msListResponse[msRowDoc]
		if err := r.ms.Get(ctx, "/entity/"+entity, q, &page); err != nil {
			return fmt.Errorf("GET /entity/%s: %w", entity, err)
		}
		if total == 0 {
			total = page.Meta.Size
		}
		if len(page.Rows) == 0 {
			break
		}
		for _, row := range page.Rows {
			err := r.processDocRow(ctx, row, entity, ourType, jobID, &skippedNoWH, &skippedNoProd)
			if err != nil {
				// Ошибка одной строки не должна валить весь прогон.
				slog.Warn("mspull: doc row failed", "ms_id", row.ID, "error", err)
			}
			fetched++

			// Прогресс: писать last_line каждые 10 документов, чтобы UI видел движение.
			if fetched%10 == 0 {
				_ = r.jobs.SetLastLine(ctx, jobID,
					fmt.Sprintf("[PROGRESS] fetched=%d total=%d (last=%s)", fetched, total, row.ID))
			}
		}
		r.setProgress(ctx, jobID, fetched, total)
		offset += len(page.Rows)
		if offset >= page.Meta.Size {
			break
		}
	}

	r.afterSuccess(ctx, jobID)
	slog.Info("mspull: docs done",
		"kind", kind, "fetched", fetched, "skipped_no_warehouse", skippedNoWH,
		"skipped_no_product", skippedNoProd)
	return nil
}

// processDocRow — upsert одного документа + позиций.
func (r *Runner) processDocRow(
	ctx context.Context, row msRowDoc, entity, ourType string,
	jobID uuid.UUID, skippedNoWH, skippedNoProd *int,
) error {
	// --- склад ---
	storeHref := ""
	if row.Store != nil {
		storeHref = row.Store.Meta.Href
	} else if row.SourceStore != nil {
		storeHref = row.SourceStore.Meta.Href
	}
	storeExt := hrefUUID(storeHref)
	if storeExt == nil {
		return fmt.Errorf("нет store.meta.href")
	}
	whID, err := r.repo.LookupWarehouseID(ctx, *storeExt)
	if err != nil || whID == nil {
		*skippedNoWH++
		return fmt.Errorf("склад %s не найден локально", storeExt)
	}

	// --- target (для move) ---
	var targetID *uuid.UUID
	if entity == "move" && row.TargetStore != nil {
		if ext := hrefUUID(row.TargetStore.Meta.Href); ext != nil {
			if tid, err := r.repo.LookupWarehouseID(ctx, *ext); err == nil && tid != nil {
				targetID = tid
			} else {
				return fmt.Errorf("склад-цель %s не найден локально", ext)
			}
		} else {
			return fmt.Errorf("move без targetStore.meta.href")
		}
	}

	// --- организация ---
	var orgID *uuid.UUID
	if row.Organization != nil {
		if ext := hrefUUID(row.Organization.Meta.Href); ext != nil {
			if oid, err := r.repo.LookupOrganizationID(ctx, *ext); err == nil && oid != nil {
				orgID = oid
			}
		}
	}

	// --- контрагент (для demand) ---
	var supID *uuid.UUID
	if entity == "demand" && row.Agent != nil {
		if ext := hrefUUID(row.Agent.Meta.Href); ext != nil {
			if sid, err := r.repo.LookupSupplierID(ctx, *ext); err == nil && sid != nil {
				supID = sid
			}
		}
	}

	// --- позиции ---
	items, err := r.fetchPositions(ctx, entity, row.ID)
	if err != nil {
		return fmt.Errorf("positions: %w", err)
	}
	if len(items) == 0 {
		return fmt.Errorf("нет позиций")
	}

	// --- маппинг product MS-UUID → наш UUID ---
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

	// --- собираем MsDocItem ---
	docItems := make([]repository.MsDocItem, 0, len(items))
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
			*skippedNoProd++
			continue
		}
		// MS remap 1.2 отдаёт price/sum/vat/discount в копейках — делим на 100.
		priceRub := it.Price / 100.0
		vatRub := it.Vat / 100.0
		discountRub := it.Discount / 100.0
		sumRub := priceRub * it.Quantity
		docItems = append(docItems, repository.MsDocItem{
			ExternalID: it.ID,
			ProductID:  prodID,
			Quantity:   it.Quantity,
			Price:      priceRub,
			Vat:        vatRub,
			Discount:   discountRub,
			Sum:        sumRub,
		})
	}
	if len(docItems) == 0 {
		return fmt.Errorf("после маппинга не осталось позиций")
	}

	// --- шапка ---
	var totalRub float64
	if row.Sum != nil {
		totalRub = *row.Sum // в МС sum уже в рублях (в remap 1.2 sum — в копейках!)
	}
	// В /entity sum в копейках, в meta.href — рубли. Проверим: у positions.price копейки, у doc.sum тоже.
	// Исторические PS-скрипты делали [math]::Round($d.sum / 100, 2) → значит sum в копейках.
	// Здесь у нас уже positions.* в "рублях" (в нашем product.price — рубли).
	// Строка ниже — заглушка: totalRub вычислим из позиций.
	totalRub = 0
	for _, it := range docItems {
		totalRub += it.Sum
	}

	status := "draft"
	if row.Applicable {
		status = "posted"
	}

	var moment *time.Time
	if row.Moment != nil {
		moment = parseMSMoment(row.Moment)
	}
	var incDate *time.Time
	if row.IncomingDate != nil {
		incDate = parseMSMoment(row.IncomingDate)
	}

	doc := repository.MsDocument{
		ExternalID:        row.ID,
		Type:              ourType,
		Number:            row.Name,
		Status:            status,
		Moment:            moment,
		Description:       row.Description,
		IncomingNumber:    row.IncomingNumber,
		IncomingDate:      incDate,
		Total:             totalRub,
		VatEnabled:        row.VatEnabled,
		VatIncluded:       row.VatIncluded,
		WarehouseID:       *whID,
		TargetWarehouseID: targetID,
		SupplierID:        supID,
		OrganizationID:    orgID,
	}
	if err := r.repo.UpsertDocument(ctx, doc, docItems); err != nil {
		return fmt.Errorf("upsert: %w", err)
	}
	return nil
}

// fetchPositions — пагинированный GET /entity/{entity}/{id}/positions.
func (r *Runner) fetchPositions(ctx context.Context, entity string, docID uuid.UUID) ([]msRowPosition, error) {
	all := []msRowPosition{}
	offset := 0
	for {
		q := url.Values{}
		q.Set("limit", "1000")
		q.Set("offset", fmt.Sprint(offset))
		var page msListResponse[msRowPosition]
		path := "/entity/" + entity + "/" + docID.String() + "/positions"
		if err := r.ms.Get(ctx, path, q, &page); err != nil {
			return nil, err
		}
		all = append(all, page.Rows...)
		if len(page.Rows) == 0 || offset+len(page.Rows) >= page.Meta.Size {
			break
		}
		offset += len(page.Rows)
	}
	return all, nil
}

// ---------- публичные обёртки ----------

func (r *Runner) PullEnter(ctx context.Context, jobID uuid.UUID) error {
	t := msDocType["enter"]
	return r.pullDocType(ctx, "pull-docs-enter", t.entity, t.our, jobID)
}

func (r *Runner) PullDemand(ctx context.Context, jobID uuid.UUID) error {
	t := msDocType["demand"]
	return r.pullDocType(ctx, "pull-docs-demand", t.entity, t.our, jobID)
}

func (r *Runner) PullLoss(ctx context.Context, jobID uuid.UUID) error {
	t := msDocType["loss"]
	return r.pullDocType(ctx, "pull-docs-loss", t.entity, t.our, jobID)
}

func (r *Runner) PullMove(ctx context.Context, jobID uuid.UUID) error {
	t := msDocType["move"]
	return r.pullDocType(ctx, "pull-docs-move", t.entity, t.our, jobID)
}
