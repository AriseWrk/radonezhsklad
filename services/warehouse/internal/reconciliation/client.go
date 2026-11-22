package reconciliation

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/radonezhsklad/shared/msapi"
)

const msPageSize = 1000

// MSClient — отдельный клиент чтения данных МойСклад для сверки.
// Все обращения выполняются исключительно методом GET.
type MSClient struct {
	api *msapi.Client
}

func NewMSClient(api *msapi.Client) (*MSClient, error) {
	if api == nil {
		return nil, fmt.Errorf("reconciliation: MoySklad client is nil")
	}
	return &MSClient{api: api}, nil
}

type msListMeta struct {
	Size   int `json:"size"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type msListResponse[T any] struct {
	Meta msListMeta `json:"meta"`
	Rows []T        `json:"rows"`
}

type MSEntityMeta struct {
	Href string `json:"href"`
}

type MSReference struct {
	Meta MSEntityMeta `json:"meta"`
	ID   uuid.UUID    `json:"id,omitempty"`
	Name string       `json:"name,omitempty"`
}

type MSPriceType struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type MSSalePrice struct {
	Value     float64     `json:"value"`
	PriceType MSReference `json:"priceType"`
}

type MSProduct struct {
	ID           uuid.UUID     `json:"id"`
	Name         string        `json:"name"`
	Article      string        `json:"article"`
	Code         string        `json:"code"`
	ExternalCode string        `json:"externalCode"`
	Archived     bool          `json:"archived"`
	SalePrices   []MSSalePrice `json:"salePrices"`
}

type MSStore struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ExternalCode string    `json:"externalCode"`
	Archived     bool      `json:"archived"`
}

type MSStoreStock struct {
	Meta  MSEntityMeta `json:"meta"`
	Name  string       `json:"name"`
	Stock float64      `json:"stock"`
}

type MSStockRow struct {
	Meta         MSEntityMeta   `json:"meta"`
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Article      string         `json:"article"`
	Code         string         `json:"code"`
	ExternalCode string         `json:"externalCode"`
	StockByStore []MSStoreStock `json:"stockByStore"`
}

// entityID извлекает UUID из meta.href.
func entityID(href string) (uuid.UUID, error) {
	href = strings.TrimRight(strings.TrimSpace(href), "/")
	if href == "" {
		return uuid.Nil, fmt.Errorf("empty entity href")
	}

	part := href
	if i := strings.LastIndex(part, "/"); i >= 0 {
		part = part[i+1:]
	}

	id, err := uuid.Parse(part)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid entity href %q: %w", href, err)
	}
	return id, nil
}

// listAll загружает все страницы списка, не считая частичный ответ полным.
func listAll[T any](ctx context.Context, api *msapi.Client, path string, pageSize int, extra url.Values) ([]T, error) {
	var result []T
	offset := 0

	for {
		q := url.Values{}
		for key, values := range extra {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		q.Set("limit", strconv.Itoa(pageSize))
		q.Set("offset", strconv.Itoa(offset))

		var page msListResponse[T]
		if err := api.Get(ctx, path, q, &page); err != nil {
			return nil, fmt.Errorf("GET %s offset=%d: %w", path, offset, err)
		}

		result = append(result, page.Rows...)

		if len(page.Rows) == 0 || len(result) >= page.Meta.Size {
			break
		}

		nextOffset := offset + len(page.Rows)
		if nextOffset <= offset {
			return nil, fmt.Errorf("GET %s: pagination did not advance", path)
		}
		offset = nextOffset
	}

	return result, nil
}

func (c *MSClient) ListProducts(ctx context.Context) ([]MSProduct, error) {
	q := url.Values{}
	q.Set("archived", "false")
	q.Set("expand", "salePrices")

	rows, err := listAll[MSProduct](ctx, c.api, "/entity/product", msPageSize, q)
	if err != nil {
		return nil, fmt.Errorf("list MoySklad products: %w", err)
	}
	return rows, nil
}

func (c *MSClient) ListStores(ctx context.Context) ([]MSStore, error) {
	q := url.Values{}
	q.Set("archived", "false")

	rows, err := listAll[MSStore](ctx, c.api, "/entity/store", msPageSize, q)
	if err != nil {
		return nil, fmt.Errorf("list MoySklad stores: %w", err)
	}
	return rows, nil
}

func (c *MSClient) ListPriceTypes(ctx context.Context) ([]MSPriceType, error) {
	var rows []MSPriceType
	if err := c.api.Get(
		ctx,
		"/context/companysettings/pricetype",
		url.Values{},
		&rows,
	); err != nil {
		return nil, fmt.Errorf("list MoySklad price types: %w", err)
	}
	return rows, nil
}
func (c *MSClient) ListStockByStore(ctx context.Context) ([]MSStockRow, error) {
	q := url.Values{}
	q.Set("stockMode", "all")

	rows, err := listAll[MSStockRow](
		ctx,
		c.api,
		"/report/stock/bystore",
		100,
		q,
	)
	if err != nil {
		return nil, fmt.Errorf("list MoySklad stock by store: %w", err)
	}
	return rows, nil
}

// PriceInRubles переводит цену МойСклад из копеек в рубли.
func (p MSSalePrice) PriceInRubles() float64 {
	return p.Value / 100
}
