package airalo

import "context"

// ESimVoucherItem requests redemption of one or more voucher codes against a package.
type ESimVoucherItem struct {
	PackageID        string   `json:"package_id"`
	Codes            []string `json:"codes"`
	BookingReference string   `json:"booking_reference,omitempty"`
}

// SubmitESimVoucherParams configures POST /v2/voucher/esim.
type SubmitESimVoucherParams struct {
	Vouchers []ESimVoucherItem `json:"vouchers"`
}

// ESimVoucherResult confirms redemption of a voucher batch for one package.
type ESimVoucherResult struct {
	PackageID        string   `json:"package_id"`
	Codes            []string `json:"codes"`
	BookingReference string   `json:"booking_reference"`
}

// SubmitESimVoucher redeems one or more pre-purchased voucher codes for eSIMs.
func (c *Client) SubmitESimVoucher(ctx context.Context, params SubmitESimVoucherParams) ([]ESimVoucherResult, error) {
	return do[[]ESimVoucherResult](ctx, c, requestOptions{
		method:     "POST",
		path:       "/voucher/esim",
		jsonBody:   params,
		authorized: true,
	})
}
