package airalo

import (
	"context"
	"net/http"
	"testing"
)

func TestRequestRefund_sendsMultipartFields(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/refund" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got := r.FormValue("reason"); got != string(RefundReasonInstallationFailure) {
			t.Fatalf("reason = %q", got)
		}
		if got := r.Form["iccids[]"]; len(got) != 2 {
			t.Fatalf("iccids[] = %v", got)
		}
		if got := r.FormValue("email"); got != "email@example.com" {
			t.Fatalf("email = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	err := c.RequestRefund(context.Background(), RequestRefundParams{
		ICCIDs: []string{"894000000000001", "894000000000002"},
		Reason: RefundReasonInstallationFailure,
		Email:  "email@example.com",
	})
	if err != nil {
		t.Fatalf("RequestRefund() error = %v", err)
	}
}

func TestRequestRefund_requiresNotesForOthers(t *testing.T) {
	c, srv := newAuthorizedTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got := r.FormValue("notes"); got != "custom reason text" {
			t.Fatalf("notes = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	err := c.RequestRefund(context.Background(), RequestRefundParams{
		ICCIDs: []string{"894000000000001"},
		Reason: RefundReasonOthers,
		Notes:  "custom reason text",
	})
	if err != nil {
		t.Fatalf("RequestRefund() error = %v", err)
	}
}
