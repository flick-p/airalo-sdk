package airalo

import (
	"context"
	"net/http"
	"testing"
)

func TestListCompatibleDevices(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/compatible-devices" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"model":"zangya_sprout","os":"android","brand":"bq","name":"Aquaris X2"}]}`))
	})
	defer srv.Close()

	got, err := c.ListCompatibleDevices(context.Background())
	if err != nil {
		t.Fatalf("ListCompatibleDevices() error = %v", err)
	}
	if len(got) != 1 || got[0].Brand != "bq" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestListCompatibleDevicesLite(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/compatible-devices-lite" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":[{"model":"bramble","os":"android","brand":"Google","name":"Pixel 4a (5G)"}]}`))
	})
	defer srv.Close()

	got, err := c.ListCompatibleDevicesLite(context.Background())
	if err != nil {
		t.Fatalf("ListCompatibleDevicesLite() error = %v", err)
	}
	if len(got) != 1 || got[0].Name != "Pixel 4a (5G)" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
