// Package redistoken provides a Redis-backed airalo.TokenStore, letting
// multiple replicas of a service share one cached Airalo access token
// instead of each fetching and caching their own.
package redistoken

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	airalo "github.com/airalo/airalo-go"
	"github.com/redis/go-redis/v9"
)

var (
	_ airalo.TokenStore         = (*Store)(nil)
	_ airalo.TokenRefreshLocker = (*Store)(nil)
)

const (
	defaultKeyPrefix = "airalo:token"
	defaultLockTTL   = 30 * time.Second
	lockRetryDelay   = 100 * time.Millisecond
)

// cachedToken is the JSON shape stored in Redis.
type cachedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Option configures a Store.
type Option func(*Store)

// WithKeyPrefix sets the Redis key (and lock key, suffixed ":lock") used to
// store the cached token. Defaults to "airalo:token". Set a distinct prefix
// per credential set if multiple Airalo clients share one Redis instance.
func WithKeyPrefix(prefix string) Option {
	return func(s *Store) { s.keyPrefix = prefix }
}

// WithLockTTL sets how long the refresh lock is held before it
// automatically expires, as a safety net if a holder crashes mid-refresh.
// Defaults to 30s.
func WithLockTTL(ttl time.Duration) Option {
	return func(s *Store) { s.lockTTL = ttl }
}

// Store is a Redis-backed airalo.TokenStore that also implements
// airalo.TokenRefreshLocker to coordinate refreshes across processes.
type Store struct {
	rdb       redis.Cmdable
	keyPrefix string
	lockTTL   time.Duration
}

// New builds a Store backed by the given Redis client. client may be a
// *redis.Client, *redis.Ring, *redis.ClusterClient, or anything else
// implementing redis.Cmdable.
func New(client redis.Cmdable, opts ...Option) *Store {
	s := &Store{
		rdb:       client,
		keyPrefix: defaultKeyPrefix,
		lockTTL:   defaultLockTTL,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Store) lockKey() string {
	return s.keyPrefix + ":lock"
}

// Get implements airalo.TokenStore.
func (s *Store) Get(ctx context.Context) (string, time.Time, bool, error) {
	raw, err := s.rdb.Get(ctx, s.keyPrefix).Result()
	if errors.Is(err, redis.Nil) {
		return "", time.Time{}, false, nil
	}
	if err != nil {
		return "", time.Time{}, false, err
	}

	var ct cachedToken
	if err := json.Unmarshal([]byte(raw), &ct); err != nil {
		return "", time.Time{}, false, err
	}
	if ct.Token == "" || !time.Now().Before(ct.ExpiresAt) {
		return "", time.Time{}, false, nil
	}
	return ct.Token, ct.ExpiresAt, true, nil
}

// Set implements airalo.TokenStore.
func (s *Store) Set(ctx context.Context, token string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Already expired; nothing useful to cache, but don't error the
		// caller for it.
		return nil
	}

	b, err := json.Marshal(cachedToken{Token: token, ExpiresAt: expiresAt})
	if err != nil {
		return err
	}
	return s.rdb.Set(ctx, s.keyPrefix, b, ttl).Err()
}

// LockRefresh implements airalo.TokenRefreshLocker, coordinating token
// refreshes across processes sharing this Redis instance so only one of
// them fetches a new token from Airalo at a time.
func (s *Store) LockRefresh(ctx context.Context) (func(), error) {
	token, err := randomToken()
	if err != nil {
		return nil, err
	}

	for {
		ok, err := s.rdb.SetNX(ctx, s.lockKey(), token, s.lockTTL).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			release := func() {
				// Best-effort: run with a fresh context in case the caller's
				// ctx was already canceled/expired.
				releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				s.unlock(releaseCtx, token)
			}
			return release, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(lockRetryDelay):
		}
	}
}

// unlock releases the refresh lock, but only if it still holds the value we
// set, so a holder never deletes a lock it no longer owns (e.g. after its
// own TTL expired and another replica acquired it in the meantime). The
// GET-then-DEL isn't atomic, so there's a narrow window where this could
// delete a lock acquired by another replica right after we checked it; the
// lock TTL bounds the worst case and a stray extra token fetch is harmless.
func (s *Store) unlock(ctx context.Context, token string) {
	key := s.lockKey()
	got, err := s.rdb.Get(ctx, key).Result()
	if err != nil || got != token {
		return
	}
	s.rdb.Del(ctx, key)
}

func randomToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
