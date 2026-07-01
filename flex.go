package airalo

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexFloat decodes a JSON field that the API may render as either a number
// or a numeric string (observed for fields like order/package "price" and
// "validity", which vary in type between endpoints and even between
// responses of the same endpoint).
type FlexFloat float64

// UnmarshalJSON implements json.Unmarshaler.
func (f *FlexFloat) UnmarshalJSON(b []byte) error {
	var n float64
	if err := json.Unmarshal(b, &n); err == nil {
		*f = FlexFloat(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("airalo: FlexFloat: %w", err)
	}
	if s == "" {
		*f = 0
		return nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("airalo: FlexFloat: %w", err)
	}
	*f = FlexFloat(n)
	return nil
}
