package airalo

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// SubmitOrderParams configures POST /v2/orders and POST /v2/orders-async.
type SubmitOrderParams struct {
	// Quantity is the number of eSIMs to order. Required, maximum 50.
	Quantity int
	// PackageID identifies the package to order, as returned by GetPackages. Required.
	PackageID string
	// Type is optional; the only supported value is "sim" and it's the server-side default.
	Type string
	// Description is an optional free-form label to identify the order later.
	Description string
	// BrandSettingsName selects the brand under which the eSIM is shared. Leave empty for unbranded.
	BrandSettingsName string
	// ToEmail, if set, sends an email with the eSIM sharing details. Requires SharingOption.
	ToEmail string
	// SharingOption is required when ToEmail is set. Valid values: "link", "pdf".
	SharingOption []string
	// CopyAddress optionally CCs additional recipients when ToEmail is set.
	CopyAddress []string
	// WebhookURL is only used by SubmitOrderAsync: a URL to receive the order
	// details asynchronously once fulfilled. Overrides any notification opt-in webhook.
	WebhookURL string
}

func (p SubmitOrderParams) toFormFields() map[string]string {
	fields := map[string]string{
		"quantity":   strconv.Itoa(p.Quantity),
		"package_id": p.PackageID,
	}
	if p.Type != "" {
		fields["type"] = p.Type
	}
	if p.Description != "" {
		fields["description"] = p.Description
	}
	if p.BrandSettingsName != "" {
		fields["brand_settings_name"] = p.BrandSettingsName
	}
	if p.ToEmail != "" {
		fields["to_email"] = p.ToEmail
	}
	if p.WebhookURL != "" {
		fields["webhook_url"] = p.WebhookURL
	}
	return fields
}

func (p SubmitOrderParams) toFormArrays() map[string][]string {
	arrays := map[string][]string{}
	if len(p.SharingOption) > 0 {
		arrays["sharing_option"] = p.SharingOption
	}
	if len(p.CopyAddress) > 0 {
		arrays["copy_address"] = p.CopyAddress
	}
	return arrays
}

// SubmitOrder places a synchronous order for one or more new eSIMs.
func (c *Client) SubmitOrder(ctx context.Context, params SubmitOrderParams) (Order, error) {
	return do[Order](ctx, c, requestOptions{
		method:     "POST",
		path:       "/orders",
		formFields: params.toFormFields(),
		formArrays: params.toFormArrays(),
		authorized: true,
	})
}

// AsyncOrderAccepted is returned when an order is accepted for asynchronous processing.
// The final order details are delivered to the configured webhook once fulfilled.
type AsyncOrderAccepted struct {
	RequestID  string `json:"request_id"`
	AcceptedAt string `json:"accepted_at"`
}

// SubmitOrderAsync places an order that is fulfilled asynchronously; the order
// details are delivered later to WebhookURL (or the notification opt-in
// webhook, if WebhookURL is empty).
func (c *Client) SubmitOrderAsync(ctx context.Context, params SubmitOrderParams) (AsyncOrderAccepted, error) {
	return do[AsyncOrderAccepted](ctx, c, requestOptions{
		method:     "POST",
		path:       "/orders-async",
		formFields: params.toFormFields(),
		formArrays: params.toFormArrays(),
		authorized: true,
	})
}

// OrderStatusFilter filters ListOrders by fulfillment status.
type OrderStatusFilter string

const (
	OrderStatusCompleted         OrderStatusFilter = "completed"
	OrderStatusFailed            OrderStatusFilter = "failed"
	OrderStatusPartiallyRefunded OrderStatusFilter = "partially_refunded"
	OrderStatusRefunded          OrderStatusFilter = "refunded"
)

// ListOrdersParams filters/paginates GET /v2/orders. All fields are optional.
type ListOrdersParams struct {
	// Include adds related data to each order. Valid values: "sims", "user", "status".
	Include []string
	// CreatedAtFrom and CreatedAtTo filter by creation date range (format "2006-01-02").
	// Both must be set to apply the filter.
	CreatedAtFrom string
	CreatedAtTo   string
	// Code performs a partial match against the order code.
	Code string
	// OrderStatus filters by fulfillment status.
	OrderStatus OrderStatusFilter
	// ICCID performs a partial match against the order's sim ICCID.
	ICCID string
	// Description performs a partial match against the order description.
	Description string
	// Limit sets how many orders are returned per page.
	Limit int
	// Page selects the pagination page (1-indexed).
	Page int
}

func (p ListOrdersParams) toQuery() url.Values {
	q := url.Values{}
	if len(p.Include) > 0 {
		q.Set("include", strings.Join(p.Include, ","))
	}
	if p.CreatedAtFrom != "" && p.CreatedAtTo != "" {
		q.Set("filter[created_at]", p.CreatedAtFrom+" - "+p.CreatedAtTo)
	}
	if p.Code != "" {
		q.Set("filter[code]", p.Code)
	}
	if p.OrderStatus != "" {
		q.Set("filter[order_status]", string(p.OrderStatus))
	}
	if p.ICCID != "" {
		q.Set("filter[iccid]", p.ICCID)
	}
	if p.Description != "" {
		q.Set("filter[description]", p.Description)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	return q
}

// ListOrders retrieves a paginated list of previously placed orders.
func (c *Client) ListOrders(ctx context.Context, params ListOrdersParams) (Page[[]Order], error) {
	return doPage[[]Order](ctx, c, requestOptions{
		method:     "GET",
		path:       "/orders",
		query:      params.toQuery(),
		authorized: true,
	})
}

// GetOrder retrieves a single order by id. include may contain "sims", "user", "status".
func (c *Client) GetOrder(ctx context.Context, orderID int, include []string) (Order, error) {
	q := url.Values{}
	if len(include) > 0 {
		q.Set("include", strings.Join(include, ","))
	}
	return do[Order](ctx, c, requestOptions{
		method:     "GET",
		path:       "/orders/" + strconv.Itoa(orderID),
		query:      q,
		authorized: true,
	})
}
