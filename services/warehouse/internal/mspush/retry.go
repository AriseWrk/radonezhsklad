package mspush

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
)

// RetryDocument — повторная попытка запушить конкретный документ.
// Использует PushDocument. Best effort: ошибки только логируются.
func (p *Pusher) RetryDocument(ctx context.Context, docID uuid.UUID, internalToken string) error {
	if !p.Enabled() {
		return fmt.Errorf("mspush: disabled")
	}
	err := p.PushDocument(ctx, docID, internalToken)
	if err == nil {
		_ = p.repo.SetDocumentMSError(ctx, docID, "")
		if p.errs != nil {
			_ = p.errs.ResolveOpen(ctx, "document", docID)
		}
	}
	return err
}

// RetryPending — обходит documents и internal_orders, у которых есть ms_sync_error
// и нет external_id, и пробует запушить заново.
//
// Возвращает (попыток, успехов, ошибку верхнего уровня).
func (p *Pusher) RetryPending(ctx context.Context, internalToken string) (int, int, error) {
	if !p.Enabled() {
		return 0, 0, fmt.Errorf("mspush: disabled")
	}
	if p.errs == nil {
		return 0, 0, fmt.Errorf("mspush: error repo not configured")
	}

	// --- документы ---
	docIDs, err := p.errs.ListPendingDocuments(ctx, 100)
	if err != nil {
		return 0, 0, fmt.Errorf("list pending documents: %w", err)
	}
	attempted := 0
	succeeded := 0
	for _, id := range docIDs {
		attempted++
		if err := p.PushDocument(ctx, id, internalToken); err != nil {
			slog.Warn("mspush: retry doc failed", "doc_id", id, "error", err)
			continue
		}
		succeeded++
		if p.errs != nil {
			_ = p.errs.ResolveOpen(ctx, "document", id)
		}
	}

	// --- internal orders ---
	orderIDs, err := p.errs.ListPendingOrders(ctx, 100)
	if err == nil {
		for _, id := range orderIDs {
			attempted++
			if err := p.PushInternalOrder(ctx, id, internalToken); err != nil {
				slog.Warn("mspush: retry order failed", "order_id", id, "error", err)
				continue
			}
			succeeded++
			if p.errs != nil {
				_ = p.errs.ResolveOpen(ctx, "internalorder", id)
			}
		}
	}

	slog.Info("mspush: retry pending done", "attempted", attempted, "succeeded", succeeded)
	return attempted, succeeded, nil
}
