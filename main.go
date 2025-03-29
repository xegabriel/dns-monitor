// Critical DNS records to monitor for security:
// - MX: Routes emails to iCloud; changes can hijack mail flow.
// - SPF: Defines authorized senders; unauthorized changes enable spoofing.
// - DKIM: Cryptographically signs emails; modifications allow forged emails.
// - DMARC: Enforces email authentication policies; weakening it enables phishing.
// - A: Monitor extra non-email related record to prevent unintended IP changes.
// Monitor these records and enable alerts to detect unauthorized modifications.

package main

import (
	"context"
	"dns-monitor/internal/common"
	"dns-monitor/internal/dns"
	"dns-monitor/internal/notifications"
	"dns-monitor/internal/storage"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Set up logging with timestamps
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting DNS monitor service")
	ctx := context.Background()

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create a channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Load previous state
	prevState, err := storage.LoadPreviousState()
	if err != nil {
		log.Printf("Warning: Could not load previous state: %v", err)
		// Continue with empty state instead of failing
		prevState = common.PreviousState{Records: []common.DNSRecord{}}

		// Optionally notify about the issue
		if config.NotifyOnErrors {
			sendErrorNotification(ctx, config, "Failed to load previous state", err)
		}
	}

	// Use a ticker for consistent timing
	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	// Run the first check synchronously
	performCheck(ctx, config, &prevState)

	// Main loop with signal handling
	for {
		select {
		case <-ticker.C:
			// Perform the check synchronously on each tick
			performCheck(ctx, config, &prevState)

		case <-stop:
			log.Println("Received shutdown signal, exiting gracefully")
			return
		}
	}
}

// Perform DNS check and handle notifications
func performCheck(ctx context.Context, config common.Config, prevState *common.PreviousState) {

	log.Println("Checking DNS records...")
	currentRecords, err := dns.FetchDNSRecords(ctx, config)
	if err != nil {
		log.Printf("Error fetching DNS records: %v", err)
		if config.NotifyOnErrors {
			sendErrorNotification(ctx, config, "Failed to fetch DNS records", err)
		}
		return
	}

	changes := dns.DetectChanges(prevState.Records, currentRecords)
	if len(changes) > 0 {
		sendChangeDetectedNotification(ctx, config, changes)

		// Update stored state after alerting
		err = storage.SavePreviousState(common.PreviousState{Records: currentRecords})
		if err != nil {
			log.Printf("Failed to save updated state: %v", err)
			if config.NotifyOnErrors {
				sendErrorNotification(ctx, config, "Failed to save updated state", err)
			}
		}

		// Update in-memory state regardless of storage success
		*prevState = common.PreviousState{Records: currentRecords}
	} else {
		log.Println("No DNS changes detected")
	}
}

// Send notification about changes
func sendChangeDetectedNotification(ctx context.Context, config common.Config, changes []string) {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("⚠️ DNS CHANGES DETECTED for %s ⚠️\n\n", config.Domain))

	for i, change := range changes {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, change))
	}

	builder.WriteString(fmt.Sprintf("\nDetected at: %s", time.Now().Format(time.RFC1123)))

	message := builder.String()
	log.Println(message)
	err := notifications.SendPushoverNotification(
		ctx,
		config,
		"DNS Change Alert",
		message,
	)
	if err != nil {
		log.Printf("Error sending Pushover notification: %v", err)
	} else {
		log.Println("Pushover notification sent successfully")
	}
}

// Send notification about internal errors
func sendErrorNotification(ctx context.Context, config common.Config, subject string, err error) {

	message := fmt.Sprintf("DNS Monitor Error: %s\n\nError details: %v\n\nTime: %s",
		subject, err, time.Now().Format(time.RFC1123))

	err = notifications.SendPushoverNotification(
		ctx,
		config,
		"DNS Monitor Error",
		message,
	)

	if err != nil {
		log.Printf("Failed to send error notification: %v", err)
	}
}
