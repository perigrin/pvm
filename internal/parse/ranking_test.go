// ABOUTME: Pins which 50 core modules get declarations: the ranking of core modules by imports over perl's privlib.
// ABOUTME: The rule is tested on synthetic trees everywhere and against perl 5.42.0's privlib where it is installed.

package parse_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/perltest"
)

// coreModuleRank is one module's place in the ranking.
type coreModuleRank struct {
	Name  string
	Count int
}

// coreModuleRanking is the ranking and its totals: Modules is how many core
// modules are imported at all, Imports how many imports were counted.
type coreModuleRanking struct {
	Ranked  []coreModuleRank
	Modules int
	Imports int
}

// rankCoreModules ranks the core modules under privlib by import count.
// The rule (perigrin, 2026-10-16):
//
//   - A core module is any .pm under privlib, named with the arch
//     directory stripped (archDir: archlib relative to privlib, "" for none).
//   - An import is a line matching ^\s*(use|require)\s+[A-Z][A-Za-z0-9_:]+
//     in any .pm or .pl under privlib.
//   - An import is not counted when the importing file's top-level path
//     component, with the arch dir stripped and .pm/.pl removed, equals the
//     module's first :: component.
//   - Only imports of core modules count. Ties are broken by module name,
//     ascending.
func rankCoreModules(t *testing.T, privlib, archDir string) coreModuleRanking {
	t.Helper()
	core := map[string]bool{}
	sources := map[string]string{} // path to its slash-separated name relative to privlib, arch dir stripped
	err := filepath.WalkDir(privlib, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		if ext != ".pm" && ext != ".pl" {
			return nil
		}
		rel, err := filepath.Rel(privlib, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if archDir != "" {
			rel = strings.TrimPrefix(rel, filepath.ToSlash(archDir)+"/")
		}
		if ext == ".pm" {
			core[strings.ReplaceAll(strings.TrimSuffix(rel, ".pm"), "/", "::")] = true
		}
		sources[path] = rel
		return nil
	})
	require.NoError(t, err)

	counts := map[string]int{}
	var ranking coreModuleRanking
	for path, rel := range sources {
		top, _, _ := strings.Cut(strings.TrimSuffix(strings.TrimSuffix(rel, ".pm"), ".pl"), "/")
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		for _, line := range strings.Split(string(content), "\n") {
			m := importLine.FindStringSubmatch(line)
			if m == nil || !core[m[1]] {
				continue
			}
			if first, _, _ := strings.Cut(m[1], "::"); first == top {
				continue
			}
			counts[m[1]]++
			ranking.Imports++
		}
	}
	for name, count := range counts {
		ranking.Ranked = append(ranking.Ranked, coreModuleRank{name, count})
	}
	sort.Slice(ranking.Ranked, func(i, j int) bool {
		a, b := ranking.Ranked[i], ranking.Ranked[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		return a.Name < b.Name
	})
	ranking.Modules = len(ranking.Ranked)
	return ranking
}

// importLine is the rule's import: a use or require of a capitalised name.
var importLine = regexp.MustCompile(`^\s*(?:use|require)\s+([A-Z][A-Za-z0-9_:]+)`)

// writeTree writes files (slash-separated relative path to content) under a
// fresh temporary directory and returns it.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return root
}

func TestCoreModuleRankingRule(t *testing.T) {
	root := writeTree(t, map[string]string{
		"Carp.pm":           "package Carp;\nuse Beta;\nuse Alpha;\n",
		"Alpha.pm":          "package Alpha;\n",
		"Beta.pm":           "package Beta;\n",
		"Foo.pm":            "package Foo;\n",
		"Foo/Bar.pm":        "package Foo::Bar;\nuse strict;\nuse Foo;\nuse Carp;\n  use POSIX qw(floor);\n",
		"script.pl":         "require Carp;\nuse POSIX ();\n",
		"arch/POSIX.pm":     "package POSIX;\nuse Carp;\nuse POSIX::Sub;\n",
		"arch/POSIX/Sub.pm": "package POSIX::Sub;\n",
	})
	got := rankCoreModules(t, root, "arch")
	assert.Equal(t, []coreModuleRank{
		{"Carp", 3}, {"POSIX", 2}, {"Alpha", 1}, {"Beta", 1},
	}, got.Ranked, "own namespace excluded, arch dir stripped, .pl counted, pragma ignored, tie ascending")
	assert.Equal(t, 4, got.Modules)
	assert.Equal(t, 7, got.Imports)
}

func TestCoreModuleRankingIgnoresNonCoreImports(t *testing.T) {
	root := writeTree(t, map[string]string{
		"Carp.pm": "package Carp;\n",
		"Foo.pm":  "package Foo;\nuse Carp;\nuse Not::Core;\nrequire Missing;\n",
	})
	got := rankCoreModules(t, root, "")
	assert.Equal(t, []coreModuleRank{{"Carp", 1}}, got.Ranked)
	assert.Equal(t, 1, got.Modules)
	assert.Equal(t, 1, got.Imports)
}

func TestCoreModuleRankingIgnoresVersionAndDynamicRequire(t *testing.T) {
	root := writeTree(t, map[string]string{
		"Carp.pm": "package Carp;\n",
		"Foo.pm":  "package Foo;\nuse 5.010;\nuse v5.36;\nrequire $file;\n",
	})
	got := rankCoreModules(t, root, "")
	assert.Empty(t, got.Ranked)
	assert.Zero(t, got.Modules)
	assert.Zero(t, got.Imports)
}

func TestCoreModuleRankingNamespaceIsWholeComponent(t *testing.T) {
	root := writeTree(t, map[string]string{
		"IO/Handle.pm": "package IO::Handle;\nuse IO::File;\n",
		"IO/File.pm":   "package IO::File;\n",
		"IPC/Open3.pm": "package IPC::Open3;\nuse IO::Handle;\n",
		"Test/More.pm": "package Test::More;\n",
		"Test2/API.pm": "package Test2::API;\nuse Test::More;\n",
	})
	got := rankCoreModules(t, root, "")
	assert.Equal(t, []coreModuleRank{{"IO::Handle", 1}, {"Test::More", 1}}, got.Ranked)
	assert.Equal(t, 2, got.Modules)
	assert.Equal(t, 2, got.Imports)
}

func TestCoreModuleRankingEmptyTree(t *testing.T) {
	got := rankCoreModules(t, t.TempDir(), "")
	assert.Empty(t, got.Ranked)
	assert.Zero(t, got.Modules)
	assert.Zero(t, got.Imports)
}

func TestCoreModuleRanking(t *testing.T) {
	p := perltest.Perl542(t)
	archDir, err := filepath.Rel(p.Privlib, p.Archlib)
	require.NoError(t, err)
	got := rankCoreModules(t, p.Privlib, archDir)

	assert.Equal(t, 128, got.Modules, "core modules imported from outside their namespace")
	assert.Equal(t, 1427, got.Imports, "imports of core modules from outside their namespace")
	require.GreaterOrEqual(t, len(got.Ranked), 50)

	want, err := os.ReadFile(filepath.Join("testdata", "core_module_ranking.txt"))
	require.NoError(t, err)
	var top strings.Builder
	for _, r := range got.Ranked[:50] {
		fmt.Fprintf(&top, "%s %d\n", r.Name, r.Count)
	}
	assert.Equal(t, string(want), top.String(), "the top 50 core modules by import count")
}
