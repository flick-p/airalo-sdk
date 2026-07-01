package airalo

import (
	"context"
	"net/url"
	"strconv"
)

// Image describes a CDN-hosted image with known dimensions.
type Image struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
}

// Network describes a mobile network and the radio technologies it supports.
type Network struct {
	Name  string   `json:"name"`
	Types []string `json:"types"`
}

// Coverage describes network coverage within a country.
type Coverage struct {
	Name     string    `json:"name"`
	Code     string    `json:"code"`
	Networks []Network `json:"networks"`
}

// APNSetting describes the APN configuration for a specific platform.
type APNSetting struct {
	APNType  string `json:"apn_type"`
	APNValue string `json:"apn_value"`
}

// APN groups platform-specific APN settings.
type APN struct {
	IOS     APNSetting `json:"ios"`
	Android APNSetting `json:"android"`
}

// Prices carries per-currency pricing for a package.
type Prices struct {
	NetPrice               map[string]float64 `json:"net_price"`
	RecommendedRetailPrice map[string]float64 `json:"recommended_retail_price"`
}

// PackageOffer is a single purchasable package (sim or topup) offered by an operator.
type PackageOffer struct {
	ID                 string  `json:"id"`
	Type               string  `json:"type"` // "sim" or "topup"
	Price              float64 `json:"price"`
	NetPrice           float64 `json:"net_price"`
	Amount             int     `json:"amount"`
	Day                int     `json:"day"`
	IsUnlimited        bool    `json:"is_unlimited"`
	Title              string  `json:"title"`
	Data               string  `json:"data"`
	ShortInfo          string  `json:"short_info"`
	Voice              *int    `json:"voice"`
	Text               *int    `json:"text"`
	QRInstallation     string  `json:"qr_installation"`
	ManualInstallation string  `json:"manual_installation"`
	Prices             Prices  `json:"prices"`
}

// Operator describes a mobile operator offering eSIM packages for a country/region.
type Operator struct {
	ID               int            `json:"id"`
	Style            string         `json:"style"`
	GradientStart    string         `json:"gradient_start"`
	GradientEnd      string         `json:"gradient_end"`
	Type             string         `json:"type"` // "local" or "global"
	IsPrepaid        bool           `json:"is_prepaid"`
	Title            string         `json:"title"`
	ESimType         string         `json:"esim_type"`
	Warning          *string        `json:"warning"`
	APNType          string         `json:"apn_type"`
	APNValue         string         `json:"apn_value"`
	IsRoaming        bool           `json:"is_roaming"`
	Info             []string       `json:"info"`
	Image            Image          `json:"image"`
	PlanType         string         `json:"plan_type"`
	ActivationPolicy string         `json:"activation_policy"`
	IsKYCVerify      bool           `json:"is_kyc_verify"`
	Rechargeability  bool           `json:"rechargeability"`
	OtherInfo        string         `json:"other_info"`
	Coverages        []Coverage     `json:"coverages"`
	APN              APN            `json:"apn"`
	Packages         []PackageOffer `json:"packages"`
}

// CountryPackages groups all operators/packages available for a country or region.
type CountryPackages struct {
	Slug        string     `json:"slug"`
	CountryCode string     `json:"country_code"`
	Title       string     `json:"title"`
	Image       Image      `json:"image"`
	Operators   []Operator `json:"operators"`
}

// PackageType filters packages by operator type in GetPackagesParams.
type PackageType string

const (
	PackageTypeLocal  PackageType = "local"
	PackageTypeGlobal PackageType = "global"
)

// GetPackagesParams configures the GET /v2/packages request. All fields are optional.
type GetPackagesParams struct {
	// Type filters by "local" or "global" packages.
	Type PackageType
	// Country filters local packages by ISO country code (e.g. "US", "DE").
	Country string
	// Limit sets how many items are returned per page.
	Limit int
	// Page selects the pagination page (1-indexed).
	Page int
	// IncludeTopups includes topup packages in the response when true.
	IncludeTopups bool
}

func (p GetPackagesParams) toQuery() url.Values {
	q := url.Values{}
	if p.Type != "" {
		q.Set("filter[type]", string(p.Type))
	}
	if p.Country != "" {
		q.Set("filter[country]", p.Country)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.IncludeTopups {
		q.Set("include", "topup")
	}
	return q
}

// GetPackages retrieves the paginated catalogue of eSIM packages, optionally
// filtered by type (local/global) and/or country.
func (c *Client) GetPackages(ctx context.Context, params GetPackagesParams) (Page[[]CountryPackages], error) {
	return doPage[[]CountryPackages](ctx, c, requestOptions{
		method:     "GET",
		path:       "/packages",
		query:      params.toQuery(),
		authorized: true,
	})
}
