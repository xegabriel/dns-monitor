package dns

import (
	"context"
	"dns-monitor/internal/common"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	retries      = 3
	initialDelay = 500 * time.Millisecond
)

// FetchDNSRecords fetches DNS records for a domain from a specific DNS server.
func FetchDNSRecords(ctx context.Context, config common.Config) ([]common.DNSRecord, error) {
	var allRecords []common.DNSRecord

	// Record types to check - focusing on email-related records.
	recordTypes := []uint16{
		dns.TypeMX,    // Mail exchange records.
		dns.TypeTXT,   // Text records (SPF, DKIM, DMARC).
		dns.TypeCNAME, // CNAME records (could be used for mail subdomains).
		dns.TypeA,
	}

	// Generate the list of domains to check.
	domainsToCheck := generateDomainsToCheck(config)

	// Iterate through each domain and record type.
	for _, domainName := range domainsToCheck {
		for _, recordType := range recordTypes {
			r, err := queryDNS(ctx, domainName, recordType, config)
			if err != nil {
				log.Printf("Error querying %s for %s: %v", domainName, dns.TypeToString[recordType], err)
				continue
			}
			if r.Rcode != dns.RcodeSuccess {
				continue
			}

			for _, ans := range r.Answer {
				record := parseDNSRecord(ans, recordType)
				allRecords = append(allRecords, record)
			}
		}
	}

	return allRecords, nil
}

// generateDomainsToCheck builds the list of domains and subdomains to query.
func generateDomainsToCheck(config common.Config) []string {
	domain := config.Domain
	domains := []string{
		domain,                               // Main domain
		fmt.Sprintf("_dmarc.%s", domain),     // DMARC policy
		fmt.Sprintf("_domainkey.%s", domain), // DKIM policy
		fmt.Sprintf("www.%s", domain),        // www subdomain
	}

	// Add custom DKIM selectors.
	for _, selector := range config.CustomDkimSelectors {
		domains = append(domains, fmt.Sprintf("%s._domainkey.%s", selector, domain))
	}

	// Add any custom domains that were provided at runtime
	domains = append(domains, config.CustomDomains...)

	return domains
}

// queryDNS sends a DNS query for the given domain name and record type, with exponential backoff retries.
func queryDNS(ctx context.Context, domainName string, recordType uint16, config common.Config) (*dns.Msg, error) {
	var resp *dns.Msg

	operation := func() error {
		m := new(dns.Msg)
		m.SetQuestion(dns.Fqdn(domainName), recordType)
		m.RecursionDesired = true

		var err error
		resp, _, err = config.DNSClient.Exchange(m, config.DNSServer)
		if err != nil {
			return fmt.Errorf("DNS query failed: %w", err)
		}
		if resp.Rcode != dns.RcodeSuccess {
			return fmt.Errorf("DNS query returned non-success response: %d", resp.Rcode)
		}
		return nil
	}

	// Retry with exponential backoff.
	err := common.RetryWithExponentialBackoff(ctx, retries, initialDelay, operation)
	return resp, err
}

// parseDNSRecord converts a dns.RR into a types.DNSRecord based on its type.
func parseDNSRecord(rr dns.RR, recordType uint16) common.DNSRecord {
	record := common.DNSRecord{
		Type: dns.TypeToString[recordType],
		Name: rr.Header().Name,
		TTL:  rr.Header().Ttl,
	}

	switch v := rr.(type) {
	case *dns.MX:
		record.Value = fmt.Sprintf("%d %s", v.Preference, v.Mx)
	case *dns.TXT:
		record.Value = strings.Join(v.Txt, " ")
	case *dns.CNAME:
		record.Value = v.Target
	case *dns.A:
		record.Value = v.A.String()
	case *dns.AAAA:
		record.Value = v.AAAA.String()
	default:
		record.Value = rr.String()
	}

	return record
}

// DetectChanges identifies differences between two sets of DNS records
func DetectChanges(oldRecords, newRecords []common.DNSRecord) []string {
	// Build record maps
	oldRecordMap := buildRecordMap(oldRecords)
	newRecordMap := buildRecordMap(newRecords)

	// Collect all changes
	var changes []string
	changes = append(changes, detectAddedAndModifiedRecords(oldRecordMap, newRecordMap)...)
	changes = append(changes, detectDeletedRecords(oldRecordMap, newRecordMap)...)

	// Sort changes for consistent output
	sort.Strings(changes)
	return changes
}

// buildRecordMap creates a map of records grouped by Type:Name
func buildRecordMap(records []common.DNSRecord) map[string][]common.DNSRecord {
	recordMap := make(map[string][]common.DNSRecord)

	for _, record := range records {
		key := formatRecordKey(record.Type, record.Name)
		recordMap[key] = append(recordMap[key], record)
	}

	return recordMap
}

// formatRecordKey creates a consistent key format for the record maps
func formatRecordKey(recordType, recordName string) string {
	return fmt.Sprintf("%s:%s", recordType, recordName)
}

// detectAddedAndModifiedRecords finds new records and changed values
func detectAddedAndModifiedRecords(oldMap, newMap map[string][]common.DNSRecord) []string {
	var changes []string

	for key, newRecs := range newMap {
		oldRecs, exists := oldMap[key]

		if !exists {
			// All records in this group are new
			changes = append(changes, formatNewRecordChanges(newRecs)...)
		} else {
			// Check for added or removed values within this group
			changes = append(changes, detectValueChanges(oldRecs, newRecs)...)
		}
	}

	return changes
}

// formatNewRecordChanges creates change messages for newly added records
func formatNewRecordChanges(records []common.DNSRecord) []string {
	var changes []string

	for _, record := range records {
		changes = append(changes, fmt.Sprintf("NEW: %s %s -> %s",
			record.Type, record.Name, record.Value))
	}

	return changes
}

// detectValueChanges compares values within the same Type:Name record group
func detectValueChanges(oldRecs, newRecs []common.DNSRecord) []string {
	var changes []string

	// Create value lookup maps
	oldValueMap := createValueMap(oldRecs)
	newValueMap := createValueMap(newRecs)

	// Find added values
	for _, newRecord := range newRecs {
		if !oldValueMap[newRecord.Value] {
			changes = append(changes, fmt.Sprintf("ADDED: %s %s -> %s",
				newRecord.Type, newRecord.Name, newRecord.Value))
		}
	}

	// Find removed values
	for _, oldRecord := range oldRecs {
		if !newValueMap[oldRecord.Value] {
			changes = append(changes, fmt.Sprintf("REMOVED: %s %s -> %s",
				oldRecord.Type, oldRecord.Name, oldRecord.Value))
		}
	}

	return changes
}

// createValueMap creates a map of values for fast lookup
func createValueMap(records []common.DNSRecord) map[string]bool {
	valueMap := make(map[string]bool)

	for _, record := range records {
		valueMap[record.Value] = true
	}

	return valueMap
}

// detectDeletedRecords finds record groups that no longer exist
func detectDeletedRecords(oldMap, newMap map[string][]common.DNSRecord) []string {
	var changes []string

	for key, oldRecs := range oldMap {
		if _, exists := newMap[key]; !exists {
			for _, record := range oldRecs {
				changes = append(changes, fmt.Sprintf("DELETED: %s %s -> %s",
					record.Type, record.Name, record.Value))
			}
		}
	}

	return changes
}
