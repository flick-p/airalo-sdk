package airalo

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"
)

// fakeStore is an in-memory TokenStore shared by pointer, standing in for a
// distributed store like Redis: two tokenSources pointed at the same
// *fakeStore behave like two replicas sharing a cache.
type fakeStore struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func (f *fakeStore) Get(context.Context) (string, time.Time, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.token == "" || !time.Now().Before(f.expiresAt) {
		return "", time.Time{}, false, nil
	}
	return f.token, f.expiresAt, true, nil
}

func (f *fakeStore) Set(_ context.Context, token string, expiresAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.token = token
	f.expiresAt = expiresAt
	return nil
}

// fakeLocker wraps fakeStore and records LockRefresh/release calls.
type fakeLocker struct {
	fakeStore
	locked   int
	released int
}

func (f *fakeLocker) LockRefresh(context.Context) (func(), error) {
	f.mu.Lock()
	f.locked++
	f.mu.Unlock()
	return func() {
		f.mu.Lock()
		f.released++
		f.mu.Unlock()
	}, nil
}

func TestTokenSource_sharedStoreActsLikeMultipleReplicas(t *testing.T) {
	calls := 0
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"token_type":"Bearer","expires_in":3600,"access_token":"shared-tok"},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	store := &fakeStore{}
	replicaA := newTokenSource(c, "id", "secret", store)
	replicaB := newTokenSource(c, "id", "secret", store)

	tok, err := replicaA.Token(context.Background())
	if err != nil {
		t.Fatalf("replicaA.Token() error = %v", err)
	}
	if tok != "shared-tok" {
		t.Fatalf("replicaA.Token() = %q, want shared-tok", tok)
	}

	// replicaB shares the same store, so it should see the cached token
	// without hitting the token endpoint again.
	tok, err = replicaB.Token(context.Background())
	if err != nil {
		t.Fatalf("replicaB.Token() error = %v", err)
	}
	if tok != "shared-tok" {
		t.Fatalf("replicaB.Token() = %q, want shared-tok", tok)
	}
	if calls != 1 {
		t.Fatalf("token endpoint called %d times, want 1 (expected shared cache)", calls)
	}
}

func TestTokenSource_locksAcrossRefresh(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"token_type":"Bearer","expires_in":3600,"access_token":"locked-tok"},"meta":{"message":"success"}}`))
	})
	defer srv.Close()

	locker := &fakeLocker{}
	ts := newTokenSource(c, "id", "secret", locker)

	tok, err := ts.Token(context.Background())
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if tok != "locked-tok" {
		t.Fatalf("Token() = %q, want locked-tok", tok)
	}
	if locker.locked != 1 || locker.released != 1 {
		t.Fatalf("locked = %d, released = %d, want 1, 1", locker.locked, locker.released)
	}

	// Second call is served from cache; the lock should not be re-acquired.
	if _, err := ts.Token(context.Background()); err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if locker.locked != 1 {
		t.Fatalf("locked = %d after cached call, want 1", locker.locked)
	}
}
