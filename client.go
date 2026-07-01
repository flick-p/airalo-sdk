// Package airalo provides a Go client for the Airalo Partner API (v2).
package airalo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const (
	// ProductionBaseURL is the base URL of the Airalo Partner API production environment.
	ProductionBaseURL = "https://partners-api.airalo.com"
	// SandboxBaseURL is the base URL of the Airalo Partner API sandbox environment.
	SandboxBaseURL = "https://sandbox-partners-api.airalo.com"

	apiVersionPath = "/v2"
)

// Config holds the settings needed to construct a Client.
type Config struct {
	// ClientID is the OAuth2 client_credentials client id issued by Airalo. Required.
	ClientID string
	// ClientSecret is the OAuth2 client_credentials client secret issued by Airalo. Required.
	ClientSecret string
	// BaseURL overrides the API host. Defaults to ProductionBaseURL.
	BaseURL string
	// HTTPClient overrides the underlying HTTP client. Defaults to http.DefaultClient.
	HTTPClient *http.Client
}

// Client is an Airalo Partner API client. It is safe for concurrent use.
type Client struct {
	baseURL    string
	httpClient *http.Client
	auth       *tokenSource
}

// NewClient builds a Client from the given Config. ClientID and ClientSecret are required.
func NewClient(cfg Config) (*Client, error) {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("airalo: ClientID and ClientSecret are required")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = ProductionBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	c := &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
	}
	c.auth = newTokenSource(c, cfg.ClientID, cfg.ClientSecret)
	return c, nil
}

// envelope is the standard Airalo API response wrapper.
type envelope[T any] struct {
	Data T    `json:"data"`
	Meta meta `json:"meta"`
}

type meta struct {
	Message string `json:"message"`
}

// requestOptions configures a single API call.
type requestOptions struct {
	method      string
	path        string
	query       url.Values
	authorized  bool
	jsonBody    any
	formFields  map[string]string
	formArrays  map[string][]string
	urlEncoded  url.Values
	contentType string
}

func (c *Client) newRequest(ctx context.Context, opt requestOptions) (*http.Request, error) {
	u := c.baseURL + apiVersionPath + opt.path
	if len(opt.query) > 0 {
		u += "?" + opt.query.Encode()
	}

	var body io.Reader
	contentType := opt.contentType

	switch {
	case opt.jsonBody != nil:
		b, err := json.Marshal(opt.jsonBody)
		if err != nil {
			return nil, fmt.Errorf("airalo: encoding request body: %w", err)
		}
		body = bytes.NewReader(b)
		contentType = "application/json"

	case opt.formFields != nil || opt.formArrays != nil:
		buf := &bytes.Buffer{}
		w := multipart.NewWriter(buf)
		for k, v := range opt.formFields {
			if err := w.WriteField(k, v); err != nil {
				return nil, fmt.Errorf("airalo: encoding multipart field %q: %w", k, err)
			}
		}
		for k, values := range opt.formArrays {
			for _, v := range values {
				if err := w.WriteField(k+"[]", v); err != nil {
					return nil, fmt.Errorf("airalo: encoding multipart field %q: %w", k, err)
				}
			}
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("airalo: closing multipart writer: %w", err)
		}
		body = buf
		contentType = w.FormDataContentType()

	case opt.urlEncoded != nil:
		body = strings.NewReader(opt.urlEncoded.Encode())
		contentType = "application/x-www-form-urlencoded"
	}

	req, err := http.NewRequestWithContext(ctx, opt.method, u, body)
	if err != nil {
		return nil, fmt.Errorf("airalo: building request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	if opt.authorized {
		token, err := c.auth.Token(ctx)
		if err != nil {
			return nil, fmt.Errorf("airalo: obtaining access token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req, nil
}

// do executes the request and decodes a successful envelope into out (a pointer),
// or returns an *APIError describing a non-2xx response.
func do[T any](ctx context.Context, c *Client, opt requestOptions) (T, error) {
	var zero T

	req, err := c.newRequest(ctx, opt)
	if err != nil {
		return zero, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("airalo: performing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("airalo: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, newAPIError(resp.StatusCode, respBody)
	}

	if len(respBody) == 0 {
		return zero, nil
	}

	var env envelope[T]
	if err := json.Unmarshal(respBody, &env); err != nil {
		return zero, fmt.Errorf("airalo: decoding response body: %w", err)
	}
	return env.Data, nil
}

// doRaw executes the request like do, but decodes the entire response body
// directly into T instead of unwrapping a {"data": ...} envelope. Use this
// for the handful of endpoints that don't follow the standard envelope shape.
func doRaw[T any](ctx context.Context, c *Client, opt requestOptions) (T, error) {
	var zero T

	req, err := c.newRequest(ctx, opt)
	if err != nil {
		return zero, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return zero, fmt.Errorf("airalo: performing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zero, fmt.Errorf("airalo: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, newAPIError(resp.StatusCode, respBody)
	}

	if len(respBody) == 0 {
		return zero, nil
	}

	var out T
	if err := json.Unmarshal(respBody, &out); err != nil {
		return zero, fmt.Errorf("airalo: decoding response body: %w", err)
	}
	return out, nil
}
