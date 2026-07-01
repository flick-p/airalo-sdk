package airalo

import (
	"context"
	"net/http"
	"testing"
)

func TestListTopupPackages(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/sims/873000000000042542/topups" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":"bonbon-mobile-30days-3gb-topup","type":"topup","price":10,"amount":3072,"day":30,"is_unlimited":false,"title":"3 GB - 30 Days","data":"3 GB","short_info":"x","voice":100,"text":100}]}`))
	})
	defer srv.Close()

	got, err := c.ListTopupPackages(context.Background(), "873000000000042542")
	if err != nil {
		t.Fatalf("ListTopupPackages() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "bonbon-mobile-30days-3gb-topup" {
		t.Fatalf("unexpected result: %+v", got)
	}
	if got[0].Voice == nil || *got[0].Voice != 100 {
		t.Fatalf("unexpected voice: %+v", got[0].Voice)
	}
}

func TestListTopupPackages_empty(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[]}`))
	})
	defer srv.Close()

	got, err := c.ListTopupPackages(context.Background(), "none")
	if err != nil {
		t.Fatalf("ListTopupPackages() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %+v", got)
	}
}

func TestSubmitTopupOrder(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/orders/topups" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got := r.FormValue("package_id"); got != "change-7days-1gb-topup" {
			t.Fatalf("package_id = %q", got)
		}
		if got := r.FormValue("iccid"); got != "873000000000042542" {
			t.Fatalf("iccid = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"id":174,"code":"20230207-000174","created_at":"t","type":"topup","package_id":"bonbon-mobile-30days-3gb-topup","quantity":"1","package":"P","esim_type":"Prepaid","validity":30,"price":10,"data":"3 GB","currency":"USD","manual_installation":"","qrcode_installation":"","installation_guides":{}},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	order, err := c.SubmitTopupOrder(context.Background(), SubmitTopupOrderParams{
		PackageID: "change-7days-1gb-topup",
		ICCID:     "873000000000042542",
	})
	if err != nil {
		t.Fatalf("SubmitTopupOrder() error = %v", err)
	}
	if order.ID != 174 || order.Type != "topup" {
		t.Fatalf("unexpected order: %+v", order)
	}
}

func TestSubmitTopupOrder_purchaseLimitExceeded(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"data":{"package_id":"validation.order_exceed_limit"},"meta":{"message":"the parameter is invalid"}}`))
	})
	defer srv.Close()

	_, err := c.SubmitTopupOrder(context.Background(), SubmitTopupOrderParams{PackageID: "p", ICCID: "i"})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 422 {
		t.Fatalf("unexpected error: %v", err)
	}
}
