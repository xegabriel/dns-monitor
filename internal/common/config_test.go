package common

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables to restore later.
	origEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, kv := range origEnv {
			// Parse key and value from each entry.
			var key, value string
			fmt.Sscanf(kv, "%[^=]=%s", &key, &value)
			os.Setenv(key, value)
		}
	}()

	tests := []struct {
		name          string
		env           map[string]string
		expectError   bool
		errorContains string
		checkInterval time.Duration // expected check interval if valid
		dnsServer     string        // expected DNS server if not set in env
		notifyErrors  bool          // expected value for NotifyOnErrors
	}{
		{
			name:          "Missing DOMAIN",
			env:           map[string]string{},
			expectError:   true,
			errorContains: "DOMAIN environment variable is required",
		},
		{
			name: "Notification config error",
			env: map[string]string{
				"DOMAIN":        "test.ro",
				"NOTIFIER_TYPE": NotifierTypePushover, // but missing Pushover tokens
			},
			expectError:   true,
			errorContains: "failed to load notification config",
		},
		{
			name: "Invalid CHECK_INTERVAL",
			env: map[string]string{
				"DOMAIN":            "test.ro",
				"NOTIFIER_TYPE":     NotifierTypePushover,
				PushoverAppTokenEnv: "app_token",
				PushoverUserKeyEnv:  "user_key",
				"CHECK_INTERVAL":    "invalid",
			},
			expectError:   true,
			errorContains: "invalid CHECK_INTERVAL format",
		},
		{
			name: "Valid config with defaults",
			env: map[string]string{
				"DOMAIN":            "test.ro",
				"NOTIFIER_TYPE":     NotifierTypePushover,
				PushoverAppTokenEnv: "app_token",
				PushoverUserKeyEnv:  "user_key",
				// DNS_SERVER is not set so default should apply.
				// CHECK_INTERVAL not set so default should be 1h.
				"NOTIFY_ON_ERRORS":      "true",
				"CUSTOM_SUBDOMAINS":     "sub1",
				"CUSTOM_DKIM_SELECTORS": "selector1",
			},
			expectError:   false,
			checkInterval: 1 * time.Hour,
			dnsServer:     "1.1.1.1:53",
			notifyErrors:  true,
		},
		{
			name: "Valid config with custom CHECK_INTERVAL and DNS_SERVER",
			env: map[string]string{
				"DOMAIN":                "example.com",
				"NOTIFIER_TYPE":         NotifierTypeTelegram,
				TelegramBotTokenEnv:     "bot_token",
				TelegramChatIDsEnv:      "12345",
				"CHECK_INTERVAL":        "30m",
				"DNS_SERVER":            "8.8.8.8:53",
				"NOTIFY_ON_ERRORS":      "false",
				"CUSTOM_SUBDOMAINS":     "sub2",
				"CUSTOM_DKIM_SELECTORS": "selector2",
			},
			expectError:   false,
			checkInterval: 30 * time.Minute,
			dnsServer:     "8.8.8.8:53",
			notifyErrors:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			// Set environment variables for test case.
			for key, val := range tc.env {
				os.Setenv(key, val)
			}

			cfg, err := LoadConfig()
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				// Check required values.
				if cfg.Domain != tc.env["DOMAIN"] {
					t.Errorf("expected domain %q, got %q", tc.env["DOMAIN"], cfg.Domain)
				}
				if cfg.DNSServer != tc.dnsServer {
					t.Errorf("expected DNS server %q, got %q", tc.dnsServer, cfg.DNSServer)
				}
				if cfg.CheckInterval != tc.checkInterval {
					t.Errorf("expected check interval %v, got %v", tc.checkInterval, cfg.CheckInterval)
				}
				if cfg.NotifyOnErrors != tc.notifyErrors {
					t.Errorf("expected NotifyOnErrors %v, got %v", tc.notifyErrors, cfg.NotifyOnErrors)
				}
				// Validate custom subdomains and selectors (dummy parser returns a slice with the raw string).
				if len(cfg.CustomSubdomains) == 0 || cfg.CustomSubdomains[0] != tc.env["CUSTOM_SUBDOMAINS"] {
					t.Errorf("expected custom subdomains %v, got %v", []string{tc.env["CUSTOM_SUBDOMAINS"]}, cfg.CustomSubdomains)
				}
				if len(cfg.CustomDkimSelectors) == 0 || cfg.CustomDkimSelectors[0] != tc.env["CUSTOM_DKIM_SELECTORS"] {
					t.Errorf("expected custom DKIM selectors %v, got %v", []string{tc.env["CUSTOM_DKIM_SELECTORS"]}, cfg.CustomDkimSelectors)
				}
			}
		})
	}
}

