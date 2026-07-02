package airalo

import (
	"context"
	"net/url"
	"sync"
	"time"
)

// refreshBuffer is how long before the token's reported expiry we proactively
// refresh it, to avoid racing an in-flight request against expiration.
const refreshBuffer = 60 * time.Second

// TokenResponse is the payload returned by POST /v2/token.
type TokenResponse struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

// TokenStore persists a cached OAuth2 access token. The default store used
// when Config.TokenStore is nil keeps the token in memory for the lifetime
// of the Client. Supplying a shared store (e.g. the redistoken package)
// lets multiple Client instances/replicas reuse the same cached token
// instead of each fetching their own.
type TokenStore interface {
	// Get returns the cached token and its expiry. ok is false if there is
	// no cached token or a store-level error occurred.
	Get(ctx context.Context) (token string, expiresAt time.Time, ok bool, err error)
	// Set stores the token and the time at which it should be considered
	// expired (and re-fetched).
	Set(ctx context.Context, token string, expiresAt time.Time) error
}

// TokenRefreshLocker is an optional interface a TokenStore may implement to
// coordinate token refreshes across processes sharing the same store, so
// that when the cached token expires, only one process fetches a new one
// from Airalo while the others wait and then re-read the store.
type TokenRefreshLocker interface {
	// LockRefresh blocks until the refresh lock is acquired or ctx is done,
	// and returns a function that releases it.
	LockRefresh(ctx context.Context) (release func(), err error)
}

// memoryTokenStore is the default TokenStore: an in-memory cache scoped to
// a single tokenSource/Client.
type memoryTokenStore struct {
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func (m *memoryTokenStore) Get(context.Context) (string, time.Time, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.token == "" || !time.Now().Before(m.expiresAt) {
		return "", time.Time{}, false, nil
	}
	return m.token, m.expiresAt, true, nil
}

func (m *memoryTokenStore) Set(_ context.Context, token string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = token
	m.expiresAt = expiresAt
	return nil
}

// tokenSource fetches and caches an OAuth2 client_credentials access token
// via a TokenStore, refreshing it automatically as it nears expiry. Safe for
// concurrent use.
type tokenSource struct {
	client       *Client
	clientID     string
	clientSecret string
	store        TokenStore

	// mu serializes refreshes within this process; it complements (but does
	// not replace) any cross-process coordination the store provides via
	// TokenRefreshLocker.
	mu sync.Mutex
}

func newTokenSource(c *Client, clientID, clientSecret string, store TokenStore) *tokenSource {
	if store == nil {
		store = &memoryTokenStore{}
	}
	return &tokenSource{client: c, clientID: clientID, clientSecret: clientSecret, store: store}
}

// Token returns a valid access token, fetching or refreshing it if necessary.
func (t *tokenSource) Token(ctx context.Context) (string, error) {
	if token, expiresAt, ok, err := t.store.Get(ctx); err != nil {
		return "", err
	} else if ok && time.Now().Before(expiresAt) {
		return token, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Re-check now that we hold the lock: another goroutine in this process
	// (or another replica, if the store is shared) may have just refreshed.
	if token, expiresAt, ok, err := t.store.Get(ctx); err != nil {
		return "", err
	} else if ok && time.Now().Before(expiresAt) {
		return token, nil
	}

	if locker, ok := t.store.(TokenRefreshLocker); ok {
		release, err := locker.LockRefresh(ctx)
		if err == nil {
			defer release()

			// Whoever held the lock before us may have already refreshed.
			if token, expiresAt, ok, err := t.store.Get(ctx); err != nil {
				return "", err
			} else if ok && time.Now().Before(expiresAt) {
				return token, nil
			}
		}
		// If locking failed (e.g. ctx deadline), fall through and fetch
		// directly rather than blocking indefinitely.
	}

	resp, err := t.fetch(ctx)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - refreshBuffer)
	if err := t.store.Set(ctx, resp.AccessToken, expiresAt); err != nil {
		return "", err
	}
	return resp.AccessToken, nil
}

func (t *tokenSource) fetch(ctx context.Context) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", t.clientID)
	form.Set("client_secret", t.clientSecret)
	form.Set("grant_type", "client_credentials")

	resp, err := do[TokenResponse](ctx, t.client, requestOptions{
		method:     "POST",
		path:       "/token",
		urlEncoded: form,
		authorized: false,
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
