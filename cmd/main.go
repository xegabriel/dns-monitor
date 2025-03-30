package main

import (
	"context"
	"dns-monitor/internal/common"
	"dns-monitor/internal/dns"
	"dns-monitor/internal/notification"
	"dns-monitor/internal/notification/providers"
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

	factory := notification.NewFactory(&config.NotificationConfig)
	notifier, err := factory.CreateNotifier()
	if err != nil {
		log.Fatalf("Failed to create notifier: %v", err)
	}
	log.Printf("🔔 Notifier %s created successfully 🔔", config.NotificationConfig.NotifierType)

	// Create a channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Load previous state
	prevState, err := storage.LoadPreviousState(config.Domain)
	if err != nil {
		log.Printf("Warning: Could not load previous state: %v", err)
		// Continue with empty state instead of failing
		prevState = common.PreviousState{Records: []common.DNSRecord{}}

		// Optionally notify about the issue
		if config.NotifyOnErrors {
			sendErrorNotification(ctx, notifier, "Failed to load previous state", err)
		}
	}

	// Use a ticker for consistent timing
	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	// Run the first check synchronously
	performCheck(ctx, notifier, *config, &prevState)

	// Main loop with signal handling
	for {
		select {
		case <-ticker.C:
			// Perform the check synchronously on each tick
			performCheck(ctx, notifier, *config, &prevState)

		case <-stop:
			log.Println("Received shutdown signal, exiting gracefully")
			return
		}
	}
}

// Perform DNS check and handle notifications
func performCheck(ctx context.Context, notifier providers.Notifier, config common.Config, prevState *common.PreviousState) {

	log.Printf("⏳ Checking DNS records for %s every %s ... ⏳", config.Domain, config.CheckInterval)
	currentRecords, err := dns.FetchDNSRecords(ctx, config)
	if err != nil {
		log.Printf("Error fetching DNS records: %v", err)
		if config.NotifyOnErrors {
			sendErrorNotification(ctx, notifier, "Failed to fetch DNS records", err)
		}
		return
	}

	log.Printf("⚖️ Fetched %d DNS records. Comparing them with the previous %d records... ⚖️", len(currentRecords), len(prevState.Records))
	changes := dns.DetectChanges(prevState.Records, currentRecords)
	if len(changes) > 0 {
		sendChangeDetectedNotification(ctx, notifier, config.Domain, changes)

		// Update stored state after alerting
		err = storage.SavePreviousState(common.PreviousState{Records: currentRecords}, config.Domain)
		if err != nil {
			log.Printf("Failed to save updated state: %v", err)
			if config.NotifyOnErrors {
				sendErrorNotification(ctx, notifier, "Failed to save updated state", err)
			}
		}

		// Update in-memory state regardless of storage success
		*prevState = common.PreviousState{Records: currentRecords}
	} else {
		log.Println("✅ No DNS changes detected ✅")
	}
}

// Send notification about changes
func sendChangeDetectedNotification(ctx context.Context, notifier providers.Notifier, domain string, changes []string) {

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("⚠️ DNS CHANGES DETECTED for %s ⚠️\n\n", domain))

	for i, change := range changes {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, change))
	}

	builder.WriteString(fmt.Sprintf("\nDetected at: %s", time.Now().Format(time.RFC1123)))

	message := builder.String()
	log.Println(message)

	err := notifier.SendNotification(ctx, "DNS Change Alert", message)
	if err != nil {
		log.Printf("❌ Error sending notification: %v ❌", err)
	} else {
		log.Println("✅ Notification sent successfully ✅")
	}
}

// Send notification about internal errors
func sendErrorNotification(ctx context.Context, notifier providers.Notifier, subject string, err error) {

	message := fmt.Sprintf("❌ DNS Monitor Error: %s\n\nError details: %v\n\nTime: %s ❌",
		subject, err, time.Now().Format(time.RFC1123))

	err = notifier.SendNotification(ctx, "DNS Monitor Error", message)
	if err != nil {
		log.Printf("Failed to send error notification: %v", err)
	} else {
		log.Println("Error notification sent successfully")
	}

	if err != nil {
		log.Printf("Failed to send error notification: %v", err)
	}
}
