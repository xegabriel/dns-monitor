package notifications

import (
	"context"
	"dns-monitor/internal/common"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	pushoverAPIURL = "https://api.pushover.net/1/messages.json"
	maxRetries     = 5
	initialDelay   = 500 * time.Millisecond
)

func SendPushoverNotification(ctx context.Context, config common.Config, title, message string) error {
	// Validate parameters.
	if err := validatePushoverParameters(title, message); err != nil {
		return err
	}

	// Build request data.
	data := buildPushoverData(config, title, message)

	// Define the operation to retry.
	operation := func() error {
		req, err := http.NewRequestWithContext(ctx, "POST", pushoverAPIURL, strings.NewReader(data.Encode()))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := config.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("non-OK response from Pushover: %d - %s", resp.StatusCode, string(body))
		}
		return nil
	}

	return common.RetryWithExponentialBackoff(ctx, maxRetries, initialDelay, operation)
}

// validatePushoverParameters ensures that all required parameters are provided.
func validatePushoverParameters(title, message string) error {
	if title == "" || message == "" {
		return fmt.Errorf("all parameters must be provided")
	}
	return nil
}

// buildPushoverData constructs the form data for the Pushover request.
func buildPushoverData(config common.Config, title, message string) url.Values {
	data := url.Values{}
	data.Set("token", config.PushoverToken)
	data.Set("user", config.PushoverUser)
	data.Set("title", title)
	data.Set("message", message)
	data.Set("priority", "1") // High priority
	return data
}
