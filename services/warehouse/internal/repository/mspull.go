package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MsPullRepo — upsert'ы справочников из МойСклад по external_id.
//
// Все методы идемпотентны: ON CONFLICT (external_id) DO UPDATE.
// source ставится в 'ms' — чтобы наш push знал, что это чужое.
type MsPullRepo struct {
	db *pgxpool.Pool
}

func NewMsPullRepo(db *pgxpool.Pool) *MsPullRepo { return &MsPullRepo{db: db} }

// ---- Counterparty (МС) → suppliers ----

type MsCounterparty struct {
	ExternalID    uuid.UUID
	Name          string
	ExternalCode  *string
	INN           *string
	KPP           *string
	OGRN          *string
	OKPO          *string
	CompanyType   *string // legal | entrepreneur | individual
	LegalAddress  *string
	ActualAddress *string
	Fax           *string
	Description   *string
	Archived      bool
	Updated       *time.Time
}

func (r *MsPullRepo) UpsertCounterparty(ctx context.Context, c MsCounterparty) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO suppliers (
    external_id, external_code, name, inn, kpp, ogrn, okpo,
    counterparty_type, legal_address, actual_address, fax, comment,
    archived, source
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12,
    $13, 'ms'
)
ON CONFLICT (external_id) WHERE external_id IS NOT NULL DO UPDATE SET
    external_code     = EXCLUDED.external_code,
    name              = EXCLUDED.name,
    inn               = EXCLUDED.inn,
    kpp               = EXCLUDED.kpp,
    ogrn              = EXCLUDED.ogrn,
    okpo              = EXCLUDED.okpo,
    counterparty_type = EXCLUDED.counterparty_type,
    legal_address     = EXCLUDED.legal_address,
    actual_address    = EXCLUDED.actual_address,
    fax               = EXCLUDED.fax,
    comment           = EXCLUDED.comment,
    archived          = EXCLUDED.archived,
    source            = 'ms'`,
		c.ExternalID, c.ExternalCode, c.Name, c.INN, c.KPP, c.OGRN, c.OKPO,
		c.CompanyType, c.LegalAddress, c.ActualAddress, c.Fax, c.Description,
		c.Archived,
	)
	return err
}

// ---- Store (МС) → warehouses ----

type MsStore struct {
	ExternalID   uuid.UUID
	Name         string
	ExternalCode *string
	Address      *string
	Archived     bool
	Updated      *time.Time
}

func (r *MsPullRepo) UpsertWarehouse(ctx context.Context, s MsStore) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO warehouses (
    external_id, external_code, name, address, is_active, source
) VALUES (
    $1, $2, $3, $4, $5, 'ms'
)
ON CONFLICT (external_id) WHERE external_id IS NOT NULL DO UPDATE SET
    external_code = EXCLUDED.external_code,
    name          = EXCLUDED.name,
    address       = EXCLUDED.address,
    is_active     = EXCLUDED.is_active,
    source        = 'ms'`,
		s.ExternalID, s.ExternalCode, s.Name, s.Address, !s.Archived,
	)
	return err
}

// ---- Organization (МС) → organizations ----

type MsOrganization struct {
	ExternalID   uuid.UUID
	Name         string
	ExternalCode *string
	KPP          *string
	OGRN         *string
	OKPO         *string
	LegalAddress *string
	Email        *string
	Archived     bool
	Updated      *time.Time
}

func (r *MsPullRepo) UpsertOrganization(ctx context.Context, o MsOrganization) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO organizations (
    external_id, external_code, name, kpp, ogrn, okpo,
    legal_address, email, archived, source
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, 'ms'
)
ON CONFLICT (external_id) WHERE external_id IS NOT NULL DO UPDATE SET
    external_code = EXCLUDED.external_code,
    name          = EXCLUDED.name,
    kpp           = EXCLUDED.kpp,
    ogrn          = EXCLUDED.ogrn,
    okpo          = EXCLUDED.okpo,
    legal_address = EXCLUDED.legal_address,
    email         = EXCLUDED.email,
    archived      = EXCLUDED.archived,
    source        = 'ms'`,
		o.ExternalID, o.ExternalCode, o.Name, o.KPP, o.OGRN, o.OKPO,
		o.LegalAddress, o.Email, o.Archived,
	)
	return err
}

// ---- Project (МС) → projects ----

type MsProject struct {
	ExternalID uuid.UUID
	Name       string
	Archived   bool
	Updated    *time.Time
}

func (r *MsPullRepo) UpsertProject(ctx context.Context, p MsProject) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO projects (external_id, name, archived)
VALUES ($1, $2, $3)
ON CONFLICT (external_id) DO UPDATE SET
    name     = EXCLUDED.name,
    archived = EXCLUDED.archived`,
		p.ExternalID, p.Name, p.Archived,
	)
	return err
}