func TestLoadNotificationConfig(t *testing.T) {
	// save original environment variables so we can restore later
	origEnv := os.Environ()
	defer func() {
		for _, kv := range origEnv {
			// split into key and value
			var key, value string
			fmt.Sscanf(kv, "%[^=]=%s", &key, &value)
			os.Setenv(key, value)
		}
	}()

	tests := []struct {
		name          string
		env           map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Missing NOTIFIER_TYPE",
			env:           map[string]string{},
			expectError:   true,
			errorContains: "NOTIFIER_TYPE environment variable is required",
		},
		{
			name: "Invalid NOTIFIER_TYPE",
			env: map[string]string{
				"NOTIFIER_TYPE": "invalid",
			},
			expectError:   true,
			errorContains: "unsupported notifier type: invalid",
		},
		{
			name: "Pushover missing token",
			env: map[string]string{
				"NOTIFIER_TYPE":    NotifierTypePushover,
				PushoverUserKeyEnv: "user_key",
			},
			expectError:   true,
			errorContains: PushoverAppTokenEnv,
		},
		{
			name: "Pushover missing user",
			env: map[string]string{
				"NOTIFIER_TYPE":     NotifierTypePushover,
				PushoverAppTokenEnv: "app_token",
			},
			expectError:   true,
			errorContains: PushoverUserKeyEnv,
		},
		{
			name: "Pushover valid config",
			env: map[string]string{
				"NOTIFIER_TYPE":     NotifierTypePushover,
				PushoverAppTokenEnv: "app_token",
				PushoverUserKeyEnv:  "user_key",
			},
			expectError: false,
		},
		{
			name: "Telegram missing bot token",
			env: map[string]string{
				"NOTIFIER_TYPE":    NotifierTypeTelegram,
				TelegramChatIDsEnv: "12345",
			},
			expectError:   true,
			errorContains: TelegramBotTokenEnv,
		},
		{
			name: "Telegram missing chat ids",
			env: map[string]string{
				"NOTIFIER_TYPE":     NotifierTypeTelegram,
				TelegramBotTokenEnv: "bot_token",
			},
			expectError:   true,
			errorContains: TelegramChatIDsEnv,
		},
		{
			name: "Telegram valid config",
			env: map[string]string{
				"NOTIFIER_TYPE":     NotifierTypeTelegram,
				TelegramBotTokenEnv: "bot_token",
				TelegramChatIDsEnv:  "12345",
			},
			expectError: false,
		},
		{
			name: "Unsupported notifier type",
			env: map[string]string{
				"NOTIFIER_TYPE": "unsupported",
			},
			expectError:   true,
			errorContains: "unsupported notifier type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// clear relevant env variables first
			os.Clearenv()
			for key, val := range tc.env {
				os.Setenv(key, val)
			}

			cfg, err := loadNotificationConfig()
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				if !contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg == nil {
					t.Fatal("expected a valid config, got nil")
				}
				// additional assertions for valid config can be added here
			}
		})
	}
}

func TestGetValidEntries_Int(t *testing.T) {
	envVar := "TEST_INT"
	tests := []struct {
		name     string
		envValue string
		expected []int64
	}{
		{
			name:     "empty environment variable returns empty slice",
			envValue: "",
			expected: []int64{},
		},
		{
			name:     "valid comma-separated integers with spaces",
			envValue: "1, 2,3 , 4",
			expected: []int64{1, 2, 3, 4},
		},
		{
			name:     "invalid entries are skipped",
			envValue: "a, 2, three, 4",
			expected: []int64{2, 4},
		},
		{
			name:     "handles extra commas and spaces",
			envValue: ", 1, 2,, 3,",
			expected: []int64{1, 2, 3},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the environment variable for the test.
			os.Setenv(envVar, tc.envValue)
			// Call the generic function.
			result := getValidEntries(envVar, parseInt64)
			// Compare the result with the expected slice.
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestGetValidEntries_String(t *testing.T) {
	envVar := "TEST_STRING"
	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "empty environment variable returns empty slice",
			envValue: "",
			expected: []string{},
		},
		{
			name:     "valid comma-separated strings with spaces",
			envValue: "apple, banana, cherry ",
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name:     "handles extra commas and spaces",
			envValue: " , dog, cat,, elephant, ",
			expected: []string{"dog", "cat", "elephant"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set the environment variable for the test.
			os.Setenv(envVar, tc.envValue)
			// Call the generic function.
			result := getValidEntries(envVar, parseString)
			// Trim any possible spaces from the expected results (shouldn't be necessary, but just to be sure).
			for i, v := range tc.expected {
				tc.expected[i] = strings.TrimSpace(v)
			}
			// Compare the result with the expected slice.
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// contains is a simple helper function to check if a substring exists in a string.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (len(substr) == 0 || (s != "" && substr != "" && (s[0:len(substr)] == substr || contains(s[1:], substr)))))
}
