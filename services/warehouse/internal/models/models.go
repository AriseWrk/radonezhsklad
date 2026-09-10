package models

import (
"time"

"github.com/google/uuid"
)

type Warehouse struct {
ID        uuid.UUID `json:"id"`
Name      string    `json:"name"`
Address   *string   `json:"address,omitempty"`
IsActive  bool      `json:"is_active"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

type StockBalance struct {
ID          uuid.UUID `json:"id"`
WarehouseID uuid.UUID `json:"warehouse_id"`
ProductID   uuid.UUID `json:"product_id"`
Quantity    float64   `json:"quantity"`
UpdatedAt   time.Time `json:"updated_at"`
}

type Document struct {
ID                uuid.UUID  `json:"id"`
Type              string     `json:"type"`
Number            string     `json:"number"`
Status            string     `json:"status"`
WarehouseID       uuid.UUID  `json:"warehouse_id"`
TargetWarehouseID *uuid.UUID `json:"target_warehouse_id,omitempty"`
Comment           *string    `json:"comment,omitempty"`
CreatedBy         *uuid.UUID `json:"created_by,omitempty"`
CreatedAt         time.Time  `json:"created_at"`
UpdatedAt         time.Time  `json:"updated_at"`
PostedAt          *time.Time `json:"posted_at,omitempty"`
CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
Items             []DocItem  `json:"items,omitempty"`
}

type DocItem struct {
ID         uuid.UUID `json:"id"`
DocumentID uuid.UUID `json:"document_id"`
ProductID  uuid.UUID `json:"product_id"`
Quantity   float64   `json:"quantity"`
Price      float64   `json:"price"`
CreatedAt  time.Time `json:"created_at"`
}

type StockMovement struct {
ID            uuid.UUID  `json:"id"`
WarehouseID   uuid.UUID  `json:"warehouse_id"`
ProductID     uuid.UUID  `json:"product_id"`
DocumentID    *uuid.UUID `json:"document_id,omitempty"`
QuantityDelta float64    `json:"quantity_delta"`
CreatedAt     time.Time  `json:"created_at"`
}