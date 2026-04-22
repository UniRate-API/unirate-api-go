// Basic runnable example for the UniRate Go client.
//
//	UNIRATE_API_KEY=your-key go run ./examples/basic
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	unirate "github.com/UniRate-API/unirate-api-go"
)

func main() {
	key := os.Getenv("UNIRATE_API_KEY")
	if key == "" {
		log.Fatal("UNIRATE_API_KEY env var is required")
	}

	client := unirate.New(key, unirate.WithTimeout(15*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rate, err := client.GetRate(ctx, "USD", "EUR")
	if err != nil {
		log.Fatalf("GetRate: %v", err)
	}
	fmt.Printf("USD -> EUR: %.4f\n", rate)

	euros, err := client.Convert(ctx, 100, "USD", "EUR")
	if err != nil {
		log.Fatalf("Convert: %v", err)
	}
	fmt.Printf("100 USD = %.2f EUR\n", euros)

	currencies, err := client.GetSupportedCurrencies(ctx)
	if err != nil {
		log.Fatalf("GetSupportedCurrencies: %v", err)
	}
	fmt.Printf("%d currencies supported\n", len(currencies))
}
