// Package mspush отправляет созданные у нас документы и внутренние заказы
// в МойСклад, чтобы офис, работающий в МС, их видел.
//
// Логика вызова:
//   - после локального commit'а (в service.PostDocument / InternalOrderService.Post)
//   - синхронно, в контексте HTTP-запроса
//   - при ошибке МС документ остаётся локально, пишем ms_sync_error
//   - повторный push не делается, если external_id уже заполнен
//
// Идемпотентность: при POST передаём syncId = UUID нашего документа,
// МС дедуплицирует по нему и вернёт конфликт при повторе.
package mspush

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/radonezhsklad/shared/msapi"
	"github.com/radonezhsklad/warehouse/internal/models"
	"github.com/radonezhsklad/warehouse/internal/product"
	"github.com/radonezhsklad/warehouse/internal/repository"
)

// entityByDocType — маппинг наших типов документов в сущности МС.
// inventory не пушим — учтём отдельно (у нас correction через документы).
var entityByDocType = map[string]string{
	"receipt":  "enter",
	"shipment": "demand",
	"writeoff": "loss",
	"transfer": "move",
}

// Pusher — координатор push'а.
type Pusher struct {
	ms      *msapi.Client
	product *product.Client
	repo    *repository.Repo
	orders  *repository.InternalOrderRepo
	enabled bool
}

func New(ms *msapi.Client, pc *product.Client, repo *repository.Repo, orders *repository.InternalOrderRepo, enabled bool) *Pusher {
	return &Pusher{ms: ms, product: pc, repo: repo, orders: orders, enabled: enabled}
}

func (p *Pusher) Enabled() bool { return p.enabled && p.ms != nil }

// msMoment форматирует time в «MS-локальное» (МСК) без TZ.
func msMoment(t time.Time) string {
	msk := time.FixedZone("MSK", 3*60*60)
	return t.In(msk).Format("2006-01-02 15:04:05.000")
}

// toKopecks — рубли (float) → копейки (int), как ждёт МС.
func toKopecks(rub float64) int64 {
	return int64(math.Round(rub * 100))
}

// -------- payload --------

type msPosition struct {
	Quantity   float64    `json:"quantity"`
	Price      int64      `json:"price"` // копейки
	VAT        float64    `json:"vat"`
	Assortment msapi.Meta `json:"assortment"`
}

type msDocumentPayload struct {
	Name         string       `json:"name"`
	SyncID       string       `json:"syncId"`
	Moment       string       `json:"moment"`
	Applicable   bool         `json:"applicable"`
	Description  string       `json:"description,omitempty"`
	Organization msapi.Meta   `json:"organization"`
	Store        msapi.Meta   `json:"store,omitempty"`
	TargetStore  *msapi.Meta  `json:"targetStore,omitempty"`
	Agent        *msapi.Meta  `json:"agent,omitempty"`
	Positions    []msPosition `json:"positions"`
}

type msOrderPayload struct {
	Name         string       `json:"name"`
	SyncID       string       `json:"syncId"`
	Moment       string       `json:"moment"`
	Applicable   bool         `json:"applicable"`
	Description  string       `json:"description,omitempty"`
	Organization msapi.Meta   `json:"organization"`
	Store        msapi.Meta   `json:"store,omitempty"`
	Positions    []msPosition `json:"positions"`
}

// -------- helpers --------

// productExternalMap возвращает карту наш_product_id → ms_uuid.
// Если у какого-то продукта нет external_id — он в карте отсутствует.
func (p *Pusher) productExternalMap(ctx context.Context, productIDs []uuid.UUID, token string) (map[uuid.UUID]*uuid.UUID, error) {
	out := make(map[uuid.UUID]*uuid.UUID, len(productIDs))
	if len(productIDs) == 0 {
		return out, nil
	}
	prods, err := p.product.ListByIDs(ctx, token, productIDs)
	if err != nil {
		return nil, fmt.Errorf("product list: %w", err)
	}
	for _, pr := range prods {
		out[pr.ID] = pr.ExternalID
	}
	return out, nil
}

