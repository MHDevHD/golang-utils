package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ConvertResponse represents the JSON structure returned by the
// Frankfurter currency conversion API.
type ConvertResponse struct {
	Amount float64            `json:"amount"` // Original amount requested
	Base   string             `json:"base"`   // Base currency code
	Rates  map[string]float64 `json:"rates"`  // Converted rates keyed by currency code
}

// currencySymbols maps ISO currency codes to their display symbols.
// The Frankfurter API provides rates only, so symbols are handled locally.
var currencySymbols = map[string]string{
	"USD": "$",
	"CAD": "$",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"AUD": "$",
	"NZD": "$",
	"CHF": "CHF",
	"CNY": "¥",
	"SEK": "kr",
	"NOK": "kr",
	"DKK": "kr",
	"INR": "₹",
}

// listCurrencies fetches and displays all supported currencies
// from the Frankfurter API.
func listCurrencies() {
	resp, err := http.Get("https://api.frankfurter.app/currencies")
	if err != nil {
		fmt.Println("Request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// The API returns a map of currency code -> currency name
	var currencies map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&currencies); err != nil {
		fmt.Println("Failed to decode response:", err)
		os.Exit(1)
	}

	fmt.Println("Supported currencies:")
	for code, name := range currencies {
		fmt.Printf("%s - %s\n", code, name)
	}
}

func main() {
	// Define command-line flags
	listFlag := flag.Bool("list", false, "List all supported currencies")
	amountFlag := flag.Float64("amount", 0, "Amount to convert")
	fromFlag := flag.String("from", "", "Source currency code (e.g., USD)")
	toFlag := flag.String("to", "", "Target currency code (e.g., CAD)")
	symbolFlag := flag.Bool("symbol", false, "Show currency symbol in output")

	// Parse command-line flags
	flag.Parse()

	// If --list is provided, show supported currencies and exit
	if *listFlag {
		listCurrencies()
		return
	}

	// Validate required flags for conversion
	if *amountFlag <= 0 || *fromFlag == "" || *toFlag == "" {
		fmt.Println("Usage for conversion:")
		fmt.Println("  -amount <value> -from <currency> -to <currency>")
		fmt.Println("  -symbol   Show the currency symbol in the output")
		fmt.Println("Usage to list currencies:")
		fmt.Println("  -list")
		os.Exit(1)
	}

	// Normalize currency codes to uppercase
	from := strings.ToUpper(*fromFlag)
	to := strings.ToUpper(*toFlag)

	// Build the Frankfurter API request URL
	url := fmt.Sprintf(
		"https://api.frankfurter.app/latest?amount=%f&from=%s&to=%s",
		*amountFlag, from, to,
	)

	// Perform the HTTP request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Decode the JSON response into our struct
	var data ConvertResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Failed to decode response:", err)
		os.Exit(1)
	}

	// Extract the converted value for the target currency
	result, ok := data.Rates[to]
	if !ok {
		fmt.Println("Currency not supported")
		os.Exit(1)
	}

	// Output formatting:
	// - If --symbol is provided, show the currency symbol when available
	// - Otherwise, print just the numeric value
	if *symbolFlag {
		if symbol, ok := currencySymbols[to]; ok {
			fmt.Printf("%s%.2f\n", symbol, result)
		} else {
			// Fallback for currencies without a known symbol
			fmt.Printf("%.2f %s\n", result, to)
		}
	} else {
		fmt.Printf("%.2f\n", result)
	}
}
