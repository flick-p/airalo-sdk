package airalo

// Sim represents a physical/eSIM record. Which fields are populated depends
// on the endpoint: the compact form nested under Order.Sims carries the
// installation essentials, while GetESim/ListESims responses may add
// VoucherCode, BrandSettingsName, Order, and Simable when requested via
// the `include` query parameter.
type Sim struct {
	ID                         int      `json:"id"`
	CreatedAt                  string   `json:"created_at"`
	ICCID                      string   `json:"iccid"`
	LPA                        string   `json:"lpa"`
	IMSIs                      *string  `json:"imsis"`
	MatchingID                 string   `json:"matching_id"`
	QRCode                     string   `json:"qrcode"`
	QRCodeURL                  string   `json:"qrcode_url"`
	DirectAppleInstallationURL string   `json:"direct_apple_installation_url"`
	VoucherCode                *string  `json:"voucher_code"`
	AiraloCode                 *string  `json:"airalo_code"`
	APNType                    string   `json:"apn_type"`
	APNValue                   *string  `json:"apn_value"`
	IsRoaming                  bool     `json:"is_roaming"`
	ConfirmationCode           *string  `json:"confirmation_code"`
	BrandSettingsName          *string  `json:"brand_settings_name"`
	Order                      *Order   `json:"order"`
	Simable                    *Simable `json:"simable"`
}

// Status is a generic named status, e.g. an order's fulfillment status.
type Status struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// OrderUser describes the partner-side user/account an order or eSIM belongs to.
type OrderUser struct {
	ID         int     `json:"id"`
	CreatedAt  string  `json:"created_at"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Mobile     *string `json:"mobile"`
	Address    *string `json:"address"`
	State      *string `json:"state"`
	City       *string `json:"city"`
	PostalCode *string `json:"postal_code"`
	CountryID  *int    `json:"country_id"`
	Company    string  `json:"company"`
}

// Sharing describes an eSIM sharing link generated when an order specifies to_email.
type Sharing struct {
	Link       string `json:"link"`
	AccessCode string `json:"access_code"`
}

// Simable is the order/topup summary embedded in a Sim when fetched via
// GetESim/ListESims with include=order.
type Simable struct {
	ID                 int               `json:"id"`
	CreatedAt          string            `json:"created_at"`
	Code               string            `json:"code"`
	Description        *string           `json:"description"`
	Type               string            `json:"type"`
	PackageID          string            `json:"package_id"`
	Quantity           FlexInt           `json:"quantity"`
	Package            string            `json:"package"`
	ESimType           string            `json:"esim_type"`
	Validity           FlexInt           `json:"validity"`
	Price              FlexFloat         `json:"price"`
	Data               string            `json:"data"`
	Currency           string            `json:"currency"`
	ManualInstallation string            `json:"manual_installation"`
	QRCodeInstallation string            `json:"qrcode_installation"`
	InstallationGuides map[string]string `json:"installation_guides"`
	Status             *Status           `json:"status"`
	User               *OrderUser        `json:"user"`
	Sharing            *Sharing          `json:"sharing"`
}

// Order represents a placed order for a new eSIM ("sim") or a top-up applied
// to an existing eSIM ("topup"). Some numeric fields are typed FlexInt /
// FlexFloat because the API renders them inconsistently as JSON numbers or
// numeric strings depending on the endpoint.
type Order struct {
	ID                 int               `json:"id"`
	Code               string            `json:"code"`
	CreatedAt          string            `json:"created_at"`
	Description        *string           `json:"description"`
	Type               string            `json:"type"`
	PackageID          string            `json:"package_id"`
	Quantity           FlexInt           `json:"quantity"`
	Package            string            `json:"package"`
	ESimType           string            `json:"esim_type"`
	Validity           FlexInt           `json:"validity"`
	Price              FlexFloat         `json:"price"`
	Data               string            `json:"data"`
	Currency           string            `json:"currency"`
	ManualInstallation string            `json:"manual_installation"`
	QRCodeInstallation string            `json:"qrcode_installation"`
	InstallationGuides map[string]string `json:"installation_guides"`
	BrandSettingsName  *string           `json:"brand_settings_name"`
	Sims               []Sim             `json:"sims"`
	User               *OrderUser        `json:"user"`
	Status             *Status           `json:"status"`
}
