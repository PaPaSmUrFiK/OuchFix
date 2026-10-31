// Package httpclient — HTTP-прокси gateway к внутреннему REST API сервиса:
// знает только базовый адрес сервиса и декодирует JSON-ответ как есть.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	pkgerrors "github.com/PaPaSmUrFiK/OuchFix/pkg/errors"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

// remoteError сохраняет текст ошибки сервиса как есть и привязывает его к
// доменной ошибке, чтобы gateway отдал тот же код и сообщение, что и раньше.
type remoteError struct {
	msg  string
	kind error
}

func (e *remoteError) Error() string { return e.msg }
func (e *remoteError) Unwrap() error { return e.kind }

// Get делает GET path?query и декодирует JSON-ответ в out.
func (c *Client) Get(ctx context.Context, path string, query url.Values, out any) error {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	return c.do(ctx, http.MethodGet, path, nil, out)
}

// Post отправляет body как JSON и декодирует JSON-ответ в out.
func (c *Client) Post(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode %s: %w", path, err)
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("build request %s: %w", path, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
		return nil
	}

	var kind error
	switch resp.StatusCode {
	case http.StatusNotFound:
		kind = pkgerrors.ErrNotFound
	case http.StatusBadRequest:
		kind = pkgerrors.ErrInvalidArgument
	case http.StatusNotImplemented:
		kind = pkgerrors.ErrNotImplemented
	case http.StatusConflict:
		kind = pkgerrors.ErrConflict
	case http.StatusUnprocessableEntity:
		kind = pkgerrors.ErrFailedPrecondition
	default:
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("call %s: status %d: %s", path, resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var errBody struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil || errBody.Error == "" {
		return kind
	}
	return &remoteError{msg: errBody.Error, kind: kind}
}
