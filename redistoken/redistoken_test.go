package redistoken

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestStore(t *testing.T) (*Store, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return New(rdb), rdb
}

func TestStore_getMissReturnsNotOK(t *testing.T) {
	s, _ := newTestStore(t)
	_, _, ok, err := s.Get(context.Background())
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if ok {
		t.Fatal("Get() ok = true on empty store, want false")
	}
}

func TestStore_setThenGetRoundTrips(t *testing.T) {
	s, _ := newTestStore(t)
	ctx := context.Background()
	want := time.Now().Add(time.Hour).Truncate(time.Millisecond)

	if err := s.Set(ctx, "tok-123", want); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	tok, expiresAt, ok, err := s.Get(ctx)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if tok != "tok-123" {
		t.Fatalf("Get() token = %q, want tok-123", tok)
	}
	if !expiresAt.Equal(want) {
		t.Fatalf("Get() expiresAt = %v, want %v", expiresAt, want)
	}
}

func TestStore_setWithPastExpiryIsNotCached(t *testing.T) {
	s, _ := newTestStore(t)
	ctx := context.Background()

	if err := s.Set(ctx, "stale", time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if _, _, ok, err := s.Get(ctx); err != nil {
		t.Fatalf("Get() error = %v", err)
	} else if ok {
		t.Fatal("Get() ok = true for an already-expired token, want false")
	}
}

func TestStore_keyPrefixIsolatesCredentials(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	a := New(rdb, WithKeyPrefix("airalo:prod"))
	b := New(rdb, WithKeyPrefix("airalo:sandbox"))
	ctx := context.Background()

	if err := a.Set(ctx, "prod-tok", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("a.Set() error = %v", err)
	}
	if _, _, ok, err := b.Get(ctx); err != nil {
		t.Fatalf("b.Get() error = %v", err)
	} else if ok {
		t.Fatal("b.Get() ok = true, want false (distinct prefixes must not collide)")
	}
}

func TestStore_lockRefreshSerializesConcurrentReplicas(t *testing.T) {
	mr := miniredis.RunT(t)
	ctx := context.Background()

	rdbA := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdbB := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdbA.Close(); rdbB.Close() })

	replicaA := New(rdbA, WithLockTTL(2*time.Second))
	replicaB := New(rdbB, WithLockTTL(2*time.Second))

	releaseA, err := replicaA.LockRefresh(ctx)
	if err != nil {
		t.Fatalf("replicaA.LockRefresh() error = %v", err)
	}

	var (
		mu        sync.Mutex
		bAcquired bool
	)
	done := make(chan struct{})
	go func() {
		defer close(done)
		releaseB, err := replicaB.LockRefresh(ctx)
		if err != nil {
			t.Errorf("replicaB.LockRefresh() error = %v", err)
			return
		}
		mu.Lock()
		bAcquired = true
		mu.Unlock()
		releaseB()
	}()

	// Give replicaB a moment to attempt and block on the lock replicaA holds.
	time.Sleep(150 * time.Millisecond)
	mu.Lock()
	gotEarly := bAcquired
	mu.Unlock()
	if gotEarly {
		t.Fatal("replicaB acquired the lock while replicaA still held it")
	}

	releaseA()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("replicaB never acquired the lock after replicaA released it")
	}

	mu.Lock()
	defer mu.Unlock()
	if !bAcquired {
		t.Fatal("replicaB did not acquire the lock")
	}
}
