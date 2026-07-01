package airalo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCreateFutureOrder_sendsJSONAndPreservesExtra(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/future-orders" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		if body["package_id"] != "change-7days-1gb" {
			t.Fatalf("package_id = %v", body["package_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"package_id":"change-7days-1gb","quantity":1,"due_date":"2025-04-09 10:00","status":"scheduled","some_unmodeled_field":true},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.CreateFutureOrder(context.Background(), CreateFutureOrderParams{
		Quantity:  1,
		PackageID: "change-7days-1gb",
		DueDate:   "2025-04-09 10:00",
	})
	if err != nil {
		t.Fatalf("CreateFutureOrder() error = %v", err)
	}
	if got.PackageID != "change-7days-1gb" || got.Status == nil || *got.Status != "scheduled" {
		t.Fatalf("unexpected result: %+v", got)
	}
	if _, ok := got.Extra["some_unmodeled_field"]; !ok {
		t.Fatalf("expected unmodeled field preserved, got %+v", got.Extra)
	}
}

func TestListFutureOrders_buildsQuery(t *testing.T) {
	var gotQuery string
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"package_id":"p","quantity":1,"due_date":"d","status":"pending"}],"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.ListFutureOrders(context.Background(), ListFutureOrdersParams{
		Status:      FutureOrderStatusPending,
		Limit:       10,
		FromDueDate: "2025-01-01 00:00",
		ToDueDate:   "2025-02-01 00:00",
	})
	if err != nil {
		t.Fatalf("ListFutureOrders() error = %v", err)
	}
	wantQuery := "from_due_date=2025-01-01+00%3A00&limit=10&status=pending&to_due_date=2025-02-01+00%3A00"
	if gotQuery != wantQuery {
		t.Fatalf("query = %q, want %q", gotQuery, wantQuery)
	}
	if len(got) != 1 || got[0].PackageID != "p" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestCancelFutureOrders_sendsRequestIDs(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/cancel-future-orders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var parsed struct {
			RequestIDs []string `json:"request_ids"`
		}
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("decoding body: %v", err)
		}
		if len(parsed.RequestIDs) != 2 {
			t.Fatalf("request_ids = %v", parsed.RequestIDs)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	err := c.CancelFutureOrders(context.Background(), []string{"a", "b"})
	if err != nil {
		t.Fatalf("CancelFutureOrders() error = %v", err)
	}
}
