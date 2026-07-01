package airalo

import (
	"context"
	"net/http"
	"testing"
)

func TestGetBalance(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/balance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"balances":{"name":"balance","availableBalance":{"amount":0,"currency":"USD"}}},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	got, err := c.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if got.Balances.AvailableBalance.Currency != "USD" {
		t.Fatalf("unexpected result: %+v", got)
	}
}
