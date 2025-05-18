// ABOUTME: Command registry for the CLI framework
// ABOUTME: Provides command registration and discovery

package cli

import (
	"github.com/spf13/cobra"
)

// CommandProvider defines a function that creates a command
type CommandProvider func() *cobra.Command

// CommandRegistry is a registry of commands
type CommandRegistry struct {
	commands map[string]CommandProvider
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]CommandProvider),
	}
}

// Register registers a command with the registry
func (r *CommandRegistry) Register(name string, provider CommandProvider) {
	r.commands[name] = provider
}

// Get returns a command provider by name
func (r *CommandRegistry) Get(name string) (CommandProvider, bool) {
	provider, ok := r.commands[name]
	return provider, ok
}

// CreateCommand creates a command by name
func (r *CommandRegistry) CreateCommand(name string) (*cobra.Command, bool) {
	provider, ok := r.Get(name)
	if !ok {
		return nil, false
	}

	return provider(), true
}

// GlobalRegistry is the global command registry
var GlobalRegistry = NewCommandRegistry()
