// ABOUTME: Tests for TTY detection and color mode logic.
// ABOUTME: Verifies NO_COLOR, CLICOLOR, CLICOLOR_FORCE environment variable handling.

package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/cli/ui"
)

func TestDetectColorModeDefault(t *testing.T) {
	// Clear all color-related env vars to test default behavior.
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("CLICOLOR")
	os.Unsetenv("CLICOLOR_FORCE")

	mode := detectColorMode()
	assert.Equal(t, ui.ColorAuto, mode)
}

func TestDetectColorModeNoColor(t *testing.T) {
	// NO_COLOR (any value, including empty) disables color.
	orig, had := os.LookupEnv("NO_COLOR")
	defer func() {
		if had {
			os.Setenv("NO_COLOR", orig)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	os.Setenv("NO_COLOR", "")
	assert.Equal(t, ui.ColorNever, detectColorMode())

	os.Setenv("NO_COLOR", "1")
	assert.Equal(t, ui.ColorNever, detectColorMode())
}

func TestDetectColorModeClicolorForce(t *testing.T) {
	// CLICOLOR_FORCE=1 forces color even when not a TTY.
	origNC, hadNC := os.LookupEnv("NO_COLOR")
	origCF, hadCF := os.LookupEnv("CLICOLOR_FORCE")
	defer func() {
		if hadNC {
			os.Setenv("NO_COLOR", origNC)
		} else {
			os.Unsetenv("NO_COLOR")
		}
		if hadCF {
			os.Setenv("CLICOLOR_FORCE", origCF)
		} else {
			os.Unsetenv("CLICOLOR_FORCE")
		}
	}()

	os.Unsetenv("NO_COLOR")
	os.Setenv("CLICOLOR_FORCE", "1")
	assert.Equal(t, ui.ColorAlways, detectColorMode())
}

func TestDetectColorModeNoColorTakesPrecedence(t *testing.T) {
	// NO_COLOR takes precedence over CLICOLOR_FORCE.
	origNC, hadNC := os.LookupEnv("NO_COLOR")
	origCF, hadCF := os.LookupEnv("CLICOLOR_FORCE")
	defer func() {
		if hadNC {
			os.Setenv("NO_COLOR", origNC)
		} else {
			os.Unsetenv("NO_COLOR")
		}
		if hadCF {
			os.Setenv("CLICOLOR_FORCE", origCF)
		} else {
			os.Unsetenv("CLICOLOR_FORCE")
		}
	}()

	os.Setenv("NO_COLOR", "1")
	os.Setenv("CLICOLOR_FORCE", "1")
	assert.Equal(t, ui.ColorNever, detectColorMode())
}

func TestDetectColorModeClicolorDisabled(t *testing.T) {
	// CLICOLOR=0 disables color.
	origNC, hadNC := os.LookupEnv("NO_COLOR")
	origCF, hadCF := os.LookupEnv("CLICOLOR_FORCE")
	origCC, hadCC := os.LookupEnv("CLICOLOR")
	defer func() {
		if hadNC {
			os.Setenv("NO_COLOR", origNC)
		} else {
			os.Unsetenv("NO_COLOR")
		}
		if hadCF {
			os.Setenv("CLICOLOR_FORCE", origCF)
		} else {
			os.Unsetenv("CLICOLOR_FORCE")
		}
		if hadCC {
			os.Setenv("CLICOLOR", origCC)
		} else {
			os.Unsetenv("CLICOLOR")
		}
	}()

	os.Unsetenv("NO_COLOR")
	os.Unsetenv("CLICOLOR_FORCE")
	os.Setenv("CLICOLOR", "0")
	assert.Equal(t, ui.ColorNever, detectColorMode())
}

func TestIsTerminalWithRegularFile(t *testing.T) {
	// A regular file is not a terminal.
	f, err := os.CreateTemp("", "tty_test")
	if err != nil {
		t.Skip("cannot create temp file")
	}
	defer os.Remove(f.Name())
	defer f.Close()

	assert.False(t, isTerminal(f))
}
