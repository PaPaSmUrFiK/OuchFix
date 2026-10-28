package client

import (
	"context"
	"net/url"
	"strconv"

	"github.com/PaPaSmUrFiK/OuchFix/internal/payment/domain"
	httpclient "github.com/PaPaSmUrFiK/OuchFix/pkg/transport/http/client"
)

type Payment struct {
	http *httpclient.Client
}

func NewPayment(http *httpclient.Client) *Payment {
	return &Payment{http: http}
}

func (c *Payment) ListPayments(ctx context.Context, patientID *int64) ([]domain.Payment, error) {
	query := url.Values{}
	if patientID != nil {
		query.Set("patientId", strconv.FormatInt(*patientID, 10))
	}
	var out []domain.Payment
	return out, c.http.Get(ctx, "/payments", query, &out)
}
