package models

import (
"time"

"github.com/google/uuid"
)

type Customer struct {
ID               uuid.UUID `json:"id"`
Name             string    `json:"name"`
FullName         *string   `json:"full_name,omitempty"`
LastName         *string   `json:"last_name,omitempty"`
FirstName        *string   `json:"first_name,omitempty"`
MiddleName       *string   `json:"middle_name,omitempty"`
Phone            *string   `json:"phone,omitempty"`
Fax              *string   `json:"fax,omitempty"`
Email            *string   `json:"email,omitempty"`
Address          *string   `json:"address,omitempty"`
LegalAddress     *string   `json:"legal_address,omitempty"`
ActualAddress    *string   `json:"actual_address,omitempty"`
INN              *string   `json:"inn,omitempty"`
KPP              *string   `json:"kpp,omitempty"`
OGRN             *string   `json:"ogrn,omitempty"`
OKPO             *string   `json:"okpo,omitempty"`
ExternalCode     *string   `json:"external_code,omitempty"`
CounterpartyType *string   `json:"counterparty_type,omitempty"`
Status           string    `json:"status"`
GroupName        *string   `json:"group_name,omitempty"`
Comment          *string   `json:"comment,omitempty"`
Archived         bool      `json:"archived"`
CreatedAt        time.Time `json:"created_at"`
UpdatedAt        time.Time `json:"updated_at"`
}

type Contract struct {
ID             uuid.UUID  `json:"id"`
Number         string     `json:"number"`
ContractType   string     `json:"contract_type"`
Code           *string    `json:"code,omitempty"`
DocDate        time.Time  `json:"doc_date"`
CustomerID     *uuid.UUID `json:"customer_id,omitempty"`
OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
Amount         float64    `json:"amount"`
Currency       string     `json:"currency"`
Paid           float64    `json:"paid"`
Fulfilled      float64    `json:"fulfilled"`
Comment        *string    `json:"comment,omitempty"`
PrintedAt      *time.Time `json:"printed_at,omitempty"`
SentAt         *time.Time `json:"sent_at,omitempty"`
Archived       bool       `json:"archived"`
CreatedAt      time.Time  `json:"created_at"`
UpdatedAt      time.Time  `json:"updated_at"`
}

type Order struct {
ID             uuid.UUID   `json:"id"`
Number         string      `json:"number"`
CustomerID     uuid.UUID   `json:"customer_id"`
WarehouseID    uuid.UUID   `json:"warehouse_id"`
Status         string      `json:"status"`
Total          float64     `json:"total"`
Currency       string      `json:"currency"`
Comment        *string     `json:"comment,omitempty"`
WarehouseDocID *uuid.UUID  `json:"warehouse_doc_id,omitempty"`
CreatedBy      *uuid.UUID  `json:"created_by,omitempty"`
CreatedAt      time.Time   `json:"created_at"`
UpdatedAt      time.Time   `json:"updated_at"`
ConfirmedAt    *time.Time  `json:"confirmed_at,omitempty"`
ShippedAt      *time.Time  `json:"shipped_at,omitempty"`
CancelledAt    *time.Time  `json:"cancelled_at,omitempty"`
Items          []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
ID        uuid.UUID `json:"id"`
OrderID   uuid.UUID `json:"order_id"`
ProductID uuid.UUID `json:"product_id"`
Quantity  float64   `json:"quantity"`
Price     float64   `json:"price"`
CreatedAt time.Time `json:"created_at"`
}