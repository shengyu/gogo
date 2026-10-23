package main

import "time"

func fixtures() ([]Transaction, map[string]bool) {
	stamp := func(value string) string {
		parsed, _ := time.Parse(time.RFC3339, value)
		return parsed.UTC().Format(time.RFC3339)
	}
	rows := []Transaction{
		{"tx_001", "acct_001", "COMPLETED", 1599, "USD", stamp("2026-10-01T00:00:00Z")},
		{"tx_002", "acct_001", "PENDING", 2500, "USD", stamp("2026-10-01T10:00:00Z")},
		{"tx_003", "acct_001", "FAILED", 0, "USD", stamp("2026-10-02T10:00:00Z")},
		{"tx_004", "acct_001", "REFUNDED", 1599, "USD", stamp("2026-10-03T10:00:00Z")},
		{"tx_005", "acct_001", "COMPLETED", 8000000000000000000, "USD", stamp("2026-10-05T12:30:00Z")},
		{"tx_006", "acct_001", "COMPLETED", 0, "USD", stamp("2026-10-07T10:00:00Z")},
		{"tx_007", "acct_001", "PENDING", 4200, "USD", stamp("2026-10-09T10:00:00Z")},
		{"tx_008", "acct_001", "COMPLETED", 810, "USD", stamp("2026-10-12T10:00:00Z")},
		{"tx_009", "acct_001", "FAILED", 700, "USD", stamp("2026-10-15T10:00:00Z")},
		{"tx_010", "acct_001", "COMPLETED", 1234, "USD", stamp("2026-10-31T23:59:59Z")},
		{"tx_011", "acct_001", "REFUNDED", 340, "USD", stamp("2026-10-20T10:00:00Z")},
		{"tx_012", "acct_001", "COMPLETED", 5000, "USD", stamp("2026-10-25T10:00:00Z")},
		{"tx_013", "acct_001", "COMPLETED", 99, "USD", stamp("2026-10-29T10:00:00Z")},
		{"tx_014", "acct_001", "PENDING", 600, "USD", stamp("2026-10-30T10:00:00Z")},
		{"tx_015", "acct_001", "COMPLETED", 1000, "USD", stamp("2026-10-18T10:00:00Z")},
		{"tx_016", "acct_001", "FAILED", 30, "USD", stamp("2026-10-10T10:00:00Z")},
		{"tx_017", "acct_002", "COMPLETED", 7000, "USD", stamp("2026-10-01T00:00:00Z")},
		{"tx_018", "acct_002", "PENDING", 800, "USD", stamp("2026-10-04T10:00:00Z")},
		{"tx_019", "acct_002", "FAILED", 120, "USD", stamp("2026-10-11T10:00:00Z")},
		{"tx_020", "acct_002", "REFUNDED", 300, "USD", stamp("2026-10-22T10:00:00Z")},
		{"tx_021", "acct_002", "COMPLETED", 9900, "USD", stamp("2026-10-31T23:59:59Z")},
		{"tx_022", "acct_002", "COMPLETED", 50, "USD", stamp("2026-10-14T10:00:00Z")},
	}
	return rows, map[string]bool{"acct_001": true, "acct_002": true, "acct_empty": true}
}
