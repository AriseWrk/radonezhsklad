package service

import (
    "context"

    "github.com/google/uuid"

    "github.com/radonezhsklad/warehouse/internal/models"
    "github.com/radonezhsklad/warehouse/internal/repository"
    apperr "github.com/radonezhsklad/shared/errors"
)

type InventoryService struct {
    repo    *repository.InventoryRepo
    docRepo *repository.Repo
}

func NewInventoryService(repo *repository.InventoryRepo, docRepo *repository.Repo) *InventoryService {
    return &InventoryService{repo: repo, docRepo: docRepo}
}

func (s *InventoryService) List(ctx context.Context, f repository.InventoryFilters) ([]models.Inventory, error) {
    return s.repo.List(ctx, f)
}

func (s *InventoryService) Get(ctx context.Context, id uuid.UUID) (*models.Inventory, error) {
    inv, err := s.repo.Get(ctx, id)
    if err != nil { return nil, apperr.Internal("get inventory", err) }
    if inv == nil { return nil, apperr.NotFound("inventory not found") }
    return inv, nil
}

func (s *InventoryService) Neighbors(ctx context.Context, id uuid.UUID) (*repository.InventoryNeighbors, error) {
    n, err := s.repo.Neighbors(ctx, id)
    if err != nil { return nil, apperr.Internal("get inventory neighbors", err) }
    if n == nil { return nil, apperr.NotFound("inventory not found") }
    return n, nil
}


func (s *InventoryService) CreateCorrection(ctx context.Context, invID uuid.UUID, kind string) (*models.Document, error) {
    inv, err := s.repo.Get(ctx, invID)
    if err != nil { return nil, apperr.Internal("get inventory", err) }
    if inv == nil { return nil, apperr.NotFound("inventory not found") }

    var docType, suffix, commentPrefix string
    switch kind {
    case "shortage":
        docType = "writeoff"
        suffix = "СП"
        commentPrefix = "Списание недостач по инвентаризации №"
    case "surplus":
        docType = "receipt"
        suffix = "ОП"
        commentPrefix = "Оприходование избытков по инвентаризации №"
    default:
        return nil, apperr.BadRequest("kind must be shortage or surplus")
    }

    exists, err := s.docRepo.DocumentExistsForInventory(ctx, invID, docType)
    if err != nil { return nil, apperr.Internal("check existing", err) }
    if exists { return nil, apperr.BadRequest("корректирующий документ уже создан") }

    items := make([]repository.DocItemInput, 0)
    for _, it := range inv.Items {
        var qty float64
        if kind == "shortage" && it.CorrectionAmount < 0 {
            qty = -it.CorrectionAmount
        } else if kind == "surplus" && it.CorrectionAmount > 0 {
            qty = it.CorrectionAmount
        } else {
            continue
        }
        if qty <= 0 { continue }
        items = append(items, repository.DocItemInput{
            ProductID: it.ProductID,
            Quantity:  qty,
            Price:     it.Price,
        })
    }
    if len(items) == 0 {
        if kind == "shortage" { return nil, apperr.BadRequest("в инвентаризации нет недостач") }
        return nil, apperr.BadRequest("в инвентаризации нет избытков")
    }
    if inv.WarehouseID == nil { return nil, apperr.BadRequest("у инвентаризации не указан склад") }

    number := inv.Number + "-" + suffix
    comment := commentPrefix + inv.Number
    if inv.Comment != nil && *inv.Comment != "" { comment += ". " + *inv.Comment }

    doc, err := s.docRepo.CreateDocument(ctx, repository.DocumentInput{
        Type:              docType,
        Number:            number,
        WarehouseID:       *inv.WarehouseID,
        OrganizationID:    inv.OrganizationID,
        Comment:           &comment,
        SourceInventoryID: &invID,
        Items:             items,
    })
    if err != nil { return nil, apperr.Internal("create correction doc", err) }
    return doc, nil
}
