package warehouse

import (
"bytes"
"context"
"encoding/json"
"fmt"
"io"
"net/http"
"time"

"github.com/google/uuid"
)

type Client struct {
baseURL string
http    *http.Client
}

func New(baseURL string) *Client {
return &Client{
baseURL: baseURL,
http:    &http.Client{Timeout: 15 * time.Second},
}
}

type ShipmentItem struct {
ProductID uuid.UUID `json:"product_id"`
Quantity  float64   `json:"quantity"`
Price     float64   `json:"price"`
}

type CreateShipmentReq struct {
Type        string         `json:"type"`
Number      string         `json:"number"`
WarehouseID uuid.UUID      `json:"warehouse_id"`
Comment     string         `json:"comment,omitempty"`
Items       []ShipmentItem `json:"items"`
}

type Document struct {
ID     uuid.UUID `json:"id"`
Type   string    `json:"type"`
Number string    `json:"number"`
Status string    `json:"status"`
}

// CreateAndPostShipment — создаёт и сразу проводит документ отгрузки.
// Возвращает ID созданного документа.
func (c *Client) CreateAndPostShipment(ctx context.Context, token string, req CreateShipmentReq) (uuid.UUID, error) {
body, err := json.Marshal(req)
if err != nil { return uuid.Nil, err }

httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
c.baseURL+"/api/v1/documents", bytes.NewReader(body))
if err != nil { return uuid.Nil, err }
httpReq.Header.Set("Content-Type", "application/json")
httpReq.Header.Set("Authorization", "Bearer "+token)

resp, err := c.http.Do(httpReq)
if err != nil { return uuid.Nil, fmt.Errorf("warehouse unreachable: %w", err) }
defer resp.Body.Close()

if resp.StatusCode >= 300 {
raw, _ := io.ReadAll(resp.Body)
return uuid.Nil, fmt.Errorf("warehouse error %d: %s", resp.StatusCode, string(raw))
}

var doc Document
if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
return uuid.Nil, err
}

// проводим
postReq, _ := http.NewRequestWithContext(ctx, http.MethodPost,
c.baseURL+"/api/v1/documents/"+doc.ID.String()+"/post", nil)
postReq.Header.Set("Authorization", "Bearer "+token)

postResp, err := c.http.Do(postReq)
if err != nil { return uuid.Nil, fmt.Errorf("warehouse post failed: %w", err) }
defer postResp.Body.Close()

if postResp.StatusCode >= 300 {
raw, _ := io.ReadAll(postResp.Body)
return uuid.Nil, fmt.Errorf("warehouse post error %d: %s", postResp.StatusCode, string(raw))
}

return doc.ID, nil
}