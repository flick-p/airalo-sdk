package airalo

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestOptInNotification_webhookLowData(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/notifications/opt-in" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["type"] != "webhook_low_data" || body["webhook_url"] != "https://example.com" {
			t.Fatalf("unexpected body: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"notification":{"type":"webhook_low_data","contact_point":"https://example.com"}},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.OptInNotification(context.Background(), NotificationOptInParams{
		Type:       NotificationTypeWebhookLowData,
		WebhookURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("OptInNotification() error = %v", err)
	}
	if got.Notification.Type != NotificationTypeWebhookLowData || got.Notification.ContactPoint != "https://example.com" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestOptInNotification_webhookCreditLimitWithLevels(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		levels, _ := body["levels"].([]any)
		if len(levels) != 4 {
			t.Fatalf("levels = %v", body["levels"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"notification":{"type":"webhook_credit_limit","contact_point":"https://example.com","levels":[50,70,80,90]}},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.OptInNotification(context.Background(), NotificationOptInParams{
		Type:       NotificationTypeWebhookCreditLimit,
		WebhookURL: "https://example.com",
		Levels:     []int{50, 70, 80, 90},
	})
	if err != nil {
		t.Fatalf("OptInNotification() error = %v", err)
	}
	if len(got.Notification.Levels) != 4 {
		t.Fatalf("unexpected levels: %+v", got.Notification.Levels)
	}
}

func TestGetNotificationDetails(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/notifications/opt-in" || r.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"notification":{"type":"async_orders","contact_point":"https://example.com"}},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.GetNotificationDetails(context.Background())
	if err != nil {
		t.Fatalf("GetNotificationDetails() error = %v", err)
	}
	if got.Notification.Type != NotificationTypeAsyncOrders {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestOptOutNotification(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/notifications/opt-out" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["type"] != "webhook_low_data" {
			t.Fatalf("unexpected body: %v", body)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	if err := c.OptOutNotification(context.Background(), NotificationTypeWebhookLowData); err != nil {
		t.Fatalf("OptOutNotification() error = %v", err)
	}
}

func TestSimulateWebhook_nonEnvelopedResponse(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/simulator/webhook" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body SimulateWebhookParams
		json.NewDecoder(r.Body).Decode(&body)
		if body.Event != "low_data_notification" {
			t.Fatalf("unexpected body: %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":"Notification sent"}`))
	})
	defer srv.Close()

	got, err := c.SimulateWebhook(context.Background(), SimulateWebhookParams{
		Event:      "low_data_notification",
		Type:       "expire_1",
		ICCID:      "8997212330099025334",
		WebhookURL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("SimulateWebhook() error = %v", err)
	}
	if got.Success != "Notification sent" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
