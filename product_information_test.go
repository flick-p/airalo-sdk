package airalo

import (
	"context"
	"net/http"
	"testing"
)

func TestGetProductInformation_decodesKnownAndExtraFields(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/packages/change-7days-1gb/product-information" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"package_slug": "change-7days-1gb",
				"pib_version": "1",
				"service_internet": true,
				"data_allowance_mb": 1024,
				"validity_days": 7,
				"is_unlimited": false,
				"network_providers": ["T-Mobile", "Verizon"],
				"retail_price_eur": 4.1,
				"partner_brand_name": "Change",
				"auto_renewal": false,
				"some_new_field_not_yet_modeled": "surprise"
			},
			"meta": {"message": "success"}
		}`))
	})
	defer srv.Close()

	got, err := c.GetProductInformation(context.Background(), "change-7days-1gb")
	if err != nil {
		t.Fatalf("GetProductInformation() error = %v", err)
	}
	if got.PackageSlug != "change-7days-1gb" || !got.ServiceInternet {
		t.Fatalf("unexpected decode: %+v", got)
	}
	if got.DataAllowanceMB == nil || *got.DataAllowanceMB != 1024 {
		t.Fatalf("DataAllowanceMB = %v, want 1024", got.DataAllowanceMB)
	}
	if len(got.NetworkProviders) != 2 {
		t.Fatalf("NetworkProviders = %v", got.NetworkProviders)
	}
	if _, ok := got.Extra["some_new_field_not_yet_modeled"]; !ok {
		t.Fatalf("expected unmodeled field preserved in Extra, got %+v", got.Extra)
	}
}

func TestGetProductInformation_notFound(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"data":[],"meta":{"message":"messages.resource_not_found"}}`))
	})
	defer srv.Close()

	_, err := c.GetProductInformation(context.Background(), "unknown-slug")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 404 {
		t.Fatalf("unexpected error: %v", err)
	}
}
