package airalo

import (
	"encoding/json"
	"fmt"
)

// APIError represents a non-2xx response from the Airalo Partner API.
//
// Validation failures (HTTP 422) typically populate Fields with a map of
// field name -> error message. RawData preserves the original "data" payload
// for any shape that doesn't fit that pattern (e.g. an array or empty object).
type APIError struct {
	StatusCode int
	Message    string
	Fields     map[string]string
	RawData    json.RawMessage
}

func (e *APIError) Error() string {
	if len(e.Fields) > 0 {
		return fmt.Sprintf("airalo: request failed with status %d: %s %v", e.StatusCode, e.Message, e.Fields)
	}
	return fmt.Sprintf("airalo: request failed with status %d: %s", e.StatusCode, e.Message)
}

func newAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode}

	var raw struct {
		Data json.RawMessage `json:"data"`
		Meta meta            `json:"meta"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		apiErr.Message = string(body)
		return apiErr
	}

	apiErr.Message = raw.Meta.Message
	apiErr.RawData = raw.Data

	var fields map[string]string
	if json.Unmarshal(raw.Data, &fields) == nil {
		apiErr.Fields = fields
	}

	if apiErr.Message == "" && len(body) > 0 {
		apiErr.Message = string(body)
	}

	return apiErr
}
