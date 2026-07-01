package airalo

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// GetESim retrieves a single eSIM by ICCID. include may contain "order",
// "order.status", "order.user", and "share" to expand related data.
func (c *Client) GetESim(ctx context.Context, iccid string, include []string) (Sim, error) {
	q := url.Values{}
	if len(include) > 0 {
		q.Set("include", strings.Join(include, ","))
	}
	return do[Sim](ctx, c, requestOptions{
		method:     "GET",
		path:       "/sims/" + url.PathEscape(iccid),
		query:      q,
		authorized: true,
	})
}

// ListESimsParams filters/paginates GET /v2/sims. All fields are optional.
type ListESimsParams struct {
	// Include adds related data. Valid values: "order", "order.status", "order.user", "share".
	Include []string
	// CreatedAtFrom and CreatedAtTo filter by creation date range (format "2006-01-02").
	// Both must be set to apply the filter.
	CreatedAtFrom string
	CreatedAtTo   string
	// ICCID performs a partial match against the eSIM's ICCID.
	ICCID string
	// Limit sets how many eSIMs are returned per page.
	Limit int
	// Page selects the pagination page (1-indexed).
	Page int
}

func (p ListESimsParams) toQuery() url.Values {
	q := url.Values{}
	if len(p.Include) > 0 {
		q.Set("include", strings.Join(p.Include, ","))
	}
	if p.CreatedAtFrom != "" && p.CreatedAtTo != "" {
		q.Set("filter[created_at]", p.CreatedAtFrom+" - "+p.CreatedAtTo)
	}
	if p.ICCID != "" {
		q.Set("filter[iccid]", p.ICCID)
	}
	if p.Limit > 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	return q
}

// ListESims retrieves a paginated list of eSIMs.
func (c *Client) ListESims(ctx context.Context, params ListESimsParams) (Page[[]Sim], error) {
	return doPage[[]Sim](ctx, c, requestOptions{
		method:     "GET",
		path:       "/sims",
		query:      params.toQuery(),
		authorized: true,
	})
}

// BrandUpdateResult confirms the brand applied to an eSIM.
type BrandUpdateResult struct {
	BrandSettingsName *string `json:"brand_settings_name"`
}

// UpdateESimBrand sets (or clears, when brandSettingsName is empty) the brand
// under which an eSIM's sharing pages/emails are presented.
func (c *Client) UpdateESimBrand(ctx context.Context, iccid string, brandSettingsName string) (BrandUpdateResult, error) {
	return do[BrandUpdateResult](ctx, c, requestOptions{
		method: "PUT",
		path:   "/sims/" + url.PathEscape(iccid) + "/brand",
		formFields: map[string]string{
			"brand_settings_name": brandSettingsName,
		},
		authorized: true,
	})
}

// PackageSummary is the compact package descriptor embedded in a PackageHistoryEntry.
type PackageSummary struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Price       FlexFloat `json:"price"`
	NetPrice    FlexFloat `json:"net_price"`
	Amount      int       `json:"amount"`
	Day         int       `json:"day"`
	IsUnlimited bool      `json:"is_unlimited"`
	Title       string    `json:"title"`
	Data        string    `json:"data"`
	ShortInfo   *string   `json:"short_info"`
}

// PackageHistoryEntry describes one package (initial or top-up) ever applied to an eSIM.
type PackageHistoryEntry struct {
	ID          int            `json:"id"`
	Status      string         `json:"status"`
	Remaining   int            `json:"remaining"`
	ActivatedAt string         `json:"activated_at"`
	ExpiredAt   *string        `json:"expired_at"`
	FinishedAt  *string        `json:"finished_at"`
	Package     PackageSummary `json:"package"`
}

// GetESimPackageHistory retrieves the full history of packages (including
// top-ups) ever applied to an eSIM, identified by ICCID.
func (c *Client) GetESimPackageHistory(ctx context.Context, iccid string) ([]PackageHistoryEntry, error) {
	return do[[]PackageHistoryEntry](ctx, c, requestOptions{
		method:     "GET",
		path:       "/sims/" + url.PathEscape(iccid) + "/packages",
		authorized: true,
	})
}

// Usage reports the remaining and total data/voice/text allowance for an eSIM.
type Usage struct {
	Remaining      int    `json:"remaining"`
	Total          int    `json:"total"`
	ExpiredAt      string `json:"expired_at"`
	IsUnlimited    bool   `json:"is_unlimited"`
	Status         string `json:"status"` // NOT_ACTIVE, ACTIVE, FINISHED, UNKNOWN, EXPIRED
	RemainingVoice int    `json:"remaining_voice"`
	RemainingText  int    `json:"remaining_text"`
	TotalVoice     int    `json:"total_voice"`
	TotalText      int    `json:"total_text"`
}

// GetUsage retrieves the current data/voice/text usage for an eSIM, identified by ICCID.
func (c *Client) GetUsage(ctx context.Context, iccid string) (Usage, error) {
	return do[Usage](ctx, c, requestOptions{
		method:     "GET",
		path:       "/sims/" + url.PathEscape(iccid) + "/usage",
		authorized: true,
	})
}

// InstallationSteps maps step number (as a string, e.g. "1") to its instruction text.
type InstallationSteps map[string]string

// QRCodeInstallation describes QR-code based eSIM installation for one device profile.
type QRCodeInstallation struct {
	Steps                      InstallationSteps `json:"steps"`
	QRCodeData                 string            `json:"qr_code_data"`
	QRCodeURL                  string            `json:"qr_code_url"`
	DirectAppleInstallationURL string            `json:"direct_apple_installation_url,omitempty"`
}

// ManualInstallation describes manual (SM-DP+ address) eSIM installation for one device profile.
type ManualInstallation struct {
	Steps                        InstallationSteps `json:"steps"`
	SMDPAddressAndActivationCode string            `json:"smdp_address_and_activation_code"`
}

// NetworkSetup describes post-installation network/APN setup for one device profile.
type NetworkSetup struct {
	Steps     InstallationSteps `json:"steps"`
	APNType   string            `json:"apn_type"`
	APNValue  string            `json:"apn_value"`
	IsRoaming bool              `json:"is_roaming"`
}

// DeviceInstallation groups installation guidance for a specific device model/OS version.
type DeviceInstallation struct {
	Model                 *string             `json:"model"`
	Version               *string             `json:"version"`
	InstallationViaQRCode *QRCodeInstallation `json:"installation_via_qr_code"`
	InstallationManual    *ManualInstallation `json:"installation_manual"`
	NetworkSetup          *NetworkSetup       `json:"network_setup"`
}

// InstallationInstructions groups per-platform device installation guidance in one language.
type InstallationInstructions struct {
	Language string               `json:"language"`
	IOS      []DeviceInstallation `json:"ios"`
	Android  []DeviceInstallation `json:"android"`
}

// InstallationInstructionsResult wraps the instructions payload returned by GetInstallationInstructions.
type InstallationInstructionsResult struct {
	Instructions InstallationInstructions `json:"instructions"`
}

// GetInstallationInstructions retrieves step-by-step iOS/Android installation
// guidance for an eSIM, identified by ICCID.
func (c *Client) GetInstallationInstructions(ctx context.Context, iccid string) (InstallationInstructionsResult, error) {
	return do[InstallationInstructionsResult](ctx, c, requestOptions{
		method:     "GET",
		path:       "/sims/" + url.PathEscape(iccid) + "/instructions",
		authorized: true,
	})
}
