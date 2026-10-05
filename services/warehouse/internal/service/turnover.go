package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/radonezhsklad/warehouse/internal/repository"
)

func (s *Service) TurnoverReport(ctx context.Context, warehouseID *uuid.UUID, from, to time.Time) ([]repository.TurnoverRow, error) {
	return s.repo.TurnoverReport(ctx, warehouseID, from, to)
}
