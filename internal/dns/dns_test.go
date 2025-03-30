package dns

import (
	"dns-monitor/internal/common"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

func TestGenerateDomainsToCheck(t *testing.T) {
	tests := []struct {
		name             string
		config           common.Config
		expectedDomains  []string
		notExpectedCount int // used to make sure there are no duplicates
	}{
		{
			name: "Basic domain with no custom selectors or subdomains",
			config: common.Config{
				Domain:              "example.com",
				CustomDkimSelectors: []string{},
				CustomSubdomains:    []string{},
			},
			expectedDomains: []string{
				"example.com",
				"_dmarc.example.com",
				"_domainkey.example.com",
				"www.example.com",
			},
			notExpectedCount: 4, // Should have exactly 4 domains
		},
		{
			name: "Domain with custom DKIM selectors",
			config: common.Config{
				Domain:              "example.com",
				CustomDkimSelectors: []string{"selector1", "selector2"},
				CustomSubdomains:    []string{},
			},
			expectedDomains: []string{
				"example.com",
				"_dmarc.example.com",
				"_domainkey.example.com",
				"www.example.com",
				"selector1._domainkey.example.com",
				"selector2._domainkey.example.com",
			},
			notExpectedCount: 6, // Should have exactly 6 domains
		},
		{
			name: "Domain with custom subdomains",
			config: common.Config{
				Domain:              "example.com",
				CustomDkimSelectors: []string{},
				CustomSubdomains:    []string{"mail", "blog"},
			},
			expectedDomains: []string{
				"example.com",
				"_dmarc.example.com",
				"_domainkey.example.com",
				"www.example.com",
				"mail.example.com",
				"blog.example.com",
			},
			notExpectedCount: 6, // Should have exactly 6 domains
		},
		{
			name: "Domain with duplicated subdomains",
			config: common.Config{
				Domain:              "example.com",
				CustomDkimSelectors: []string{"selector1"},
				CustomSubdomains:    []string{"www", "www"}, // Duplicate www
			},
			expectedDomains: []string{
				"example.com",
				"_dmarc.example.com",
				"_domainkey.example.com",
				"www.example.com",
				"selector1._domainkey.example.com",
			},
			notExpectedCount: 5, // Should have exactly 5 domains after deduplication
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDomainsToCheck(tt.config)

			// Check that all expected domains are present
			for _, expected := range tt.expectedDomains {
				found := false
				for _, actual := range result {
					if actual == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected domain %s not found in result", expected)
			}

			// Check that there are no duplicate domains
			assert.Equal(t, tt.notExpectedCount, len(result), "Unexpected number of domains after deduplication")
		})
	}
}

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "No duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "With duplicates",
			input:    []string{"a", "b", "a", "c", "b", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			name:     "Empty slice",
			input:    []string(nil),
			expected: []string(nil),
		},
		{
			name:     "All duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deduplicate(tt.input)

			// Sort both slices for consistent comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseDNSRecord(t *testing.T) {
	tests := []struct {
		name       string
		rr         dns.RR
		recordType uint16
		expected   common.DNSRecord
	}{
		{
			name: "MX Record",
			rr: &dns.MX{
				Hdr: dns.RR_Header{
					Name: "example.com.",
					Ttl:  3600,
				},
				Preference: 10,
				Mx:         "mail.example.com.",
			},
			recordType: dns.TypeMX,
			expected: common.DNSRecord{
				Type:  "MX",
				Name:  "example.com.",
				TTL:   3600,
				Value: "10 mail.example.com.",
			},
		},
		{
			name: "TXT Record",
			rr: &dns.TXT{
				Hdr: dns.RR_Header{
					Name: "example.com.",
					Ttl:  3600,
				},
				Txt: []string{"v=spf1", "include:_spf.example.com", "-all"},
			},
			recordType: dns.TypeTXT,
			expected: common.DNSRecord{
				Type:  "TXT",
				Name:  "example.com.",
				TTL:   3600,
				Value: "v=spf1 include:_spf.example.com -all",
			},
		},
		{
			name: "CNAME Record",
			rr: &dns.CNAME{
				Hdr: dns.RR_Header{
					Name: "www.example.com.",
					Ttl:  3600,
				},
				Target: "example.com.",
			},
			recordType: dns.TypeCNAME,
			expected: common.DNSRecord{
				Type:  "CNAME",
				Name:  "www.example.com.",
				TTL:   3600,
				Value: "example.com.",
			},
		},
		{
			name: "A Record",
			rr: &dns.A{
				Hdr: dns.RR_Header{
					Name: "example.com.",
					Ttl:  3600,
				},
				A: []byte{192, 0, 2, 1},
			},
			recordType: dns.TypeA,
			expected: common.DNSRecord{
				Type:  "A",
				Name:  "example.com.",
				TTL:   3600,
				Value: "192.0.2.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDNSRecord(tt.rr, tt.recordType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectChanges(t *testing.T) {
	tests := []struct {
		name       string
		oldRecords []common.DNSRecord
		newRecords []common.DNSRecord
		expected   []string
	}{
		{
			name: "No changes",
			oldRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
			},
			newRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
			},
			expected: []string(nil),
		},
		{
			name: "Added records",
			oldRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
			},
			newRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
			},
			expected: []string{
				"NEW: TXT example.com. -> v=spf1 -all",
			},
		},
		{
			name: "Deleted records",
			oldRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
			},
			newRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
			},
			expected: []string{
				"DELETED: TXT example.com. -> v=spf1 -all",
			},
		},
		{
			name: "Modified records",
			oldRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
			},
			newRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "20 mail2.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
			},
			expected: []string{
				"ADDED: MX example.com. -> 20 mail2.example.com.",
				"REMOVED: MX example.com. -> 10 mail.example.com.",
			},
		},
		{
			name: "Mixed changes",
			oldRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "10 mail.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
				{Type: "CNAME", Name: "www.example.com.", Value: "example.com."},
			},
			newRecords: []common.DNSRecord{
				{Type: "MX", Name: "example.com.", Value: "20 mail2.example.com."},
				{Type: "TXT", Name: "example.com.", Value: "v=spf1 include:_spf.google.com -all"},
				{Type: "TXT", Name: "_dmarc.example.com.", Value: "v=DMARC1; p=reject;"},
			},
			expected: []string{
				"ADDED: MX example.com. -> 20 mail2.example.com.",
				"ADDED: TXT example.com. -> v=spf1 include:_spf.google.com -all",
				"DELETED: CNAME www.example.com. -> example.com.",
				"NEW: TXT _dmarc.example.com. -> v=DMARC1; p=reject;",
				"REMOVED: MX example.com. -> 10 mail.example.com.",
				"REMOVED: TXT example.com. -> v=spf1 -all",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectChanges(tt.oldRecords, tt.newRecords)

			// Sort both slices for consistent comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildRecordMap(t *testing.T) {
	records := []common.DNSRecord{
		{Type: "MX", Name: "example.com.", Value: "10 mail1.example.com."},
		{Type: "MX", Name: "example.com.", Value: "20 mail2.example.com."},
		{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
	}

	expected := map[string][]common.DNSRecord{
		"MX:example.com.": {
			{Type: "MX", Name: "example.com.", Value: "10 mail1.example.com."},
			{Type: "MX", Name: "example.com.", Value: "20 mail2.example.com."},
		},
		"TXT:example.com.": {
			{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
		},
	}

	result := buildRecordMap(records)

	// Compare maps carefully since Go doesn't have a direct map comparison
	assert.Equal(t, len(expected), len(result), "Map size mismatch")

	for key, expectedRecords := range expected {
		resultRecords, ok := result[key]
		assert.True(t, ok, "Key %s not found in result map", key)

		// Compare slices
		assert.Equal(t, len(expectedRecords), len(resultRecords), "Record count mismatch for key %s", key)

		// For each expected record, find a matching record in the result
		for _, expRec := range expectedRecords {
			found := false
			for _, resRec := range resultRecords {
				if reflect.DeepEqual(expRec, resRec) {
					found = true
					break
				}
			}
			assert.True(t, found, "Record %v not found in result for key %s", expRec, key)
		}
	}
}

func TestFormatRecordKey(t *testing.T) {
	tests := []struct {
		recordType string
		recordName string
		expected   string
	}{
		{"MX", "example.com.", "MX:example.com."},
		{"TXT", "_dmarc.example.com.", "TXT:_dmarc.example.com."},
		{"CNAME", "www.example.com.", "CNAME:www.example.com."},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%s", tt.recordType, tt.recordName), func(t *testing.T) {
			result := formatRecordKey(tt.recordType, tt.recordName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateValueMap(t *testing.T) {
	records := []common.DNSRecord{
		{Type: "MX", Name: "example.com.", Value: "10 mail1.example.com."},
		{Type: "MX", Name: "example.com.", Value: "20 mail2.example.com."},
		{Type: "TXT", Name: "example.com.", Value: "v=spf1 -all"},
	}

	expected := map[string]bool{
		"10 mail1.example.com.": true,
		"20 mail2.example.com.": true,
		"v=spf1 -all":           true,
	}

	result := createValueMap(records)

	assert.Equal(t, expected, result)
}
