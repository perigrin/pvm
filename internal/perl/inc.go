// ABOUTME: INC replicates the @INC a perl builds, from PERL5LIB and its Config.pm read as text, never running perl.
// ABOUTME: A static checker reads untrusted code, and running perl runs whatever PERL5OPT or sitecustomize.pl says.

package perl

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// incConfigKeys are the Config.pm entries perl pushes onto @INC after
// PERL5LIB, in the order perl.c pushes them.
var incConfigKeys = []string{
	"sitearchexp", "sitelibexp", "vendorarchexp", "vendorlibexp", "archlibexp", "privlibexp",
}

// configEntry matches one `key => 'value',` line of Config.pm's literal
// hash, the value optionally wrapped in a relocatable perl's relocate_inc().
// configpm writes each \ in a value as \\ and each ' as \'.
var configEntry = regexp.MustCompile(`^\s+(\w+) => (?:relocate_inc\()?'((?:[^'\\]|\\.)*)'`)

// configUnescape reads a single-quoted Perl string's body as perl does:
// \\ is \ and \' is ', and any other backslash stands for itself.
var configUnescape = strings.NewReplacer(`\\`, `\`, `\'`, `'`)

// INC is the @INC the perl at perlPath builds under the given PERL5LIB,
// without running it: each PERL5LIB entry, preceded by its version/arch,
// version and arch subdirectories that exist, then the perl's own site,
// vendor and core library directories as its Config.pm names them.
//
// perlPath is a pvm-style install's bin/perl, whose Config.pm sits in
// lib/perl5/<version>/<archname>/. With no perlPath, or no Config.pm
// there, @INC is PERL5LIB alone.
//
// ponytail: inc_version_list and otherlibdirs are not read; every perl pvm
// installs leaves them empty. `use lib` is the parser's to apply.
func INC(perlPath, perl5lib string) []string {
	config := readConfig(perlPath)
	var inc []string
	for _, dir := range filepath.SplitList(perl5lib) {
		if dir == "" {
			continue
		}
		// perl appends each subdirectory with a slash, and does not clean
		// the result.
		version, arch := config["version"], config["archname"]
		for _, sub := range []string{version + "/" + arch, version, arch} {
			if version == "" || arch == "" {
				break
			}
			if fi, err := os.Stat(dir + "/" + sub); err == nil && fi.IsDir() {
				inc = append(inc, dir+"/"+sub)
			}
		}
		inc = append(inc, dir)
	}
	for _, key := range incConfigKeys {
		if dir := config[key]; dir != "" {
			inc = append(inc, dir)
		}
	}
	return inc
}

// readConfig reads the Config.pm of the pvm-style install perlPath is the
// perl of, as text, and returns its literal entries. A relocatable perl's
// `.../` is the directory perlPath is in. It is empty when there is no
// Config.pm to read.
func readConfig(perlPath string) map[string]string {
	config := map[string]string{}
	if perlPath == "" {
		return config
	}
	binDir := filepath.Dir(perlPath)
	// Config_heavy.pl sits only beside the Config.pm that is perl's, not
	// beside Net/Config.pm.
	heavy, _ := filepath.Glob(filepath.Join(filepath.Dir(binDir), "lib", "perl5", "*", "*", "Config_heavy.pl"))
	if len(heavy) == 0 {
		return config
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(heavy[0]), "Config.pm"))
	if err != nil {
		return config
	}
	for _, line := range strings.Split(string(data), "\n") {
		m := configEntry.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		value := configUnescape.Replace(m[2])
		if rest, ok := strings.CutPrefix(value, ".../"); ok {
			value = filepath.Join(binDir, rest)
		}
		config[m[1]] = value
	}
	return config
}
