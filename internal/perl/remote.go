// ABOUTME: Remote configuration types for custom Perl fork sources
// ABOUTME: Provides Remote struct, RemoteList with CRUD operations, validation, and merge support

package perl

import (
	"fmt"
	"regexp"

	"tamarou.com/pvm/internal/errors"
)

// Remote error codes
const (
	ErrInvalidRemote  = "204" // Invalid remote configuration
	ErrRemoteExists   = "205" // Remote already exists
	ErrRemoteNotFound = "206" // Remote not found
	ErrReservedRemote = "207" // Reserved remote name
)

// Remote represents a configured fork source
type Remote struct {
	Name string `toml:"name" json:"name"`
	URL  string `toml:"url" json:"url"`
	Type string `toml:"type" json:"type"`
}

// remoteNamePattern validates remote names: [a-z0-9][a-z0-9-]*
var remoteNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Validate checks that the remote configuration is valid
func (r *Remote) Validate() error {
	if r.Name == "" {
		return errors.NewVersionError(ErrInvalidRemote, "remote name cannot be empty", nil)
	}
	if !remoteNameRe.MatchString(r.Name) {
		return errors.NewVersionError(ErrInvalidRemote,
			fmt.Sprintf("invalid remote name %q: must match [a-z0-9][a-z0-9-]*", r.Name), nil)
	}
	if r.Name == "origin" {
		return errors.NewVersionError(ErrReservedRemote, "'origin' is reserved and cannot be used as a remote name", nil)
	}
	if r.URL == "" {
		return errors.NewVersionError(ErrInvalidRemote, "remote URL cannot be empty", nil)
	}
	return nil
}

// RemoteList manages a collection of remotes
type RemoteList struct {
	remotes []Remote
}

// NewRemoteList creates an empty RemoteList
func NewRemoteList() *RemoteList {
	return &RemoteList{}
}

// Add adds a remote to the list after validation
func (rl *RemoteList) Add(r Remote) error {
	// Default type to "git"
	if r.Type == "" {
		r.Type = "git"
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if _, ok := rl.Find(r.Name); ok {
		return errors.NewVersionError(ErrRemoteExists,
			fmt.Sprintf("remote %q already exists", r.Name), nil)
	}
	rl.remotes = append(rl.remotes, r)
	return nil
}

// Remove removes a remote by name
func (rl *RemoteList) Remove(name string) error {
	if name == "origin" {
		return errors.NewVersionError(ErrReservedRemote, "'origin' is reserved and cannot be removed", nil)
	}
	for i, r := range rl.remotes {
		if r.Name == name {
			rl.remotes = append(rl.remotes[:i], rl.remotes[i+1:]...)
			return nil
		}
	}
	return errors.NewVersionError(ErrRemoteNotFound,
		fmt.Sprintf("remote %q not found", name), nil)
}

// Find looks up a remote by name
func (rl *RemoteList) Find(name string) (Remote, bool) {
	for _, r := range rl.remotes {
		if r.Name == name {
			return r, true
		}
	}
	return Remote{}, false
}

// All returns a copy of all remotes
func (rl *RemoteList) All() []Remote {
	result := make([]Remote, len(rl.remotes))
	copy(result, rl.remotes)
	return result
}

// MergeRemoteLists merges two remote lists. Remotes from the override list
// take precedence over remotes with the same name in the base list.
func MergeRemoteLists(base, override *RemoteList) *RemoteList {
	merged := NewRemoteList()

	// Add all from base
	for _, r := range base.remotes {
		merged.remotes = append(merged.remotes, r)
	}

	// Add or replace from override
	for _, r := range override.remotes {
		found := false
		for i, existing := range merged.remotes {
			if existing.Name == r.Name {
				merged.remotes[i] = r
				found = true
				break
			}
		}
		if !found {
			merged.remotes = append(merged.remotes, r)
		}
	}

	return merged
}
