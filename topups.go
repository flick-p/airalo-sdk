package airalo

import (
	"context"
	"net/url"
)

// ListTopupPackages retrieves the top-up packages available for an eSIM, identified by ICCID.
func (c *Client) ListTopupPackages(ctx context.Context, iccid string) ([]PackageOffer, error) {
	return do[[]PackageOffer](ctx, c, requestOptions{
		method:     "GET",
		path:       "/sims/" + url.PathEscape(iccid) + "/topups",
		authorized: true,
	})
}

// SubmitTopupOrderParams configures POST /v2/orders/topups.
type SubmitTopupOrderParams struct {
	// PackageID is the top-up package id, from ListTopupPackages. Required.
	PackageID string
	// ICCID identifies the eSIM to top up. Required.
	ICCID string
	// Description is an optional free-form label to identify the order later.
	Description string
}

func (p SubmitTopupOrderParams) toFormFields() map[string]string {
	fields := map[string]string{
		"package_id": p.PackageID,
		"iccid":      p.ICCID,
	}
	if p.Description != "" {
		fields["description"] = p.Description
	}
	return fields
}

// SubmitTopupOrder purchases a top-up package for an existing eSIM.
func (c *Client) SubmitTopupOrder(ctx context.Context, params SubmitTopupOrderParams) (Order, error) {
	return do[Order](ctx, c, requestOptions{
		method:     "POST",
		path:       "/orders/topups",
		formFields: params.toFormFields(),
		authorized: true,
	})
}
