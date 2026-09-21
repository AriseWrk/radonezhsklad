package product

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID         uuid.UUID  `json:"id"`
	ExternalID *uuid.UUID `json:"external_id,omitempty"`
	Name       string     `json:"name"`
	SKU        *string    `json:"sku"`
	Price      float64    `json:"price"`
	CostPrice  float64    `json:"cost_price"`
	MinStock   float64    `json:"min_stock"`
	Currency   string     `json:"currency"`
	UnitID     *uuid.UUID `json:"unit_id"`
}

type Unit struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	ShortName string    `json:"short_name"`
}

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) ListByIDs(ctx context.Context, token string, ids []uuid.UUID) ([]Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = "id=" + id.String()
	}
	url := c.baseURL + "/api/v1/products/list?" + strings.Join(parts, "&")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("product unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("product error %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Items []Product `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) ListUnits(ctx context.Context, token string) ([]Unit, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/units", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("units error %d", resp.StatusCode)
	}

	var out struct {
		Items []Unit `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Items, nil
}
