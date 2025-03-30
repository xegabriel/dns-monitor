package common

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (

	// Pushover config
	PushoverUserKeyEnv  = "PUSHOVER_USER_KEY"
	PushoverAppTokenEnv = "PUSHOVER_APP_TOKEN"

	// Telegram config
	TelegramBotTokenEnv = "TELEGRAM_BOT_TOKEN"
	TelegramChatIDsEnv  = "TELEGRAM_CHAT_IDS" // Comma-separated list of chat IDs

	// Add more environment variables for other providers as needed

)

// Load configuration from environment variables with validation
func LoadConfig() (*Config, error) {
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return &Config{}, errors.New("DOMAIN environment variable is required (e.g., test.ro)")
	}

	notificationConfig, err := loadNotificationConfig()
	if err != nil {
		return &Config{}, fmt.Errorf("failed to load notification config: %v", err)
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
			return &Config{}, fmt.Errorf("invalid CHECK_INTERVAL format: %v", err)
		}
		if duration < 1*time.Minute {
			log.Println("⚠️ Warning: CHECK_INTERVAL less than 1 minute may cause excessive API calls ⚠️")
		}
		checkInterval = duration
	}

	// Notification on errors setting
	notifyOnErrors := false
	if os.Getenv("NOTIFY_ON_ERRORS") == "true" {
		notifyOnErrors = true
	}
	log.Printf("🔔 Notify on errors: %v 🔔", notifyOnErrors)

	validCustomSubdomains := getValidEntries("CUSTOM_SUBDOMAINS", parseString)
	log.Printf("🌐 Custom subdomains: %v 🌐", validCustomSubdomains)

	validCustomDkimSelectors := getValidEntries("CUSTOM_DKIM_SELECTORS", parseString)
	log.Printf("🛡️ Custom DKIM selectors: %v 🛡️", validCustomDkimSelectors)

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

	return &Config{
		Domain:              domain,
		CustomSubdomains:    validCustomSubdomains,
		CustomDkimSelectors: validCustomDkimSelectors,
		DNSServer:           dnsServer,
		DNSClient:           *dnsClient,
		CheckInterval:       checkInterval,
		HTTPClient:          client,
		NotificationConfig:  *notificationConfig,
		NotifyOnErrors:      notifyOnErrors,
	}, nil

}

// NotificationConfig holds the configuration for notification services
func loadNotificationConfig() (*NotificationConfig, error) {
	notifierType := os.Getenv("NOTIFIER_TYPE")
	if !IsValidNotifierType(notifierType) {
		return &NotificationConfig{}, errors.New("NOTIFIER_TYPE environment variable is required")
	}
	config := NotificationConfig{
		NotifierType: notifierType,
	}
	switch notifierType {
	case NotifierTypePushover:
		config.PushoverToken = os.Getenv(PushoverAppTokenEnv)
		config.PushoverUser = os.Getenv(PushoverUserKeyEnv)
		if config.PushoverToken == "" || config.PushoverUser == "" {
			return nil, fmt.Errorf("%s and %s environment variable are required", PushoverAppTokenEnv, PushoverUserKeyEnv)
		}
	case NotifierTypeTelegram:
		config.TelegramBotToken = os.Getenv(TelegramBotTokenEnv)
		config.TelegramChatIDs = getValidEntries(TelegramChatIDsEnv, parseInt64)
		if config.TelegramBotToken == "" || len(config.TelegramChatIDs) == 0 {
			return nil, fmt.Errorf("%s and %s environment variable are required", TelegramBotTokenEnv, TelegramChatIDsEnv)
		}
	default:
		return nil, fmt.Errorf("unsupported notifier type: %s", notifierType)
	}

	return &config, nil
}

// IsValidNotifierType checks if the provided notifier type is valid.
func IsValidNotifierType(input string) bool {
	if input == "" {
		return false
	}
	for _, nt := range NotifierTypes {
		if input == nt {
			return true
		}
	}
	return false
}

// getValidEntries fetches and processes a comma-separated environment variable.
func getValidEntries[T comparable](envVar string, parseFunc func(string) (T, error)) []T {
	rawValue := os.Getenv(envVar)
	if rawValue == "" {
		return []T{}
	}

	entries := strings.FieldsFunc(rawValue, func(r rune) bool {
		return r == ','
	})

	validEntries := make([]T, 0, len(entries))
	for _, entry := range entries {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}

		parsedValue, err := parseFunc(trimmed)
		if err == nil {
			validEntries = append(validEntries, parsedValue)
		}
	}

	return validEntries
}

// Helper functions for parsing
func parseString(value string) (string, error) {
	return value, nil
}

func parseInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}
