package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Unit struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
	CreatedAt time.Time `json:"created_at"`
}

type Product struct {
	ID          uuid.UUID  `json:"id"`
	ExternalID  *uuid.UUID `json:"external_id,omitempty"`
	Name        string     `json:"name"`
	SKU         *string    `json:"sku,omitempty"`
	Barcode     *string    `json:"barcode,omitempty"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	UnitID      *uuid.UUID `json:"unit_id,omitempty"`
	Description *string    `json:"description,omitempty"`
	Price       float64    `json:"price"`
	CostPrice   float64    `json:"cost_price"`
	MinStock    float64    `json:"min_stock"`
	Currency    string     `json:"currency"`
	IsArchived  bool       `json:"is_archived"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
