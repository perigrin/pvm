// ABOUTME: Setup for PVM end-to-end tests
// ABOUTME: Contains common setup and initialization for tests

package e2e

import (
	"math/rand"
	"strings"
	"time"
)

// init initializes the random number generator
func init() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
}

// generateRandomID generates a random string ID for test purposes
func generateRandomID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := strings.Builder{}

	for i := 0; i < 8; i++ {
		result.WriteByte(chars[rand.Intn(len(chars))])
	}

	return result.String()
}
