package airalo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ProductInformation is the standardized, regulator-ready product summary for
// a single eSIM package (the "Product Information Brochure" / PIB data).
//
// Field names are drawn from the endpoint's documentation; no worked example
// was present in the source API collection. Any response fields not listed
// below are preserved verbatim in Extra so callers are never blocked on an
// undocumented or newly added field.
type ProductInformation struct {
	PackageSlug       string `json:"package_slug"`
	PIBVersion        string `json:"pib_version"`
	PIBVersionDate    string `json:"pib_version_date"`
	VersionHash       string `json:"version_hash"`
	PIBVersionPDFLink string `json:"pib_version_pdf_link"`

	ServiceInternet  bool `json:"service_internet"`
	ServiceTelephony bool `json:"service_telephony"`
	ServiceSMS       bool `json:"service_sms"`

	DataAllowanceMB      *float64 `json:"data_allowance_mb"`
	DataAllowanceDisplay string   `json:"data_allowance_display"`
	ValidityDays         int      `json:"validity_days"`
	VoiceMinutes         *int     `json:"voice_minutes"`
	SMSIncluded          *int     `json:"sms_included"`
	IsUnlimited          bool     `json:"is_unlimited"`

	NetworkProviders    []string `json:"network_providers"`
	NetworkTechnologies []string `json:"network_technologies"`

	MaxDownloadSpeedMbit *float64 `json:"max_download_speed_mbit"`
	MaxUploadSpeedMbit   *float64 `json:"max_upload_speed_mbit"`
	IsFairUsagePolicy    bool     `json:"is_fair_usage_policy"`
	FairUsagePolicy      string   `json:"fair_usage_policy"`
	ThrottleThresholdMB  *float64 `json:"throttle_threshold_mb"`

	RetailPriceEUR          *float64 `json:"retail_price_eur"`
	OutOfBundleRateEURPerMB *float64 `json:"out_of_bundle_rate_eur_per_mb"`

	PartnerBrandName string `json:"partner_brand_name"`
	PartnerLogoURL   string `json:"partner_logo_url"`
	LegalEntityName  string `json:"legal_entity_name"`
	LegalAddress     string `json:"legal_address"`
	TermsURL         string `json:"terms_url"`
	AutoRenewal      bool   `json:"auto_renewal"`

	// Extra holds any response fields not covered by the named fields above,
	// keyed by their original JSON field name.
	Extra map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON decodes known fields into their typed struct fields and
// preserves anything else in Extra.
func (p *ProductInformation) UnmarshalJSON(b []byte) error {
	type alias ProductInformation
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return fmt.Errorf("airalo: decoding ProductInformation: %w", err)
	}
	*p = ProductInformation(a)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("airalo: decoding ProductInformation: %w", err)
	}
	for _, known := range productInformationKnownFields {
		delete(raw, known)
	}
	if len(raw) > 0 {
		p.Extra = raw
	}
	return nil
}

var productInformationKnownFields = []string{
	"package_slug", "pib_version", "pib_version_date", "version_hash", "pib_version_pdf_link",
	"service_internet", "service_telephony", "service_sms",
	"data_allowance_mb", "data_allowance_display", "validity_days", "voice_minutes", "sms_included", "is_unlimited",
	"network_providers", "network_technologies",
	"max_download_speed_mbit", "max_upload_speed_mbit", "is_fair_usage_policy", "fair_usage_policy", "throttle_threshold_mb",
	"retail_price_eur", "out_of_bundle_rate_eur_per_mb",
	"partner_brand_name", "partner_logo_url", "legal_entity_name", "legal_address", "terms_url", "auto_renewal",
}

// GetProductInformation retrieves the standardized Product Information (PIB)
// for a single package, identified by its package slug.
//
// This endpoint is rate-limited by Airalo to 60 requests/minute per company
// and per unique package. An unknown package_slug returns a 404 *APIError.
func (c *Client) GetProductInformation(ctx context.Context, packageSlug string) (ProductInformation, error) {
	return do[ProductInformation](ctx, c, requestOptions{
		method:     "GET",
		path:       "/packages/" + url.PathEscape(packageSlug) + "/product-information",
		authorized: true,
	})
}
