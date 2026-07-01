package airalo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// FutureOrderStatus filters ListFutureOrders by processing status.
type FutureOrderStatus string

const (
	FutureOrderStatusPending FutureOrderStatus = "pending"
	FutureOrderStatusFailed  FutureOrderStatus = "failed"
	FutureOrderStatusRetry   FutureOrderStatus = "retry"
)

// CreateFutureOrderParams configures POST /v2/future-orders.
type CreateFutureOrderParams struct {
	// Quantity is the number of eSIMs to order.
	Quantity int `json:"quantity"`
	// PackageID identifies the package to order.
	PackageID string `json:"package_id"`
	// DueDate is when the order should be placed, formatted "2006-01-02 15:04".
	DueDate string `json:"due_date"`
	// Description is an optional free-form label to identify the order later.
	Description string `json:"description,omitempty"`
	// WebhookURL, if set, receives the fulfilled order details once the due date is reached.
	WebhookURL string `json:"webhook_url,omitempty"`
}

var futureOrderKnownFields = []string{
	"package_id", "quantity", "due_date", "description", "webhook_url", "status", "request_id",
}

// FutureOrder is a scheduled order to be fulfilled at a future due date.
//
// The source API documentation did not include a worked response example for
// the create/list future-order endpoints, so only the fields echoed back
// from the request plus commonly expected bookkeeping fields are typed;
// anything else the API returns is preserved verbatim in Extra.
type FutureOrder struct {
	PackageID   string  `json:"package_id"`
	Quantity    FlexInt `json:"quantity"`
	DueDate     string  `json:"due_date"`
	Description *string `json:"description"`
	WebhookURL  *string `json:"webhook_url"`
	Status      *string `json:"status"`
	RequestID   *string `json:"request_id"`

	Extra map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON decodes known fields and preserves the rest in Extra.
func (f *FutureOrder) UnmarshalJSON(b []byte) error {
	type alias FutureOrder
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return fmt.Errorf("airalo: decoding FutureOrder: %w", err)
	}
	*f = FutureOrder(a)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("airalo: decoding FutureOrder: %w", err)
	}
	for _, known := range futureOrderKnownFields {
		delete(raw, known)
	}
	if len(raw) > 0 {
		f.Extra = raw
	}
	return nil
}

// CreateFutureOrder schedules an order to be placed automatically at a future due date.
func (c *Client) CreateFutureOrder(ctx context.Context, params CreateFutureOrderParams) (FutureOrder, error) {
	return do[FutureOrder](ctx, c, requestOptions{
		method:     "POST",
		path:       "/future-orders",
		jsonBody:   params,
		authorized: true,
	})
}

// ListFutureOrdersParams filters GET /v2/future-orders. All fields are optional.
type ListFutureOrdersParams struct {
	// Status filters by processing status. Defaults server-side to "pending".
	Status FutureOrderStatus
	// Limit caps the number of results returned.
	Limit int
	// FromDueDate and ToDueDate filter by due date range, formatted "2006-01-02 15:04".
	FromDueDate string
	ToDueDate   string
}

func (p ListFutureOrdersParams) toQuery() url.Values {
	q := url.Values{}
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.FromDueDate != "" {
		q.Set("from_due_date", p.FromDueDate)
	}
	if p.ToDueDate != "" {
		q.Set("to_due_date", p.ToDueDate)
	}
	return q
}

// ListFutureOrders retrieves scheduled future orders.
func (c *Client) ListFutureOrders(ctx context.Context, params ListFutureOrdersParams) ([]FutureOrder, error) {
	return do[[]FutureOrder](ctx, c, requestOptions{
		method:     "GET",
		path:       "/future-orders",
		query:      params.toQuery(),
		authorized: true,
	})
}

// CancelFutureOrders cancels one or more previously scheduled future orders by request id.
func (c *Client) CancelFutureOrders(ctx context.Context, requestIDs []string) error {
	_, err := do[json.RawMessage](ctx, c, requestOptions{
		method: "POST",
		path:   "/cancel-future-orders",
		jsonBody: map[string]any{
			"request_ids": requestIDs,
		},
		authorized: true,
	})
	return err
}
