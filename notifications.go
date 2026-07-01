package airalo

import "context"

// NotificationType identifies a notification channel/event combination.
type NotificationType string

const (
	NotificationTypeWebhookLowData     NotificationType = "webhook_low_data"
	NotificationTypeEmailLowData       NotificationType = "email_low_data"
	NotificationTypeWebhookCreditLimit NotificationType = "webhook_credit_limit"
	NotificationTypeEmailCreditLimit   NotificationType = "email_credit_limit"
	NotificationTypeAsyncOrders        NotificationType = "async_orders"
)

// NotificationOptInParams configures POST /v2/notifications/opt-in.
//
// WebhookURL applies to the webhook_* types. ContactPoint/Language apply to
// the email_* types (contact_point is the destination email address).
// Levels (percentage thresholds, e.g. 50/70/80/90) applies to the
// credit_limit types only.
type NotificationOptInParams struct {
	Type         NotificationType `json:"type"`
	WebhookURL   string           `json:"webhook_url,omitempty"`
	ContactPoint string           `json:"contact_point,omitempty"`
	Language     string           `json:"language,omitempty"`
	Levels       []int            `json:"levels,omitempty"`
}

// NotificationInfo describes a configured notification's delivery details.
type NotificationInfo struct {
	Type         NotificationType `json:"type"`
	ContactPoint string           `json:"contact_point"`
	Language     string           `json:"language,omitempty"`
	Levels       []int            `json:"levels,omitempty"`
}

// NotificationDetails wraps a single configured notification.
type NotificationDetails struct {
	Notification NotificationInfo `json:"notification"`
}

// OptInNotification enables (or updates) delivery of a partner notification.
func (c *Client) OptInNotification(ctx context.Context, params NotificationOptInParams) (NotificationDetails, error) {
	return do[NotificationDetails](ctx, c, requestOptions{
		method:     "POST",
		path:       "/notifications/opt-in",
		jsonBody:   params,
		authorized: true,
	})
}

// GetNotificationDetails retrieves the currently configured notification.
//
// The source API documentation exposes this as GET /v2/notifications/opt-in
// with no query parameter to select which notification type to fetch, even
// though low_data/credit_limit/async_orders are configured independently.
// Inspect the returned NotificationInfo.Type to see which one came back; if
// the live API requires disambiguation this method may need a type param
// added once that's confirmed against a real account.
func (c *Client) GetNotificationDetails(ctx context.Context) (NotificationDetails, error) {
	return do[NotificationDetails](ctx, c, requestOptions{
		method:     "GET",
		path:       "/notifications/opt-in",
		authorized: true,
	})
}

// OptOutNotification disables a previously configured partner notification.
func (c *Client) OptOutNotification(ctx context.Context, notificationType NotificationType) error {
	_, err := do[map[string]any](ctx, c, requestOptions{
		method:     "POST",
		path:       "/notifications/opt-out",
		jsonBody:   map[string]any{"type": notificationType},
		authorized: true,
	})
	return err
}

// SimulateWebhookParams configures POST /v2/simulator/webhook, which triggers
// a test delivery of a webhook-based notification.
type SimulateWebhookParams struct {
	// Event identifies the notification event to simulate, e.g. "low_data_notification".
	Event string `json:"event"`
	// Type is an event-specific sub-type, e.g. "expire_1".
	Type string `json:"type"`
	// ICCID identifies the eSIM the simulated event relates to.
	ICCID string `json:"iccid"`
	// WebhookURL is where the simulated payload is delivered.
	WebhookURL string `json:"webhook_url"`
}

// WebhookSimulatorResult confirms a simulated webhook delivery attempt.
//
// Unlike most endpoints, this one does not use the standard {"data": ...}
// envelope; the response body is a bare {"success": "..."} object.
type WebhookSimulatorResult struct {
	Success string `json:"success"`
}

// SimulateWebhook triggers a test delivery of a webhook-based notification,
// useful for verifying your webhook receiver end-to-end.
func (c *Client) SimulateWebhook(ctx context.Context, params SimulateWebhookParams) (WebhookSimulatorResult, error) {
	return doRaw[WebhookSimulatorResult](ctx, c, requestOptions{
		method:     "POST",
		path:       "/simulator/webhook",
		jsonBody:   params,
		authorized: true,
	})
}
