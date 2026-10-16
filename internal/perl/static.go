// ABOUTME: StaticPerl finds the perl a project names from its .perl-version or .pvm/pvm.toml and pvm's registry.
// ABOUTME: It reads files only: a static checker never runs perl and never asks MetaCPAN what "latest" is.

package perl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/platform"
	"tamarou.com/pvm/internal/xdg"
)

// StaticPerl is the perl the project at dir names: the closest
// .perl-version from dir up, or failing that the closest .pvm/pvm.toml's
// default_perl through that file's version_aliases, looked up in pvm's
// registry. `system` is the system perl imported into the registry, and a
// name containing "/" is a fork's display name there.
//
// It returns the perl's path, or "" and why the named perl was not found.
// Both are "" when the project names no perl. The aliases only MetaCPAN
// or a run of perl can answer (latest, stable, dev and the like) are not
// resolved.
func StaticPerl(dir string) (path, why string) {
	name, source, err := staticPerlName(dir)
	if err != nil {
		return "", fmt.Sprintf("%s cannot be read: %v", source, err)
	}
	if name == "" {
		return "", ""
	}
	switch strings.TrimPrefix(name, "@") {
	case "stable", "latest-stable", "dev", "latest-dev", "latest":
		return "", fmt.Sprintf("%s names %q, which only MetaCPAN can resolve", source, name)
	}
	if strings.HasPrefix(name, "@") {
		return "", fmt.Sprintf("%s names the unknown alias %q", source, name)
	}

	var version string
	switch {
	case name == "system":
		installed, _ := GetInstalledVersions()
		for _, info := range installed {
			if info.Source == "system" {
				version = info.Version
				break
			}
		}
		if version == "" {
			return "", fmt.Sprintf("%s names the system perl, which is not imported into pvm", source)
		}
	case strings.Contains(name, "/"):
		info := FindByDisplayName(name)
		if info == nil {
			return "", fmt.Sprintf("%s names %q, which is not installed", source, name)
		}
		perlPath := filepath.Join(info.InstallPath, "bin", platform.ExecutableName("perl"))
		if _, err := os.Stat(perlPath); err != nil {
			return "", fmt.Sprintf("%s names %q, whose perl is not at %s", source, name, perlPath)
		}
		return perlPath, ""
	default:
		version = name
	}

	perlPath, err := getPerlExecutablePath(version)
	if err != nil {
		return "", fmt.Sprintf("%s names %q, which is not installed", source, name)
	}
	return perlPath, ""
}

// staticPerlName is the perl name the project at dir gives, and the file
// that gives it, with pvm's precedence: any .perl-version from dir up
// before any .pvm/pvm.toml. A .perl-version names a version literally; a
// pvm.toml's default_perl goes through its version_aliases, which, as pvm
// parses it, include pvm's default aliases.
func staticPerlName(dir string) (name, source string, err error) {
	if files, _ := FindDotPerlVersionFiles(dir); len(files) > 0 {
		v, err := ReadPerlVersionFile(filepath.Dir(files[0]))
		return v, files[0], err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", "", nil
	}
	for d := abs; ; d = filepath.Dir(d) {
		path := xdg.GetProjectConfigPath(d)
		if _, err := os.Stat(path); err == nil {
			cfg, err := config.ParseFile(path)
			if err != nil {
				return "", path, err
			}
			if cfg.PVM == nil || cfg.PVM.DefaultPerl == "" {
				return "", "", nil
			}
			if v, ok := cfg.PVM.VersionAliases[strings.TrimPrefix(cfg.PVM.DefaultPerl, "@")]; ok {
				return v, path, nil
			}
			return cfg.PVM.DefaultPerl, path, nil
		}
		if filepath.Dir(d) == d {
			return "", "", nil
		}
	}
}
