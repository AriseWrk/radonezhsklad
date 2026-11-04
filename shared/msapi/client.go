// Package msapi — минимальный Go-клиент МойСклад API remap 1.2.
package msapi

import (
        "bytes"
        "compress/gzip"
        "context"
        "encoding/json"
        "errors"
        "fmt"
        "io"
        "net/http"
        "net/url"
        "os"
        "strconv"
        "strings"
        "time"

        "github.com/google/uuid"
)

const DefaultBaseURL = "https://api.moysklad.ru/api/remap/1.2"

type Client struct {
        baseURL string
        token   string
        http    *http.Client
}

func New(token, tokenFile, baseURL string) (*Client, error) {
        if token == "" && tokenFile != "" {
                b, err := os.ReadFile(tokenFile)
                if err != nil {
                        return nil, fmt.Errorf("msapi: read token file %q: %w", tokenFile, err)
                }
                token = strings.TrimSpace(string(b))
        }
        if token == "" {
                return nil, errors.New("msapi: token is empty (set MS_TOKEN or MS_TOKEN_FILE)")
        }
        if baseURL == "" {
                baseURL = DefaultBaseURL
        }
        return &Client{
                baseURL: strings.TrimRight(baseURL, "/"),
                token:   token,
                http:    &http.Client{Timeout: 30 * time.Second},
        }, nil
}

func (c *Client) BaseURL() string { return c.baseURL }

type msError struct {
        Errors []struct {
                Error    string `json:"error"`
                Code     int    `json:"code"`
                MoreInfo string `json:"moreInfo"`
        } `json:"errors"`
}

func (e *msError) String() string {
        if e == nil || len(e.Errors) == 0 {
                return "msapi: unknown error"
        }
        parts := make([]string, 0, len(e.Errors))
        for _, x := range e.Errors {
                s := x.Error
                if x.MoreInfo != "" {
                        s += " (" + x.MoreInfo + ")"
                }
                if x.Code != 0 {
                        s = fmt.Sprintf("[%d] %s", x.Code, s)
                }
                parts = append(parts, s)
        }
        return "msapi: " + strings.Join(parts, "; ")
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
        var bodyBytes []byte
        if body != nil {
                b, err := json.Marshal(body)
                if err != nil {
                        return fmt.Errorf("msapi: marshal body: %w", err)
                }
                bodyBytes = b
        }

        u := c.baseURL + path
        if len(query) > 0 {
                u += "?" + query.Encode()
        }

        const maxAttempts = 3
        var lastErr error
        for attempt := 1; attempt <= maxAttempts; attempt++ {
                var rdr io.Reader
                if bodyBytes != nil {
                        rdr = bytes.NewReader(bodyBytes)
                }
                req, err := http.NewRequestWithContext(ctx, method, u, rdr)
                if err != nil {
                        return fmt.Errorf("msapi: new request: %w", err)
                }
                req.Header.Set("Authorization", "Bearer "+c.token)
                req.Header.Set("Accept", "application/json;charset=utf-8")
                req.Header.Set("Accept-Encoding", "gzip")
                if bodyBytes != nil {
                        req.Header.Set("Content-Type", "application/json")
                }

                resp, err := c.http.Do(req)
                if err != nil {
                        lastErr = fmt.Errorf("msapi: do: %w", err)
                        sleepBackoff(ctx, attempt)
                        continue
                }

                var reader io.Reader = resp.Body
                if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
                        gz, gerr := gzip.NewReader(resp.Body)
                        if gerr == nil {
                                reader = gz
                        }
                }
                respBody, rerr := io.ReadAll(reader)
                resp.Body.Close()
                if rerr != nil {
                        lastErr = fmt.Errorf("msapi: read body: %w", rerr)
                        sleepBackoff(ctx, attempt)
                        continue
                }

                if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                        if out == nil {
                                return nil
                        }
                        if err := json.Unmarshal(respBody, out); err != nil {
                                return fmt.Errorf("msapi: unmarshal response: %w", err)
                        }
                        return nil
                }

                var msErr msError
                _ = json.Unmarshal(respBody, &msErr)

                if resp.StatusCode == 429 || resp.StatusCode >= 500 {
                        lastErr = fmt.Errorf("msapi: HTTP %d: %s", resp.StatusCode, msErr.String())
                        slept := false
                        if ra := resp.Header.Get("Retry-After"); ra != "" {
                                if secs, err := strconv.Atoi(strings.TrimSpace(ra)); err == nil {
                                        d := time.Duration(secs) * time.Second
                                        if d > 15*time.Second {
                                                d = 15 * time.Second
                                        }
                                        select {
                                        case <-time.After(d):
                                                slept = true
                                        case <-ctx.Done():
                                                return ctx.Err()
                                        }
                                }
                        }
                        if !slept {
                                sleepBackoff(ctx, attempt)
                        }
                        continue
                }

                if len(msErr.Errors) > 0 {
                        return fmt.Errorf("msapi: HTTP %d: %s", resp.StatusCode, msErr.String())
                }
                return fmt.Errorf("msapi: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
        }
        return lastErr
}

func sleepBackoff(ctx context.Context, attempt int) {
        d := time.Duration(attempt) * 500 * time.Millisecond
        select {
        case <-time.After(d):
        case <-ctx.Done():
        }
}

func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
        return c.do(ctx, http.MethodPost, path, nil, body, out)
}

func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
        return c.do(ctx, http.MethodGet, path, query, nil, out)
}

func (c *Client) Put(ctx context.Context, path string, body any, out any) error {
        return c.do(ctx, http.MethodPut, path, nil, body, out)
}

func (c *Client) Delete(ctx context.Context, path string) error {
        return c.do(ctx, http.MethodDelete, path, nil, nil, nil)
}

type Meta struct {
        Meta MetaRef `json:"meta"`
}

type MetaRef struct {
        Href         string `json:"href"`
        MetadataHref string `json:"metadataHref"`
        Type         string `json:"type"`
        MediaType    string `json:"mediaType,omitempty"`
}

func (c *Client) EntityMeta(entityType string, id uuid.UUID) Meta {
        return Meta{
                Meta: MetaRef{
                        Href:         fmt.Sprintf("%s/entity/%s/%s", c.baseURL, entityType, id.String()),
                        MetadataHref: fmt.Sprintf("%s/entity/%s/metadata", c.baseURL, entityType),
                        Type:         entityType,
                        MediaType:    "application/json",
                },
        }
}

func (c *Client) HrefMeta(href, entityType string) Meta {
        return Meta{
                Meta: MetaRef{
                        Href:         href,
                        MetadataHref: fmt.Sprintf("%s/entity/%s/metadata", c.baseURL, entityType),
                        Type:         entityType,
                        MediaType:    "application/json",
                },
        }
}