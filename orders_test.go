package airalo

import (
	"context"
	"io"
	"mime"
	"net/http"
	"testing"
)

func TestSubmitOrder_sendsMultipartAndDecodesOrder(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/orders" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mt != "multipart/form-data" {
			t.Fatalf("unexpected content type: %v %v", mt, err)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got := r.FormValue("package_id"); got != "kallur-digital-7days-1gb" {
			t.Fatalf("package_id = %q", got)
		}
		if got := r.FormValue("quantity"); got != "1" {
			t.Fatalf("quantity = %q", got)
		}
		if got := r.Form["sharing_option[]"]; len(got) != 1 || got[0] != "link" {
			t.Fatalf("sharing_option[] = %v", got)
		}
		_ = params
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"id": 9666, "code": "20230227-009666", "created_at": "2023-02-27 14:09:55",
				"type": "sim", "package_id": "kallur-digital-7days-1gb", "quantity": "1",
				"package": "Kallur Digital-1 GB - 7 Days", "esim_type": "Prepaid", "validity": 7,
				"price": 9.5, "data": "1 GB", "currency": "USD",
				"manual_installation": "", "qrcode_installation": "",
				"installation_guides": {"en": "https://sandbox.airalo.com/installation-guide"},
				"brand_settings_name": "our perfect brand",
				"sims": [{"id": 11047, "created_at": "2023-02-27 14:09:55", "iccid": "891000000000009125",
					"lpa": "lpa.airalo.com", "imsis": null, "matching_id": "TEST", "qrcode": "LPA:1$lpa.airalo.com$TEST",
					"qrcode_url": "https://x", "direct_apple_installation_url": "https://y",
					"airalo_code": null, "apn_type": "automatic", "apn_value": null, "is_roaming": true,
					"confirmation_code": null}]
			},
			"meta": {"message": "success"}
		}`))
	})
	defer srv.Close()

	order, err := c.SubmitOrder(context.Background(), SubmitOrderParams{
		Quantity:      1,
		PackageID:     "kallur-digital-7days-1gb",
		ToEmail:       "valid@example.com",
		SharingOption: []string{"link"},
	})
	if err != nil {
		t.Fatalf("SubmitOrder() error = %v", err)
	}
	if order.ID != 9666 || order.Quantity != 1 || order.Price != 9.5 {
		t.Fatalf("unexpected order: %+v", order)
	}
	if len(order.Sims) != 1 || order.Sims[0].ICCID != "891000000000009125" {
		t.Fatalf("unexpected sims: %+v", order.Sims)
	}
}

func TestSubmitOrder_validationError(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"data":{"package_id":"The selected package is invalid.","quantity":"The quantity may not be greater than 50."},"meta":{"message":"the parameter is invalid"}}`))
	})
	defer srv.Close()

	_, err := c.SubmitOrder(context.Background(), SubmitOrderParams{Quantity: 51, PackageID: "bogus"})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 422 {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiErr.Fields["package_id"] == "" {
		t.Fatalf("expected package_id field error, got %+v", apiErr.Fields)
	}
}

func TestSubmitOrderAsync_decodesAcceptance(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/orders-async" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"data":{"request_id":"3NhR3gKmqCWK7IWppurpDX3Cg","accepted_at":"2024-07-11 15:26:02"},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.SubmitOrderAsync(context.Background(), SubmitOrderParams{Quantity: 1, PackageID: "x", WebhookURL: "https://hook"})
	if err != nil {
		t.Fatalf("SubmitOrderAsync() error = %v", err)
	}
	if got.RequestID != "3NhR3gKmqCWK7IWppurpDX3Cg" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestListOrders_buildsFiltersAndDecodesPage(t *testing.T) {
	var gotQuery string
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"id":1,"code":"c","created_at":"t","type":"sim","package_id":"p","quantity":1,"package":"P","esim_type":"Prepaid","validity":"30","price":"22.50","data":"1GB","currency":"USD","manual_installation":"","qrcode_installation":"","installation_guides":{}}],"links":{"first":"","last":"","prev":null,"next":null},"meta":{"message":"success","current_page":1,"from":1,"last_page":1,"path":"","per_page":"50","to":1,"total":1}}`))
	})
	defer srv.Close()

	page, err := c.ListOrders(context.Background(), ListOrdersParams{
		Include:       []string{"sims", "user"},
		CreatedAtFrom: "2023-01-01",
		CreatedAtTo:   "2023-01-31",
		OrderStatus:   OrderStatusCompleted,
		Limit:         50,
		Page:          1,
	})
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}
	wantQuery := "filter%5Bcreated_at%5D=2023-01-01+-+2023-01-31&filter%5Border_status%5D=completed&include=sims%2Cuser&limit=50&page=1"
	if gotQuery != wantQuery {
		t.Fatalf("query = %q, want %q", gotQuery, wantQuery)
	}
	if len(page.Data) != 1 || page.Data[0].Price != 22.5 || page.Data[0].Validity != 30 {
		t.Fatalf("unexpected page data: %+v", page.Data)
	}
}

func TestGetOrder_notFoundUnauthorized(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/orders/9565" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"data":[],"meta":{"message":"This action is unauthorized."}}`))
	})
	defer srv.Close()

	_, err := c.GetOrder(context.Background(), 9565, []string{"sims", "user", "status"})
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 401 {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubmitESimVoucher(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/voucher/esim" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Fatal("expected non-empty JSON body")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"package_id":"jaco-mobile-7days-1gb","codes":["BIXLAAAA","BSXLAAAA"],"booking_reference":"123"}],"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.SubmitESimVoucher(context.Background(), SubmitESimVoucherParams{
		Vouchers: []ESimVoucherItem{{PackageID: "jaco-mobile-7days-1gb", Codes: []string{"BIXLAAAA", "BSXLAAAA"}, BookingReference: "123"}},
	})
	if err != nil {
		t.Fatalf("SubmitESimVoucher() error = %v", err)
	}
	if len(got) != 1 || got[0].PackageID != "jaco-mobile-7days-1gb" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
