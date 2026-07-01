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

// tokenSource fetches and caches an OAuth2 client_credentials access token,
// refreshing it automatically as it nears expiry. Safe for concurrent use.
type tokenSource struct {
	client       *Client
	clientID     string
	clientSecret string

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func newTokenSource(c *Client, clientID, clientSecret string) *tokenSource {
	return &tokenSource{client: c, clientID: clientID, clientSecret: clientSecret}
}

// Token returns a valid access token, fetching or refreshing it if necessary.
func (t *tokenSource) Token(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.token != "" && time.Now().Before(t.expiresAt) {
		return t.token, nil
	}

	resp, err := t.fetch(ctx)
	if err != nil {
		return "", err
	}

	t.token = resp.AccessToken
	t.expiresAt = time.Now().Add(time.Duration(resp.ExpiresIn)*time.Second - refreshBuffer)
	return t.token, nil
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