func positionsFromDocItems(items []models.DocItem, extMap map[uuid.UUID]*uuid.UUID, meta func(string, uuid.UUID) msapi.Meta) ([]msPosition, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("нет позиций")
	}
	out := make([]msPosition, 0, len(items))
	for _, it := range items {
		ext := extMap[it.ProductID]
		if ext == nil {
			return nil, fmt.Errorf("товар %s не привязан к МС (нет external_id)", it.ProductID)
		}
		out = append(out, msPosition{
			Quantity:   it.Quantity,
			Price:      toKopecks(it.Price),
			VAT:        0, // TODO: vat_rate из document_items, когда пробросим в модель
			Assortment: meta("product", *ext),
		})
	}
	return out, nil
}

// -------- documents --------

// buildDocumentPayload формирует payload для одного из entity МС.
func (p *Pusher) buildDocumentPayload(ctx context.Context, doc *models.Document, entity, token string) (*msDocumentPayload, error) {
	refs, err := p.repo.GetDocumentExternalRefs(ctx, doc.ID)
	if err != nil {
		return nil, fmt.Errorf("refs: %w", err)
	}
	if refs == nil {
		return nil, fmt.Errorf("refs not found")
	}
	if refs.Warehouse.ExternalID == nil {
		return nil, fmt.Errorf("склад не привязан к МС (нет external_id)")
	}
	if refs.Organization == nil || refs.Organization.ExternalID == nil {
		return nil, fmt.Errorf("организация не привязана к МС (нет external_id)")
	}

	ids := make([]uuid.UUID, 0, len(doc.Items))
	for _, it := range doc.Items {
		ids = append(ids, it.ProductID)
	}
	extMap, err := p.productExternalMap(ctx, ids, token)
	if err != nil {
		return nil, err
	}
	positions, err := positionsFromDocItems(doc.Items, extMap, p.ms.EntityMeta)
	if err != nil {
		return nil, err
	}

	payload := &msDocumentPayload{
		Name:         doc.Number,
		SyncID:       doc.ID.String(),
		Moment:       msMoment(doc.CreatedAt),
		Applicable:   true,
		Organization: p.ms.EntityMeta("organization", *refs.Organization.ExternalID),
		Store:        p.ms.EntityMeta("store", *refs.Warehouse.ExternalID),
		Positions:    positions,
	}
	if doc.Comment != nil {
		payload.Description = *doc.Comment
	}
	// targetStore для перемещения
	if entity == "move" && refs.TargetWarehouse != nil && refs.TargetWarehouse.ExternalID != nil {
		m := p.ms.EntityMeta("store", *refs.TargetWarehouse.ExternalID)
		payload.TargetStore = &m
	}
	// agent для отгрузки — из supplier_id (у нас других контрагентов нет)
	if entity == "demand" && refs.Supplier != nil && refs.Supplier.ExternalID != nil {
		m := p.ms.EntityMeta("counterparty", *refs.Supplier.ExternalID)
		payload.Agent = &m
	}
	return payload, nil
}

