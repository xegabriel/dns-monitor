package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Load configuration from environment variables with validation
func LoadConfig() (Config, error) {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return Config{}, errors.New("DOMAIN environment variable is required (e.g., test.ro)")
	}

	pushoverToken := os.Getenv("PUSHOVER_TOKEN")
	if pushoverToken == "" {
		return Config{}, errors.New("PUSHOVER_TOKEN environment variable is required")
	}

	pushoverUser := os.Getenv("PUSHOVER_USER")
	if pushoverUser == "" {
		return Config{}, errors.New("PUSHOVER_USER environment variable is required")
	}

	// DNS server with default
	dnsServer := os.Getenv("DNS_SERVER")
	if dnsServer == "" {
		dnsServer = "1.1.1.1:53"
	}

	// Check interval with default and validation
	checkInterval := 1 * time.Hour
	intervalStr := os.Getenv("CHECK_INTERVAL")
	if intervalStr != "" {
		duration, err := time.ParseDuration(intervalStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid CHECK_INTERVAL format: %v", err)
		}
		if duration < 1*time.Minute {
			log.Println("Warning: CHECK_INTERVAL less than 1 minute may cause excessive API calls")
		}
		checkInterval = duration
	}

	// Notification on errors setting
	notifyOnErrors := false
	if os.Getenv("NOTIFY_ON_ERRORS") == "true" {
		notifyOnErrors = true
	}

	validCustomSubdomains := getValidEntries("CUSTOM_SUBDOMAINS")
	validCustomDkimSelectors := getValidEntries("CUSTOM_DKIM_SELECTORS")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: false,
			IdleConnTimeout:   90 * time.Second,
			MaxIdleConns:      100,
			MaxConnsPerHost:   10,
		},
	}

	// Create DNS Client
	dnsClient := new(dns.Client)

	return Config{
		Domain:              domain,
		CustomDomains:       validCustomSubdomains,
		CustomDkimSelectors: validCustomDkimSelectors,
		DNSServer:           dnsServer,
		DNSClient:           *dnsClient,
		CheckInterval:       checkInterval,
		HTTPClient:          client,
		PushoverToken:       pushoverToken,
		PushoverUser:        pushoverUser,
		NotifyOnErrors:      notifyOnErrors,
	}, nil

}

// getValidEntries fetches and processes a comma-separated environment variable.
func getValidEntries(envVar string) []string {
	rawValue := os.Getenv(envVar)
	if rawValue == "" {
		return nil
	}

	var validEntries []string
	for _, entry := range strings.Split(rawValue, ",") {
		entry = strings.TrimSpace(entry)
		if entry != "" {
			validEntries = append(validEntries, entry)
		}
	}
	return validEntries
}
