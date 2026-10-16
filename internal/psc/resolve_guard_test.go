//go:build !windows

// ABOUTME: Tests that psc check and the LSP find a project's perl without running a process or opening a socket.
// ABOUTME: Each case runs in a child test process whose children and proxied connections the parent counts.

package psc_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/psc"
)

// guardMode, guardTarget name what the child test process runs: psc check
// on the target, or the LSP opening it.
const (
	guardMode   = "PSC_GUARD_MODE"
	guardTarget = "PSC_GUARD_TARGET"
)

// TestPscGuardChild is the child process: it resolves the target's @INC
// the way guardMode says, then prints the peak memory of any process it
// ran and waited for. A process that never ran a child prints 0.
func TestPscGuardChild(t *testing.T) {
	mode, target := os.Getenv(guardMode), os.Getenv(guardTarget)
	if mode == "" {
		t.Skip("run only as the child of a guard test")
	}
	switch mode {
	case "check":
		cmd := psc.NewCommand()
		cmd.SetArgs([]string{"check", target})
		var out strings.Builder
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		_ = cmd.Execute()
	case "lsp":
		src, err := os.ReadFile(target)
		require.NoError(t, err)
		require.NoError(t, psc.NewLSPServer().OpenDocument("file://"+filepath.ToSlash(target), src))
	}
	var ru syscall.Rusage
	require.NoError(t, syscall.Getrusage(syscall.RUSAGE_CHILDREN, &ru))
	fmt.Printf("guard-children-maxrss=%d\n", ru.Maxrss)
}

// guardRun runs the child for mode on target with an empty HOME, a PATH
// whose first perl records being run, and every HTTP connection sent to a
// proxy that counts them. It returns whether a child process ran, whether
// the fake perl ran, and how many connections the proxy saw.
func guardRun(t *testing.T, mode, target string) (ranChild, ranPerl bool, conns int64) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()
	var count atomic.Int64
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			count.Add(1)
			c.Close()
		}
	}()

	bin := t.TempDir()
	marker := filepath.Join(t.TempDir(), "perl-ran")
	require.NoError(t, os.WriteFile(filepath.Join(bin, "perl"),
		[]byte("#!/bin/sh\ntouch '"+marker+"'\necho 'This is perl 5, version 42, subversion 0 (v5.42.0)'\n"), 0o755))

	home := t.TempDir()
	proxy := "http://" + ln.Addr().String()
	cmd := exec.Command(os.Args[0], "-test.run=^TestPscGuardChild$", "-test.count=1", "-test.v")
	cmd.Env = []string{
		"HOME=" + home, "XDG_CONFIG_HOME=" + home, "XDG_DATA_HOME=" + home,
		"XDG_CACHE_HOME=" + home, "XDG_STATE_HOME=" + home,
		"PATH=" + bin + string(os.PathListSeparator) + "/usr/bin:/bin",
		"HTTP_PROXY=" + proxy, "HTTPS_PROXY=" + proxy, "http_proxy=" + proxy, "https_proxy=" + proxy,
		"PVM_SKIP_NETWORK_CALLS=1",
		guardMode + "=" + mode, guardTarget + "=" + target,
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s", out)
	require.Contains(t, string(out), "guard-children-maxrss=", "%s", out)

	_, statErr := os.Stat(marker)
	return !strings.Contains(string(out), "guard-children-maxrss=0\n"), statErr == nil, count.Load()
}

// guardProject writes a.pl into a fresh directory whose .perl-version or
// .pvm/pvm.toml names version, and returns a.pl's path.
func guardProject(t *testing.T, how, version string) string {
	t.Helper()
	dir := t.TempDir()
	switch how {
	case ".perl-version":
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".perl-version"), []byte(version+"\n"), 0o644))
	case "pvm.toml":
		require.NoError(t, os.MkdirAll(filepath.Join(dir, ".pvm"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".pvm", "pvm.toml"),
			[]byte("[pvm]\ndefault_perl = \""+version+"\"\n"), 0o644))
	}
	file := filepath.Join(dir, "a.pl")
	require.NoError(t, os.WriteFile(file, []byte("use strict;\nmy $x = 1;\n"), 0o644))
	return file
}

// TestPscNeverExecutesPerl checks that a project naming `system` as its
// perl, in its .perl-version or its pvm.toml, makes neither psc check nor
// the LSP run a process: not perl from PATH, and not /usr/bin/perl -v.
func TestPscNeverExecutesPerl(t *testing.T) {
	for _, how := range []string{".perl-version", "pvm.toml"} {
		for _, mode := range []string{"check", "lsp"} {
			t.Run(how+"/"+mode, func(t *testing.T) {
				ranChild, ranPerl, _ := guardRun(t, mode, guardProject(t, how, "system"))
				require.False(t, ranPerl, "the perl on PATH was run")
				require.False(t, ranChild, "a process was run")
			})
		}
	}
}

// TestPscNeverFetchesTheNetwork checks that a project naming a perl only
// MetaCPAN can resolve (dev, latest-stable, @latest) makes neither psc
// check nor the LSP open a connection, with PVM_SKIP_NETWORK_CALLS set.
func TestPscNeverFetchesTheNetwork(t *testing.T) {
	for _, version := range []string{"dev", "latest-stable", "@latest"} {
		for _, how := range []string{".perl-version", "pvm.toml"} {
			for _, mode := range []string{"check", "lsp"} {
				t.Run(version+"/"+how+"/"+mode, func(t *testing.T) {
					_, _, conns := guardRun(t, mode, guardProject(t, how, version))
					require.Zero(t, conns, "a connection was opened")
				})
			}
		}
	}
}
