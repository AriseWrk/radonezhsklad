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

// ListByExternalIDs — товары по списку external_id (MS UUID).
// Общается с product-сервисом через GET /api/v1/products/by-external-ids.
// Разбивает список на чанки, чтобы не переполнить URL.
func (c *Client) ListByExternalIDs(ctx context.Context, token string, extIDs []uuid.UUID) ([]Product, error) {
	if len(extIDs) == 0 {
		return nil, nil
	}
	const chunk = 100
	seen := make(map[uuid.UUID]struct{}, len(extIDs))
	uniq := make([]uuid.UUID, 0, len(extIDs))
	for _, id := range extIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniq = append(uniq, id)
	}

	out := make([]Product, 0, len(uniq))
	for i := 0; i < len(uniq); i += chunk {
		end := i + chunk
		if end > len(uniq) {
			end = len(uniq)
		}
		parts := make([]string, 0, end-i)
		for _, id := range uniq[i:end] {
			parts = append(parts, "ids="+id.String())
		}
		u := c.baseURL + "/api/v1/internal/products/by-external-ids?" + strings.Join(parts, "&")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("X-Internal-Token", token)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("product unreachable: %w", err)
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode >= 300 {
				raw, _ := io.ReadAll(resp.Body)
				err = fmt.Errorf("product error %d: %s", resp.StatusCode, string(raw))
				return
			}
			var page struct {
				Items []Product `json:"items"`
			}
			if derr := json.NewDecoder(resp.Body).Decode(&page); derr != nil {
				err = derr
				return
			}
			out = append(out, page.Items...)
		}()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
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

// ListByIDsInternal — товары по нашим UUID через internal-роут (без JWT).
func (c *Client) ListByIDsInternal(ctx context.Context, token string, ids []uuid.UUID) ([]Product, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	const chunk = 100
	out := make([]Product, 0, len(ids))
	for i := 0; i < len(ids); i += chunk {
		end := i + chunk
		if end > len(ids) {
			end = len(ids)
		}
		parts := make([]string, 0, end-i)
		for _, id := range ids[i:end] {
			parts = append(parts, "id="+id.String())
		}
		u := c.baseURL + "/api/v1/internal/products/list?" + strings.Join(parts, "&")

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("X-Internal-Token", token)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("product unreachable: %w", err)
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode >= 300 {
				raw, _ := io.ReadAll(resp.Body)
				err = fmt.Errorf("product error %d: %s", resp.StatusCode, string(raw))
				return
			}
			var page struct {
				Items []Product `json:"items"`
			}
			if derr := json.NewDecoder(resp.Body).Decode(&page); derr != nil {
				err = derr
				return
			}
			out = append(out, page.Items...)
		}()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
