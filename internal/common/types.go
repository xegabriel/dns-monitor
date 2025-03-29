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
	PushoverToken       string
	PushoverUser        string
	NotifyOnErrors      bool
}
