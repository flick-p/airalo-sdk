package airalo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// FlexInt decodes a JSON field that the API may render as either a number or
// a numeric string (observed for fields like "per_page").
type FlexInt int

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexInt) UnmarshalJSON(b []byte) error {
	var n int
	if err := json.Unmarshal(b, &n); err == nil {
		*f = FlexInt(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("airalo: FlexInt: %w", err)
	}
	if s == "" {
		*f = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("airalo: FlexInt: %w", err)
	}
	*f = FlexInt(n)
	return nil
}

// PageMeta carries Laravel-style pagination metadata returned alongside list responses.
type PageMeta struct {
	Message     string  `json:"message"`
	CurrentPage int     `json:"current_page"`
	From        int     `json:"from"`
	LastPage    int     `json:"last_page"`
	Path        string  `json:"path"`
	PerPage     FlexInt `json:"per_page"`
	To          int     `json:"to"`
	Total       int     `json:"total"`
}

// PageLinks carries pagination navigation links returned alongside list responses.
type PageLinks struct {
	First string  `json:"first"`
	Last  string  `json:"last"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
}

// Page wraps a paginated list response.
type Page[T any] struct {
	Data  T         `json:"data"`
	Links PageLinks `json:"links"`
	Meta  PageMeta  `json:"meta"`
}

// doPage executes the request and decodes a paginated envelope (data + links + meta).
func doPage[T any](ctx context.Context, c *Client, opt requestOptions) (Page[T], error) {
	var zero Page[T]

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

	var page Page[T]
	if err := json.Unmarshal(respBody, &page); err != nil {
		return zero, fmt.Errorf("airalo: decoding response body: %w", err)
	}
	return page, nil
}
