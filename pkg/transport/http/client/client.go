// Package httpclient — клиент внутреннего REST API сервисов для gateway.
// В ЛР6 заменится на gRPC.
package httpclient

import (
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
	target := c.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fmt.Errorf("build request %s: %w", path, err)
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
	default:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("call %s: status %d: %s", path, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || body.Error == "" {
		return kind
	}
	return &remoteError{msg: body.Error, kind: kind}
}
