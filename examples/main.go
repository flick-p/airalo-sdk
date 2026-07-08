// Command example is a runnable walkthrough of the Airalo Partner API Go SDK.
//
// It reads credentials from the environment and only performs read-only
// calls by default. Pass -order to also place a real sandbox order (only do
// this against the sandbox environment).
//
//	export AIRALO_CLIENT_ID=...
//	export AIRALO_CLIENT_SECRET=...
//	go run ./examples
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/flick-p/airalo-sdk"
)

func main() {
	placeOrder := flag.Bool("order", false, "also place a real order for the first available package (sandbox only)")
	flag.Parse()

	clientID := os.Getenv("AIRALO_CLIENT_ID")
	clientSecret := os.Getenv("AIRALO_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		log.Fatal("set AIRALO_CLIENT_ID and AIRALO_CLIENT_SECRET")
	}

	baseURL := airalo.SandboxBaseURL
	if v := os.Getenv("AIRALO_BASE_URL"); v != "" {
		baseURL = v
	}

	client, err := airalo.NewClient(airalo.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		BaseURL:      baseURL,
	})
	if err != nil {
		log.Fatalf("airalo.NewClient: %v", err)
	}

	ctx := context.Background()

	balance, err := client.GetBalance(ctx)
	if err != nil {
		log.Fatalf("GetBalance: %v", err)
	}
	fmt.Printf("balance: %.2f %s\n", balance.Balances.AvailableBalance.Amount, balance.Balances.AvailableBalance.Currency)

	page, err := client.GetPackages(ctx, airalo.GetPackagesParams{
		Type:  airalo.PackageTypeLocal,
		Limit: 1,
	})
	if err != nil {
		log.Fatalf("GetPackages: %v", err)
	}
	if len(page.Data) == 0 || len(page.Data[0].Operators) == 0 || len(page.Data[0].Operators[0].Packages) == 0 {
		log.Fatal("no packages returned")
	}
	pkg := page.Data[0].Operators[0].Packages[0]
	fmt.Printf("first package: %s (%s, %s, $%.2f)\n", pkg.ID, page.Data[0].Title, pkg.Title, pkg.Price)

	info, err := client.GetProductInformation(ctx, pkg.ID)
	if err != nil {
		fmt.Printf("GetProductInformation(%s): %v\n", pkg.ID, err)
	} else {
		fmt.Printf("product information version: %s\n", info.PIBVersion)
	}

	if !*placeOrder {
		fmt.Println("\npass -order to also place a real order for", pkg.ID)
		return
	}

	order, err := client.SubmitOrder(ctx, airalo.SubmitOrderParams{
		Quantity:    1,
		PackageID:   pkg.ID,
		Description: "example order from airalo-go",
	})
	if err != nil {
		log.Fatalf("SubmitOrder: %v", err)
	}
	fmt.Printf("order placed: id=%d code=%s\n", order.ID, order.Code)

	for _, sim := range order.Sims {
		instructions, err := client.GetInstallationInstructions(ctx, sim.ICCID)
		if err != nil {
			log.Fatalf("GetInstallationInstructions: %v", err)
		}
		fmt.Printf("eSIM %s installation language: %s\n", sim.ICCID, instructions.Instructions.Language)
	}
}
