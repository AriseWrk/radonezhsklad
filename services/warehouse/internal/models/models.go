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

type Supplier struct {
ID        uuid.UUID `json:"id"`
Name      string    `json:"name"`
INN       *string   `json:"inn,omitempty"`
Phone     *string   `json:"phone,omitempty"`
Email     *string   `json:"email,omitempty"`
Address   *string   `json:"address,omitempty"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

type Organization struct {
ID        uuid.UUID `json:"id"`
Name      string    `json:"name"`
INN       *string   `json:"inn,omitempty"`
IsDefault bool      `json:"is_default"`
CreatedAt time.Time `json:"created_at"`
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
SupplierID        *uuid.UUID `json:"supplier_id,omitempty"`
OrganizationID    *uuid.UUID `json:"organization_id,omitempty"`
IncomingNumber    *string    `json:"incoming_number,omitempty"`
IncomingDate      *time.Time `json:"incoming_date,omitempty"`
PaidAmount        float64    `json:"paid_amount"`
PrintedAt         *time.Time `json:"printed_at,omitempty"`
SentAt            *time.Time `json:"sent_at,omitempty"`
Comment           *string    `json:"comment,omitempty"`
CreatedBy         *uuid.UUID `json:"created_by,omitempty"`
CreatedAt         time.Time  `json:"created_at"`
UpdatedAt         time.Time  `json:"updated_at"`
PostedAt          *time.Time `json:"posted_at,omitempty"`
CancelledAt       *time.Time `json:"cancelled_at,omitempty"`
Items             []DocItem  `json:"items,omitempty"`
ItemsCount        int        `json:"items_count"`
Total             float64    `json:"total"`
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
type InternalOrder struct {
ID             uuid.UUID           `json:"id"`
Number         string              `json:"number"`
DocDate        time.Time           `json:"doc_date"`
Status         string              `json:"status"`
OrganizationID *uuid.UUID          `json:"organization_id,omitempty"`
WarehouseID    *uuid.UUID          `json:"warehouse_id,omitempty"`
PlanDate       *time.Time          `json:"plan_date,omitempty"`
Project        *string             `json:"project,omitempty"`
Comment        *string             `json:"comment,omitempty"`
Total          float64             `json:"total"`
ShippedAmount  float64             `json:"shipped_amount"`
SentAt         *time.Time          `json:"sent_at,omitempty"`
PrintedAt      *time.Time          `json:"printed_at,omitempty"`
OwnerID        *uuid.UUID          `json:"owner_id,omitempty"`
OwnerDept      *string             `json:"owner_dept,omitempty"`
VatEnabled     bool                `json:"vat_enabled"`
VatIncluded    bool                `json:"vat_included"`
PostedAt       *time.Time          `json:"posted_at,omitempty"`
CancelledAt    *time.Time          `json:"cancelled_at,omitempty"`
CreatedBy      *uuid.UUID          `json:"created_by,omitempty"`
CreatedAt      time.Time           `json:"created_at"`
UpdatedAt      time.Time           `json:"updated_at"`
Items          []InternalOrderItem `json:"items,omitempty"`
ItemsCount     int                 `json:"items_count"`
}

type InternalOrderItem struct {
ID        uuid.UUID `json:"id"`
OrderID   uuid.UUID `json:"order_id"`
ProductID uuid.UUID `json:"product_id"`
Quantity  float64   `json:"quantity"`
Price     float64   `json:"price"`
VatRate   float64   `json:"vat_rate"`
Sum       float64   `json:"sum"`
CreatedAt time.Time `json:"created_at"`
}

// ---- Inventory ----

type Inventory struct {
    ID             uuid.UUID       `json:"id"`
    ExternalID     *uuid.UUID      `json:"external_id,omitempty"`
    Number         string          `json:"number"`
    DocDate        time.Time       `json:"doc_date"`
    WarehouseID    *uuid.UUID      `json:"warehouse_id,omitempty"`
    WarehouseName  string          `json:"warehouse_name,omitempty"`
    OrganizationID *uuid.UUID      `json:"organization_id,omitempty"`
    Comment        *string         `json:"comment,omitempty"`
    Total          float64         `json:"total"`
    ItemsCount     int             `json:"items_count"`
    CreatedAt      time.Time       `json:"created_at"`
    UpdatedAt      time.Time       `json:"updated_at"`
    Items          []InventoryItem `json:"items,omitempty"`
}

type InventoryItem struct {
    ID                 uuid.UUID `json:"id"`
    InventoryID        uuid.UUID `json:"inventory_id"`
    ProductID          uuid.UUID `json:"product_id"`
    ProductName        string    `json:"product_name,omitempty"`
    ProductSKU         string    `json:"product_sku,omitempty"`
    Quantity           float64   `json:"quantity"`
    CalculatedQuantity float64   `json:"calculated_quantity"`
    CorrectionAmount   float64   `json:"correction_amount"`
    Price              float64   `json:"price"`
    CorrectionSum      float64   `json:"correction_sum"`
    CreatedAt          time.Time `json:"created_at"`
}
