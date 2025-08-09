// ABOUTME: Test suite for fortune quote system
// ABOUTME: Ensures proper random selection and comprehensive quote coverage

package fortune

import (
	"strings"
	"testing"
)

func TestGetRandomQuote(t *testing.T) {
	quote := GetRandomQuote()

	if quote == "" {
		t.Error("GetRandomQuote() returned empty string")
	}

	// Verify the quote is one of our expected quotes
	found := false
	for _, expectedQuote := range quotes {
		if quote == expectedQuote {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("GetRandomQuote() returned unexpected quote: %s", quote)
	}
}

func TestGetQuoteWithSeed(t *testing.T) {
	// Test deterministic behavior with fixed seed
	seed := int64(12345)
	quote1 := GetQuoteWithSeed(seed)
	quote2 := GetQuoteWithSeed(seed)

	if quote1 != quote2 {
		t.Errorf("GetQuoteWithSeed() not deterministic: got %s and %s", quote1, quote2)
	}

	// Verify it returns a valid quote
	found := false
	for _, expectedQuote := range quotes {
		if quote1 == expectedQuote {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("GetQuoteWithSeed() returned unexpected quote: %s", quote1)
	}
}

func TestGetQuoteCount(t *testing.T) {
	expectedCount := 21 // As specified in the GitHub issue
	count := GetQuoteCount()

	if count != expectedCount {
		t.Errorf("GetQuoteCount() = %d, expected %d", count, expectedCount)
	}

	// Verify count matches actual slice length
	if count != len(quotes) {
		t.Errorf("GetQuoteCount() = %d, but quotes slice has length %d", count, len(quotes))
	}
}

func TestGetAllQuotes(t *testing.T) {
	allQuotes := GetAllQuotes()

	if len(allQuotes) != len(quotes) {
		t.Errorf("GetAllQuotes() returned %d quotes, expected %d", len(allQuotes), len(quotes))
	}

	// Verify all expected quotes are present
	for i, expectedQuote := range quotes {
		if allQuotes[i] != expectedQuote {
			t.Errorf("GetAllQuotes()[%d] = %s, expected %s", i, allQuotes[i], expectedQuote)
		}
	}

	// Verify it's a copy, not the original slice
	if &allQuotes[0] == &quotes[0] {
		t.Error("GetAllQuotes() returned reference to original slice instead of copy")
	}
}

func TestQuoteContent(t *testing.T) {
	// Verify specific quotes from the GitHub issue are present
	expectedQuotes := map[string]bool{
		"Make easy things easy and hard things possible.":               true,
		"There's more than one way to do it.":                           true,
		"But sometimes consistency is not a bad thing either.":          true,
		"Laziness, impatience, and hubris are virtues.":                 true,
		"Documentation and tests are love letters to your future self.": true,
		"Community matters: TIMTOWTDI includes \"together.\"":           true,
	}

	allQuotes := GetAllQuotes()
	for _, quote := range allQuotes {
		if expectedQuotes[quote] {
			delete(expectedQuotes, quote)
		}
	}

	if len(expectedQuotes) > 0 {
		missing := make([]string, 0, len(expectedQuotes))
		for quote := range expectedQuotes {
			missing = append(missing, quote)
		}
		t.Errorf("Missing expected quotes: %s", strings.Join(missing, ", "))
	}
}

func TestQuoteRandomness(t *testing.T) {
	// Test that different seeds produce different quotes (with high probability)
	results := make(map[string]int)

	// Generate quotes with different seeds
	for seed := int64(0); seed < 100; seed++ {
		quote := GetQuoteWithSeed(seed)
		results[quote]++
	}

	// We should get multiple different quotes
	if len(results) < 3 {
		t.Errorf("Expected multiple different quotes, got %d unique quotes", len(results))
	}

	// No single quote should dominate completely
	for quote, count := range results {
		if count > 50 { // More than 50% is suspicious for random selection
			t.Errorf("Quote appears too frequently (%d/100): %s", count, quote)
		}
	}
}

func TestQuoteCompleteness(t *testing.T) {
	// Verify we can get every quote through seeded selection
	quotesSeen := make(map[string]bool)

	// Try many different seeds to cover all quotes
	for seed := int64(0); seed < 1000; seed++ {
		quote := GetQuoteWithSeed(seed)
		quotesSeen[quote] = true
	}

	// Check if we've seen all quotes
	allQuotes := GetAllQuotes()
	for _, expectedQuote := range allQuotes {
		if !quotesSeen[expectedQuote] {
			t.Errorf("Quote never appeared in 1000 random selections: %s", expectedQuote)
		}
	}
}
