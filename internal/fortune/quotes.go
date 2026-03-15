// ABOUTME: Fortune system for PVM with inspirational Perl philosophy quotes
// ABOUTME: Provides random quote selection for shell initialization Easter egg

package fortune

import (
	"math/rand"
	"sync"
	"time"
)

var (
	// quotes contains inspirational Perl philosophy quotes for the initialization Easter egg
	quotes = []string{
		"Make easy things easy and hard things possible.",
		"Excellence doesn't need to be loud.",
		"Do what makes sense in context.",
		"Context matters more than you think.",
		"Text processing is in our DNA.",
		"Regular expressions are worth mastering.",
		"Laziness, impatience, and hubris are virtues.",
		"Perl grows with you - baby Perl is still Perl.",
		"Sigils add clarity, not confusion.",
		"JAPHs are fun, maintainability pays the bills.",
		"Expressiveness can beat uniformity.",
		"Documentation and tests are love letters to your future self.",
		"Warnings are your friend; strict is your mentor.",
		"Restraint without restriction.",
		"Magic is powerful - use it sparingly.",
		"Philosophy and practice enhance each other.",
		"Pragmatism and purity can coexist.",
		"The best code is the code that works for your problem.",
		"Community matters: TIMTOWTDI includes \"together.\"",
		"There's more than one way to do it.",
		"But sometimes consistency is not a bad thing either.",
	}

	// Thread-safe random number generator
	rngOnce sync.Once
	rng     *rand.Rand
)

// getRNG returns a thread-safe random number generator, initialized once
func getRNG() *rand.Rand {
	rngOnce.Do(func() {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
	return rng
}

// GetRandomQuote returns a randomly selected quote from the collection.
// It uses a thread-safe random number generator for randomization.
func GetRandomQuote() string {
	if len(quotes) == 0 {
		return "PVM environment initialized" // fallback message
	}

	rng := getRNG()
	index := rng.Intn(len(quotes))
	return quotes[index]
}

// GetQuoteWithSeed returns a quote selected using the provided seed.
// This function is primarily useful for testing with deterministic results.
func GetQuoteWithSeed(seed int64) string {
	if len(quotes) == 0 {
		return "PVM environment initialized" // fallback message
	}

	// Create a new RNG with the specific seed for deterministic testing
	rng := rand.New(rand.NewSource(seed))
	index := rng.Intn(len(quotes))
	return quotes[index]
}

// GetQuoteCount returns the total number of available quotes.
func GetQuoteCount() int {
	return len(quotes)
}

// GetAllQuotes returns all available quotes for testing purposes.
func GetAllQuotes() []string {
	result := make([]string, len(quotes))
	copy(result, quotes)
	return result
}
