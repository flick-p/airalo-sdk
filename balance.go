package airalo

import "context"

// Money is a currency amount pair.
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Balances groups the partner account's available balance.
type Balances struct {
	Name             string `json:"name"`
	AvailableBalance Money  `json:"availableBalance"`
}

// BalanceResult wraps the partner account balance.
type BalanceResult struct {
	Balances Balances `json:"balances"`
}

// GetBalance retrieves the partner account's current available balance.
func (c *Client) GetBalance(ctx context.Context) (BalanceResult, error) {
	return do[BalanceResult](ctx, c, requestOptions{
		method:     "GET",
		path:       "/balance",
		authorized: true,
	})
}
