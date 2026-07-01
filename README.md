# airalo-go

An unofficial Go client for the [Airalo Partner API](https://partners-api.airalo.com) (v2).

Generated from `Airalo Partner API documentation (1).json` (a Postman collection export). All
33 documented REST endpoints are covered.

## Install

```sh
go get github.com/airalo/airalo-go
```

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/airalo/airalo-go"
)

func main() {
	client, err := airalo.NewClient(airalo.Config{
		ClientID:     "your-client-id",
		ClientSecret: "your-client-secret",
		BaseURL:      airalo.SandboxBaseURL, // or airalo.ProductionBaseURL
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	page, err := client.GetPackages(ctx, airalo.GetPackagesParams{
		Type:    airalo.PackageTypeLocal,
		Country: "US",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(page.Data[0].Title, page.Meta.Total, "countries")
}
```

A fuller runnable walkthrough lives in [`examples/main.go`](examples/main.go):

```sh
export AIRALO_CLIENT_ID=...
export AIRALO_CLIENT_SECRET=...
go run ./examples
```

## Authentication

The client obtains an OAuth2 `client_credentials` access token on first use (`POST /v2/token`)
and caches it in memory, refreshing automatically shortly before it expires. You never need to
call the token endpoint yourself.

## Error handling

Non-2xx responses are returned as `*airalo.APIError`:

```go
order, err := client.SubmitOrder(ctx, params)
if err != nil {
	var apiErr *airalo.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Message)
		for field, msg := range apiErr.Fields { // populated for 422 validation errors
			fmt.Println(field, msg)
		}
	}
	return
}
```

## Pagination

List endpoints (`GetPackages`, `ListOrders`, `ListESims`) return `airalo.Page[T]`, which carries
`Data`, `Links` (first/last/prev/next), and `Meta` (current_page/last_page/total/...):

```go
page, err := client.ListOrders(ctx, airalo.ListOrdersParams{Limit: 50, Page: 1})
for page.Links.Next != nil {
	// fetch the next page using params.Page++, or follow *page.Links.Next yourself.
}
```

## Coverage

| Category | Methods |
|---|---|
| Auth | handled internally |
| Packages | `GetPackages`, `GetProductInformation` |
| Orders | `SubmitOrder`, `SubmitOrderAsync`, `ListOrders`, `GetOrder`, `SubmitESimVoucher` |
| Future orders | `CreateFutureOrder`, `ListFutureOrders`, `CancelFutureOrders` |
| eSIM lifecycle | `GetESim`, `ListESims`, `GetInstallationInstructions`, `GetUsage`, `UpdateESimBrand`, `GetESimPackageHistory` |
| Top-ups | `ListTopupPackages`, `SubmitTopupOrder` |
| Refunds | `RequestRefund` |
| Notifications | `OptInNotification`, `GetNotificationDetails`, `OptOutNotification`, `SimulateWebhook` |
| Balance | `GetBalance` |
| Devices | `ListCompatibleDevices` (deprecated), `ListCompatibleDevicesLite` |

## Notes on the source documentation

A few endpoints in the source Postman collection had no worked response example (Product
Information, Future Orders create/list, Refund). Those response types are modeled from the
endpoint's prose documentation and preserve any unrecognized JSON fields in an `Extra` map so a
schema change won't silently drop data — see the doc comments on `ProductInformation` and
`FutureOrder`.

The API renders some numeric fields (`quantity`, `price`, `validity`, `per_page`) as either JSON
numbers or numeric strings depending on the endpoint. These are decoded via `FlexInt`/`FlexFloat`,
which accept either representation.

## Development

```sh
go build ./...
go vet ./...
go test ./...
```
