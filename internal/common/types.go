package common

import (
	"net/http"
	"time"

	"github.com/miekg/dns"
)

type PreviousState struct {
	Records []DNSRecord `json:"records"`
}

type DNSRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   uint32 `json:"ttl"`
}

// Configuration struct to hold all settings
type Config struct {
	Domain              string
	CustomDomains       []string
	CustomDkimSelectors []string
	DNSServer           string
	DNSClient           dns.Client
	CheckInterval       time.Duration
	HTTPClient          *http.Client
	NotifyOnErrors      bool
	NotificationConfig  NotificationConfig
}

type NotificationConfig struct {
	NotifierType string
	// Pushover configuration
	PushoverToken string
	PushoverUser  string
	// Telegram configuration
	TelegramBotToken string
	TelegramChatIDs  []int64
}

// Notifier types
const (
	NotifierTypePushover = "pushover"
	NotifierTypeTelegram = "telegram"
	NotifierTypeSlack    = "slack"
	NotifierTypeEmail    = "email"
	// Add more notifier types as needed
)

var NotifierTypes = []string{
	NotifierTypePushover,
	NotifierTypeTelegram,
	NotifierTypeSlack,
	NotifierTypeEmail,
	// Add more notifier types as needed
}