// ---- транзакционный хелпер (для батча upsert'ов) ----

// ---- Document (МС) → documents ----

// MsDocument — шапка документа из МС.
// WarehouseID/TargetWarehouseID/SupplierID/OrganizationID могут быть nil —
// тогда документ пропускается (см. mspull.
type MsDocument struct {
	ExternalID        uuid.UUID
	Type              string // receipt | shipment | writeoff | transfer
	Number            string
	Status            string // posted | draft
	Moment            *time.Time
	Description       *string
	IncomingNumber    *string
	IncomingDate      *time.Time
	Total             float64
	VatEnabled        bool
	VatIncluded       bool
	WarehouseID       uuid.UUID
	TargetWarehouseID *uuid.UUID
	SupplierID        *uuid.UUID
	OrganizationID    *uuid.UUID
}

// MsDocItem — позиция документа из МС.
type MsDocItem struct {
	ExternalID uuid.UUID
	ProductID  uuid.UUID
	Quantity   float64
	Price      float64
	Vat        float64
	Discount   float64
	Sum        float64
}

// UpsertDocument — upsert шапки + full replace позиций.
// Только для source='ms': позиции удаляются и вставляются заново.
func (r *MsPullRepo) UpsertDocument(ctx context.Context, d MsDocument, items []MsDocItem) error {
	return r.WithTx(ctx, func(tx pgx.Tx) error {
		var docID uuid.UUID
		var postedAt *time.Time
		if d.Status == "posted" {
			postedAt = d.Moment
		}

		err := tx.QueryRow(ctx, `
INSERT INTO documents (
    type, number, status, warehouse_id, target_warehouse_id,
    supplier_id, organization_id,
    incoming_number, incoming_date,
    comment, doc_date, total, vat_enabled, vat_included,
    posted_at, external_id, external_code, source
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7,
    $8, $9,
    $10, $11, $12, $13, $14,
    $15, $16::uuid, $16::text, 'ms'
)
ON CONFLICT (external_id) DO UPDATE SET
    type                = EXCLUDED.type,
    number              = EXCLUDED.number,
    status              = EXCLUDED.status,
    warehouse_id        = EXCLUDED.warehouse_id,
    target_warehouse_id = EXCLUDED.target_warehouse_id,
    supplier_id         = EXCLUDED.supplier_id,
    organization_id     = EXCLUDED.organization_id,
    incoming_number     = EXCLUDED.incoming_number,
    incoming_date       = EXCLUDED.incoming_date,
    comment             = EXCLUDED.comment,
    doc_date            = EXCLUDED.doc_date,
    total               = EXCLUDED.total,
    vat_enabled         = EXCLUDED.vat_enabled,
    vat_included        = EXCLUDED.vat_included,
    posted_at           = EXCLUDED.posted_at,
    source              = 'ms',
    updated_at          = NOW()
RETURNING id`,
			d.Type, d.Number, d.Status, d.WarehouseID, d.TargetWarehouseID,
			d.SupplierID, d.OrganizationID,
			d.IncomingNumber, d.IncomingDate,
			d.Description, d.Moment, d.Total, d.VatEnabled, d.VatIncluded,
			postedAt, d.ExternalID,
		).Scan(&docID)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM document_items WHERE document_id = $1`, docID); err != nil {
			return err
		}

		for _, it := range items {
			if _, err := tx.Exec(ctx, `
INSERT INTO document_items (
    document_id, product_id, quantity, price, external_id,
    vat_rate, discount, sum
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				docID, it.ProductID, it.Quantity, it.Price, it.ExternalID,
				it.Vat, it.Discount, it.Sum,
			); err != nil {
				return err
			}
		}
		return nil
	})
}

// LookupWarehouseID — наш UUID склада по external_id (MS UUID). nil если нет.
func (r *MsPullRepo) LookupWarehouseID(ctx context.Context, ext uuid.UUID) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT id FROM warehouses WHERE external_id = $1`, ext).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// LookupOrganizationID — наш UUID организации по external_id.
func (r *MsPullRepo) LookupOrganizationID(ctx context.Context, ext uuid.UUID) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT id FROM organizations WHERE external_id = $1`, ext).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// LookupSupplierID — наш UUID поставщика по external_id.
func (r *MsPullRepo) LookupSupplierID(ctx context.Context, ext uuid.UUID) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT id FROM suppliers WHERE external_id = $1`, ext).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}
func (r *MsPullRepo) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
