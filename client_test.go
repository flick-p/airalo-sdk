package airalo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := NewClient(Config{
		ClientID:     "id",
		ClientSecret: "secret",
		BaseURL:      srv.URL,
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	return c, srv
}

// newAuthorizedTestClient wraps handler with automatic handling of the
// POST /v2/token exchange, so resource-level tests can focus on the
// endpoint under test without also stubbing authentication.
func newAuthorizedTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	return newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/token" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":{"token_type":"Bearer","expires_in":3600,"access_token":"test-token"},"meta":{"message":"success"}}`))
			return
		}
		handler(w, r)
	})
}

func TestNewClient_requiresCredentials(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
	if _, err := NewClient(Config{ClientID: "id"}); err == nil {
		t.Fatal("expected error for missing client secret, got nil")
	}
}

func TestTokenSource_fetchesAndCaches(t *testing.T) {
	calls := 0
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.FormValue("grant_type"); got != "client_credentials" {
			t.Fatalf("grant_type = %q, want client_credentials", got)
		}
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"token_type":"Bearer","expires_in":3600,"access_token":"tok-123"},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	tok, err := c.auth.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if tok != "tok-123" {
		t.Fatalf("Token() = %q, want tok-123", tok)
	}

	// Second call should be served from cache, not hit the server again.
	if _, err := c.auth.Token(context.Background()); err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if calls != 1 {
		t.Fatalf("token endpoint called %d times, want 1 (expected caching)", calls)
	}
}

func TestTokenSource_refetchesWhenExpired(t *testing.T) {
	calls := 0
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		// expires_in shorter than the refresh buffer forces a re-fetch every call.
		w.Write([]byte(`{"data":{"token_type":"Bearer","expires_in":1,"access_token":"tok-123"},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	if _, err := c.auth.Token(context.Background()); err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if _, err := c.auth.Token(context.Background()); err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if calls != 2 {
		t.Fatalf("token endpoint called %d times, want 2 (expected refresh)", calls)
	}
}

func TestDo_decodesEnvelope(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"name":"hi"},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	type payload struct {
		Name string `json:"name"`
	}
	got, err := do[payload](context.Background(), c, requestOptions{method: "GET", path: "/whatever"})
	if err != nil {
		t.Fatalf("do() error = %v", err)
	}
	if got.Name != "hi" {
		t.Fatalf("got.Name = %q, want hi", got.Name)
	}
}

func TestDo_returnsAPIErrorWithFields(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"data":{"package_id":"The selected package is invalid."},"meta":{"message":"the parameter is invalid"}}`))
	})
	defer srv.Close()

	_, err := do[struct{}](context.Background(), c, requestOptions{method: "GET", path: "/whatever"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != 422 {
		t.Fatalf("StatusCode = %d, want 422", apiErr.StatusCode)
	}
	if apiErr.Fields["package_id"] != "The selected package is invalid." {
		t.Fatalf("Fields[package_id] = %q, unexpected", apiErr.Fields["package_id"])
	}
}

func TestDo_handlesEmptyBody(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	defer srv.Close()

	_, err := do[struct{}](context.Background(), c, requestOptions{method: "POST", path: "/whatever"})
	if err != nil {
		t.Fatalf("do() error = %v", err)
	}
}

func TestNewRequest_authorizedAttachesBearerToken(t *testing.T) {
	var gotAuth string
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/token" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data":{"token_type":"Bearer","expires_in":3600,"access_token":"tok-abc"},"meta":{"message":"success"}}`))
			return
		}
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	_, err := do[map[string]any](context.Background(), c, requestOptions{method: "GET", path: "/balance", authorized: true})
	if err != nil {
		t.Fatalf("do() error = %v", err)
	}
	if gotAuth != "Bearer tok-abc" {
		t.Fatalf("Authorization header = %q, want %q", gotAuth, "Bearer tok-abc")
	}
}
