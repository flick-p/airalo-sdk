package airalo

import "context"

// CompatibleDevice identifies a device model known to support eSIM.
type CompatibleDevice struct {
	Model string `json:"model"`
	OS    string `json:"os"`
	Brand string `json:"brand"`
	Name  string `json:"name"`
}

// ListCompatibleDevices retrieves the full list of eSIM-compatible devices.
//
// Deprecated: Airalo has deprecated this endpoint in favor of
// ListCompatibleDevicesLite.
func (c *Client) ListCompatibleDevices(ctx context.Context) ([]CompatibleDevice, error) {
	return do[[]CompatibleDevice](ctx, c, requestOptions{
		method:     "GET",
		path:       "/compatible-devices",
		authorized: true,
	})
}

// ListCompatibleDevicesLite retrieves the lightweight list of eSIM-compatible devices.
//
// The source API documentation did not include a worked response example for
// this endpoint; it is assumed to share CompatibleDevice's shape with the
// non-lite endpoint.
func (c *Client) ListCompatibleDevicesLite(ctx context.Context) ([]CompatibleDevice, error) {
	return do[[]CompatibleDevice](ctx, c, requestOptions{
		method:     "GET",
		path:       "/compatible-devices-lite",
		authorized: true,
	})
}
