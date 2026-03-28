// ABOUTME: Fork identifier parsing and types for custom Perl fork management
// ABOUTME: Supports remote/fork@version grammar for identifying fork installations

package perl

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"tamarou.com/pvm/internal/errors"
)

// Fork error codes
const (
	ErrInvalidForkIdentifier = "201" // Invalid fork identifier format
	ErrReservedRemoteName    = "202" // Reserved remote name (e.g., "origin")
	ErrReservedForkName      = "203" // Reserved fork name (e.g., "perl")
)

// ForkIdentifier represents a parsed fork identifier with remote, fork name, and version
type ForkIdentifier struct {
	Remote      string // Remote name (e.g., "mycompany"), defaults to "origin"
	ForkName    string // Fork name (e.g., "myfork"), may be empty
	BaseVersion string // Base Perl version (e.g., "5.40.2")
}

// Validation patterns for remote and fork names
var (
	remoteNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	forkNamePattern   = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

// ParseForkIdentifier parses a fork identifier string into its components.
// Supported forms:
//   - "remote/fork@version" -> Remote=remote, ForkName=fork, BaseVersion=version
//   - "remote/version"      -> Remote=remote, ForkName="", BaseVersion=version
//   - "version"             -> Remote="origin", ForkName="", BaseVersion=version
func ParseForkIdentifier(input string) (*ForkIdentifier, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.NewVersionError(
			ErrInvalidForkIdentifier,
			"fork identifier cannot be empty",
			nil)
	}

	// Check for slash — distinguishes bare version from remote/... forms
	remote, rest, hasSlash := strings.Cut(input, "/")
	if !hasSlash {
		// Bare version form: "5.40.2"
		_, err := ParseVersion(input)
		if err != nil {
			return nil, errors.NewVersionError(
				ErrInvalidForkIdentifier,
				fmt.Sprintf("invalid version in fork identifier: %s", input),
				err)
		}
		return &ForkIdentifier{
			Remote:      "origin",
			ForkName:    "",
			BaseVersion: input,
		}, nil
	}

	// Validate remote name
	if !remoteNamePattern.MatchString(remote) {
		return nil, errors.NewVersionError(
			ErrInvalidForkIdentifier,
			fmt.Sprintf("invalid remote name %q: must match [a-z0-9][a-z0-9-]*", remote),
			nil)
	}

	// "origin" cannot appear as a remote prefix in identifiers
	if remote == "origin" {
		return nil, errors.NewVersionError(
			ErrReservedRemoteName,
			"'origin' cannot be used as a remote prefix in fork identifiers",
			nil)
	}

	// Check for @ — distinguishes remote/fork@version from remote/version
	forkName, version, hasAt := strings.Cut(rest, "@")
	if hasAt {

		if forkName == "" {
			return nil, errors.NewVersionError(
				ErrInvalidForkIdentifier,
				"fork name cannot be empty when @ is present",
				nil)
		}

		// Validate fork name
		if !forkNamePattern.MatchString(forkName) {
			return nil, errors.NewVersionError(
				ErrInvalidForkIdentifier,
				fmt.Sprintf("invalid fork name %q: must match [a-z][a-z0-9-]*", forkName),
				nil)
		}

		if forkName == "perl" {
			return nil, errors.NewVersionError(
				ErrReservedForkName,
				"'perl' cannot be used as a fork name",
				nil)
		}

		if version == "" {
			return nil, errors.NewVersionError(
				ErrInvalidForkIdentifier,
				"version cannot be empty after @",
				nil)
		}

		_, err := ParseVersion(version)
		if err != nil {
			return nil, errors.NewVersionError(
				ErrInvalidForkIdentifier,
				fmt.Sprintf("invalid version in fork identifier: %s", version),
				err)
		}

		return &ForkIdentifier{
			Remote:      remote,
			ForkName:    forkName,
			BaseVersion: version,
		}, nil
	}

	// Short form: remote/version
	_, err := ParseVersion(rest)
	if err != nil {
		return nil, errors.NewVersionError(
			ErrInvalidForkIdentifier,
			fmt.Sprintf("invalid version in fork identifier: %s", rest),
			err)
	}

	return &ForkIdentifier{
		Remote:      remote,
		ForkName:    "",
		BaseVersion: rest,
	}, nil
}

// DisplayName returns the human-readable display name for this fork identifier.
// Uses "-" as separator between fork name and version (input uses "@").
func (f *ForkIdentifier) DisplayName() string {
	if f.Remote == "origin" {
		return f.BaseVersion
	}
	if f.ForkName == "" {
		return f.Remote + "/" + f.BaseVersion
	}
	return f.Remote + "/" + f.ForkName + "-" + f.BaseVersion
}

// InstallPath returns the path segment used for installing this fork.
func (f *ForkIdentifier) InstallPath() string {
	if f.Remote == "origin" {
		return f.BaseVersion
	}
	if f.ForkName == "" {
		return f.Remote + "/" + f.BaseVersion
	}
	return f.Remote + "/" + f.ForkName + "-" + f.BaseVersion
}

// IsFork returns true if this identifier refers to a non-origin fork.
func (f *ForkIdentifier) IsFork() bool {
	return f.Remote != "origin"
}

// ForkManifest holds metadata parsed from a .pvm-fork.toml file in a fork repo
type ForkManifest struct {
	Name           string   // Fork product name (required, cannot be "perl")
	Description    string   // Human-readable description
	BaseVersion    string   // Upstream perl version this derives from (required)
	License        string   // License identifier
	ConfigureFlags []string // Additional Configure flags for building
}

// Fork manifest error codes
const (
	ErrInvalidManifest = "208" // Invalid or incomplete fork manifest
)

// Manifest filename
const forkManifestFile = ".pvm-fork.toml"

// forkManifestTOML mirrors the TOML structure for unmarshaling
type forkManifestTOML struct {
	Fork struct {
		Name        string `toml:"name"`
		Description string `toml:"description"`
		BaseVersion string `toml:"base_version"`
		License     string `toml:"license"`
	} `toml:"fork"`
	Build struct {
		ConfigureFlags []string `toml:"configure_flags"`
	} `toml:"build"`
}

// ParseForkManifest reads and validates a .pvm-fork.toml file from the given directory.
// Returns nil without error if the manifest file does not exist.
func ParseForkManifest(dir string) (*ForkManifest, error) {
	path := filepath.Join(dir, forkManifestFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read fork manifest: %w", err)
	}

	var raw forkManifestTOML
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, errors.NewVersionError(ErrInvalidManifest,
			fmt.Sprintf("parse fork manifest: %s", err), err)
	}

	if raw.Fork.Name == "" {
		return nil, errors.NewVersionError(ErrInvalidManifest,
			"fork manifest missing required field: fork.name", nil)
	}
	if raw.Fork.Name == "perl" {
		return nil, errors.NewVersionError(ErrReservedForkName,
			"'perl' cannot be used as a fork name in manifest", nil)
	}
	if raw.Fork.BaseVersion == "" {
		return nil, errors.NewVersionError(ErrInvalidManifest,
			"fork manifest missing required field: fork.base_version", nil)
	}

	return &ForkManifest{
		Name:           raw.Fork.Name,
		Description:    raw.Fork.Description,
		BaseVersion:    raw.Fork.BaseVersion,
		License:        raw.Fork.License,
		ConfigureFlags: raw.Build.ConfigureFlags,
	}, nil
}
