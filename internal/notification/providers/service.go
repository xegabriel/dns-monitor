package providers

import (
	"context"
	"dns-monitor/internal/common"
	"time"

	"github.com/nikoksr/notify"
)

// TODO: Add retries
const (
	maxRetries   = 5
	initialDelay = 500 * time.Millisecond
)

// Notifier is the interface for sending notifications
type Notifier interface {
	SendNotification(ctx context.Context, title, message string) error
}

// Service represents a notification service
type Service struct {
	notifier *notify.Notify
}

// NewService creates a new notification service
func NewService(n *notify.Notify) *Service {
	return &Service{
		notifier: n,
	}
}

// SendNotification sends a notification with the given title and message
func (s *Service) SendNotification(ctx context.Context, title, message string) error {
	operation := func() error {
		title = sanitizeString(title, 250)
		message = sanitizeString(message, 1024)
		return s.notifier.Send(ctx, title, message)
	}
	return common.RetryWithExponentialBackoff(ctx, maxRetries, initialDelay, operation)
}

func sanitizeString(s string, maxLength int) string {
	if len(s) > maxLength {
		return s[:maxLength-3] + "..."
	}
	return s
}
