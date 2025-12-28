package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// buildURL constructs the DinoPass request URL based on
// password complexity and (for strong passwords) API options.
func buildURL(
	complexity string,
	length int,
	useSymbols bool,
	useNumbers bool,
	useCapitals bool,
) (string, error) {

	base := "https://www.dinopass.com/password/"

	// Simple passwords require no query parameters
	if complexity == "simple" {
		return base + "simple", nil
	}

	// Only "simple" and "strong" are valid options
	if complexity != "strong" {
		return "", fmt.Errorf("invalid complexity: %s", complexity)
	}

	// Build the strong-password URL with query parameters
	u, err := url.Parse(base + "strong")
	if err != nil {
		return "", err
	}

	q := u.Query()

	// Length is validated before being sent to the API
	if length >= 7 && length <= 20 {
		q.Set("length", fmt.Sprintf("%d", length))
	}

	// Strong password feature flags
	q.Set("useSymbols", fmt.Sprintf("%t", useSymbols))
	q.Set("useNumbers", fmt.Sprintf("%t", useNumbers))
	q.Set("useCapitals", fmt.Sprintf("%t", useCapitals))

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// getPassword performs the HTTP request and returns
// the generated password as plain text.
func getPassword(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// DinoPass returns raw text, so trimming is sufficient
	return strings.TrimSpace(string(body)), nil
}

func main() {
	// General CLI flags
	complexity := flag.String("complexity", "simple", "Password complexity: simple or strong")
	count := flag.Int("count", 1, "Number of passwords to generate")
	delay := flag.Int("delay", 500, "Delay between requests (milliseconds)")

	// Strong-password flags (mapped directly to DinoPass API)
	length := flag.Int("length", 12, "Password length (7–20) [strong only]")
	useSymbols := flag.Bool("symbols", false, "Include symbols [strong only]")
	useNumbers := flag.Bool("numbers", true, "Include numbers [strong only]")
	useCapitals := flag.Bool("capitals", true, "Include capital letters [strong only]")

	flag.Parse()

	// Build the request URL once and reuse it
	url, err := buildURL(
		*complexity,
		*length,
		*useSymbols,
		*useNumbers,
		*useCapitals,
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Generate the requested number of passwords
	for i := 0; i < *count; i++ {
		pass, err := getPassword(url)
		if err != nil {
			fmt.Println("Error generating password:", err)
			return
		}

		fmt.Printf("[%d] %s\n", i+1, pass)

		// Small delay to avoid spamming the API
		if i < *count-1 {
			time.Sleep(time.Duration(*delay) * time.Millisecond)
		}
	}
}
