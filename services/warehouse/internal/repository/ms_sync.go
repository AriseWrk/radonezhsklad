package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ---- MS-синхронизация: выгрузка external_id связанных сущностей ----

// ExternalRef — external_id сущности МС (UUID), нужен для meta.href в payload.
type ExternalRef struct {
	OurID      uuid.UUID
	ExternalID *uuid.UUID
}

// DocumentExternalRefs — внешние ссылки, необходимые для push документа в МС.
type DocumentExternalRefs struct {
	Warehouse       ExternalRef
	TargetWarehouse *ExternalRef
	Supplier        *ExternalRef
	Organization    *ExternalRef
}

// GetDocumentExternalRefs достаёт external_id склада, склада-цели, поставщика, организации.
func (r *Repo) GetDocumentExternalRefs(ctx context.Context, docID uuid.UUID) (*DocumentExternalRefs, error) {
	var (
		whID, twID, supID, orgID *uuid.UUID
	)

	err := r.db.QueryRow(ctx, `
SELECT warehouse_id, target_warehouse_id, supplier_id, organization_id
FROM documents WHERE id = $1`, docID).Scan(&whID, &twID, &supID, &orgID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	lookup := func(table string, id *uuid.UUID) (*uuid.UUID, error) {
		if id == nil {
			return nil, nil
		}
		var ext *uuid.UUID
		q := "SELECT external_id FROM " + table + " WHERE id = $1"
		err := r.db.QueryRow(ctx, q, *id).Scan(&ext)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return ext, err
	}

	whExt, _ := lookup("warehouses", whID)
	twExt, _ := lookup("warehouses", twID)
	supExt, _ := lookup("suppliers", supID)
	orgExt, _ := lookup("organizations", orgID)

	out := &DocumentExternalRefs{}
	if whID != nil {
		out.Warehouse = ExternalRef{OurID: *whID, ExternalID: whExt}
	}
	if twID != nil {
		out.TargetWarehouse = &ExternalRef{OurID: *twID, ExternalID: twExt}
	}
	if supID != nil {
		out.Supplier = &ExternalRef{OurID: *supID, ExternalID: supExt}
	}
	if orgID != nil {
		out.Organization = &ExternalRef{OurID: *orgID, ExternalID: orgExt}
	}
	return out, nil
}

// InternalOrderExternalRefs — внешние ссылки для customerorder/internalorder.
type InternalOrderExternalRefs struct {
	Warehouse    *ExternalRef
	Organization *ExternalRef
	Project      *ExternalRef
}

func (r *Repo) GetInternalOrderExternalRefs(ctx context.Context, orderID uuid.UUID) (*InternalOrderExternalRefs, error) {
	var whID, orgID *uuid.UUID
	var projectExternal *string

	err := r.db.QueryRow(ctx, `
SELECT warehouse_id, organization_id, project
FROM internal_orders WHERE id = $1`, orderID).Scan(&whID, &orgID, &projectExternal)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	lookup := func(table string, id *uuid.UUID) (*uuid.UUID, error) {
		if id == nil {
			return nil, nil
		}
		var ext *uuid.UUID
		q := "SELECT external_id FROM " + table + " WHERE id = $1"
		err := r.db.QueryRow(ctx, q, *id).Scan(&ext)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return ext, err
	}

	whExt, _ := lookup("warehouses", whID)
	orgExt, _ := lookup("organizations", orgID)

	out := &InternalOrderExternalRefs{}
	if whID != nil {
		out.Warehouse = &ExternalRef{OurID: *whID, ExternalID: whExt}
	}
	if orgID != nil {
		out.Organization = &ExternalRef{OurID: *orgID, ExternalID: orgExt}
	}
	if projectExternal != nil && *projectExternal != "" {
		if pid, err := uuid.Parse(*projectExternal); err == nil {
			out.Project = &ExternalRef{ExternalID: &pid}
		}
	}
	return out, nil
}

// ---- запись результата push ----

func (r *Repo) SetDocumentMSSynced(ctx context.Context, id uuid.UUID, msUUID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
UPDATE documents
SET external_id = $2::uuid, ms_synced_at = NOW(), ms_sync_error = NULL,
    external_code = $2::text
WHERE id = $1`, id, msUUID.String())
	return err
}

func (r *Repo) SetDocumentMSError(ctx context.Context, id uuid.UUID, msg string) error {
	_, err := r.db.Exec(ctx, `
UPDATE documents SET ms_sync_error = $2 WHERE id = $1`, id, msg)
	return err
}

func (r *Repo) SetOrderMSSynced(ctx context.Context, id uuid.UUID, msUUID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
UPDATE internal_orders
SET external_id = $2::uuid, ms_synced_at = NOW(), ms_sync_error = NULL,
    external_code = $2::text
WHERE id = $1`, id, msUUID.String())
	return err
}

func (r *Repo) SetOrderMSError(ctx context.Context, id uuid.UUID, msg string) error {
	_, err := r.db.Exec(ctx, `
UPDATE internal_orders SET ms_sync_error = $2 WHERE id = $1`, id, msg)
	return err
}
