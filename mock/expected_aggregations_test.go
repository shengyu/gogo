package main

// Expected aggregation values are intentionally test-only server fixtures.
// They are not part of the public API documentation or server responses.
var expectedAggregationFixtures = map[string]map[string]int64{
	"acct_001": {"COMPLETED": 8000000000000009742, "PENDING": 7300, "FAILED": 730, "REFUNDED": 1939},
	"acct_002": {"COMPLETED": 16950, "PENDING": 800, "FAILED": 120, "REFUNDED": 300},
}
