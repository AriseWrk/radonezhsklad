package models

import (
"time"

"github.com/google/uuid"
)

type Customer struct {
ID        uuid.UUID `json:"id"`
Name      string    `json:"name"`
Phone     *string   `json:"phone,omitempty"`
Email     *string   `json:"email,omitempty"`
Address   *string   `json:"address,omitempty"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
ID              uuid.UUID  `json:"id"`
Number          string     `json:"number"`
CustomerID      uuid.UUID  `json:"customer_id"`
WarehouseID     uuid.UUID  `json:"warehouse_id"`
Status          string     `json:"status"`
Total           float64    `json:"total"`
Currency        string     `json:"currency"`
Comment         *string    `json:"comment,omitempty"`
WarehouseDocID  *uuid.UUID `json:"warehouse_doc_id,omitempty"`
CreatedBy       *uuid.UUID `json:"created_by,omitempty"`
CreatedAt       time.Time  `json:"created_at"`
UpdatedAt       time.Time  `json:"updated_at"`
ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
ShippedAt       *time.Time `json:"shipped_at,omitempty"`
CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
Items           []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
ID        uuid.UUID `json:"id"`
OrderID   uuid.UUID `json:"order_id"`
ProductID uuid.UUID `json:"product_id"`
Quantity  float64   `json:"quantity"`
Price     float64   `json:"price"`
CreatedAt time.Time `json:"created_at"`
}