// PushDocument — отправить документ в МС. Ошибки также пишутся в ms_sync_error.
// Возвращает nil, если push отключён, документ не posted или уже синхронизирован.
func (p *Pusher) PushDocument(ctx context.Context, docID uuid.UUID, token string) error {
	if !p.Enabled() {
		return nil
	}
	doc, err := p.repo.GetDocument(ctx, docID)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("document %s not found", docID)
	}
	if doc.ExternalID != nil {
		return nil
	}
	if doc.Status != "posted" {
		return nil
	}
	entity, ok := entityByDocType[doc.Type]
	if !ok {
		return nil
	}

	payload, err := p.buildDocumentPayload(ctx, doc, entity, token)
	if err != nil {
		_ = p.repo.SetDocumentMSError(ctx, docID, err.Error())
		return err
	}

	var resp struct {
		ID uuid.UUID `json:"id"`
	}
	if err := p.ms.Post(ctx, "/entity/"+entity, payload, &resp); err != nil {
		_ = p.repo.SetDocumentMSError(ctx, docID, err.Error())
		return err
	}
	if err := p.repo.SetDocumentMSSynced(ctx, docID, resp.ID); err != nil {
		slog.Error("mspush: SetDocumentMSSynced failed", "our_id", docID, "error", err)
		return err
	}
	if resp.ID == uuid.Nil {
		slog.Error("mspush: MS returned empty id", "our_id", docID, "resp", resp)
	}
	slog.Info("mspush: document pushed", "our_id", docID, "ms_id", resp.ID, "entity", entity, "type", doc.Type)
	return nil
}

// -------- internal orders --------

func (p *Pusher) buildOrderPayload(ctx context.Context, order *models.InternalOrder, token string) (*msOrderPayload, error) {
	refs, err := p.repo.GetInternalOrderExternalRefs(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("refs: %w", err)
	}
	if refs == nil {
		return nil, fmt.Errorf("refs not found")
	}
	if refs.Warehouse == nil || refs.Warehouse.ExternalID == nil {
		return nil, fmt.Errorf("склад заказа не привязан к МС")
	}
	if refs.Organization == nil || refs.Organization.ExternalID == nil {
		return nil, fmt.Errorf("организация заказа не привязана к МС")
	}

	ids := make([]uuid.UUID, 0, len(order.Items))
	for _, it := range order.Items {
		ids = append(ids, it.ProductID)
	}
	extMap, err := p.productExternalMap(ctx, ids, token)
	if err != nil {
		return nil, err
	}

	if len(order.Items) == 0 {
		return nil, fmt.Errorf("заказ без позиций")
	}
	positions := make([]msPosition, 0, len(order.Items))
	for _, it := range order.Items {
		ext := extMap[it.ProductID]
		if ext == nil {
			return nil, fmt.Errorf("товар %s не привязан к МС", it.ProductID)
		}
		positions = append(positions, msPosition{
			Quantity:   it.Quantity,
			Price:      toKopecks(it.Price),
			VAT:        it.VatRate,
			Assortment: p.ms.EntityMeta("product", *ext),
		})
	}

	payload := &msOrderPayload{
		Name:         order.Number,
		SyncID:       order.ID.String(),
		Moment:       msMoment(order.DocDate),
		Applicable:   true,
		Organization: p.ms.EntityMeta("organization", *refs.Organization.ExternalID),
		Store:        p.ms.EntityMeta("store", *refs.Warehouse.ExternalID),
		Positions:    positions,
	}
	if order.Comment != nil {
		payload.Description = *order.Comment
	}
	return payload, nil
}

// PushInternalOrder — отправить внутренний заказ в МС (entity: internalorder).
func (p *Pusher) PushInternalOrder(ctx context.Context, orderID uuid.UUID, token string) error {
	if !p.Enabled() {
		return nil
	}
	order, err := p.orders.Get(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order %s not found", orderID)
	}
	if order.ExternalID != nil {
		return nil
	}
	if order.Status != "posted" {
		return nil
	}

	payload, err := p.buildOrderPayload(ctx, order, token)
	if err != nil {
		_ = p.repo.SetOrderMSError(ctx, orderID, err.Error())
		return err
	}

	var resp struct {
		ID uuid.UUID `json:"id"`
	}
	if err := p.ms.Post(ctx, "/entity/internalorder", payload, &resp); err != nil {
		_ = p.repo.SetOrderMSError(ctx, orderID, err.Error())
		return err
	}
	if err := p.repo.SetOrderMSSynced(ctx, orderID, resp.ID); err != nil {
		return err
	}
	slog.Info("mspush: internal_order pushed", "our_id", orderID, "ms_id", resp.ID)
	return nil
}
