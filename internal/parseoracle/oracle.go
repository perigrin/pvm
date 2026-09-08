// ABOUTME: Asks perl how it parsed a file and returns the answer as Go values.
// ABOUTME: Ground truth for parser fidelity, so a disagreement is a fact rather than an opinion.

package parseoracle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Facts is what perl reports about its own parse of a file. The optree is the
// parse, resolved: a prototype that turned a list into a reference shows up as
// an srefgen that is simply absent without the prototype.
//
// Ops is captured after the peephole optimiser, so compare it for the presence
// of a marker op rather than for an exact sequence or a count — constant
// folding removes ops the parser really did build.
type Facts struct {
	// OK reports whether the file compiled.
	OK bool
	// Ops is the op sequence in execution order, which diffs cleanly.
	Ops []string
	// OpCount is len(Ops), as perl counted them.
	OpCount int
	// Srefgen counts the reference-taking ops a \-prototype introduces.
	Srefgen int
	// Entersub counts subroutine calls.
	Entersub int
	// Prototypes maps every sub name in scope to its prototype.
	Prototypes map[string]string
	// Stderr is perl's diagnostic when OK is false, empty otherwise. It is
	// the only thing that separates a parser being wrong from a machine
	// missing Config.pm; an exit status cannot tell those apart.
	Stderr string
}

// Options controls how perl is invoked.
type Options struct {
	// Dir is the working directory to compile from, and it is mandatory for
	// most of perl's own test suite: t/test.pl clears @INC and unshifts
	// ../lib, so those files compile only from inside a populated t/.
	// PERL5LIB cannot substitute, because @INC is cleared after it is read.
	Dir string
	// Shebang passes -T through for a file whose `#!` line asks for taint
	// mode. Perl refuses to compile such a file unless -T is also on the
	// command line, which otherwise looks like a syntax error and is not.
	Shebang bool
	// Timeout bounds a single invocation. Zero means DefaultTimeout.
	Timeout time.Duration
}

// DefaultTimeout bounds one oracle invocation. Compiling a corpus file costs
// tens of milliseconds; anything approaching this is wedged.
const DefaultTimeout = 60 * time.Second

// killDelay is how long a cancelled child has to exit on its own before it is
// killed. The script shells out to an inner `sh -c perl` that inherits the
// stdout pipe, so cancelling the context alone does not unwedge a read.
const killDelay = 2 * time.Second

// Ask compiles src and returns perl's parse facts for it.
func Ask(ctx context.Context, src []byte, opts Options) (Facts, error) {
	dir, err := os.MkdirTemp("", "parseoracle")
	if err != nil {
		return Facts{}, fmt.Errorf("oracle scratch dir: %w", err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "probe.pl")
	if err := os.WriteFile(file, src, 0o644); err != nil {
		return Facts{}, fmt.Errorf("write probe: %w", err)
	}
	return AskFile(ctx, file, opts)
}

// AskFile compiles the file at path and returns perl's parse facts for it. A
// relative path is resolved against opts.Dir, matching how perl's own tests
// refer to each other.
func AskFile(ctx context.Context, path string, opts Options) (Facts, error) {
	script, err := scriptPath()
	if err != nil {
		return Facts{}, err
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "perl", script, path)
	cmd.Dir = opts.Dir
	// WaitDelay is what actually returns control. The script's inner
	// `sh -c perl` child holds the stdout pipe open, so without it a
	// cancelled context leaves Wait blocked on a read that never ends.
	cmd.WaitDelay = killDelay

	env := os.Environ()
	if opts.Dir != "" {
		env = append(env, "ORACLE_CHDIR="+opts.Dir)
	}
	if opts.Shebang {
		env = append(env, "ORACLE_TAINT=1")
	}
	cmd.Env = env

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Facts{}, fmt.Errorf("oracle on %s: %w", path, ctxErr)
		}
		return Facts{}, fmt.Errorf("oracle on %s: %w (%s)", path, err,
			strings.TrimSpace(stderr.String()))
	}

	return decode(out)
}

// wire mirrors parse_facts.pl's JSON. It is separate from Facts because the
// script speaks in perl's idiom — ok is 0 or 1 — and Facts speaks in Go's.
type wire struct {
	OK         int               `json:"ok"`
	Ops        []string          `json:"ops"`
	OpCount    int               `json:"op_count"`
	Srefgen    int               `json:"srefgen"`
	Entersub   int               `json:"entersub"`
	Prototypes map[string]string `json:"prototypes"`
	Stderr     string            `json:"stderr"`
}

func decode(out []byte) (Facts, error) {
	var w wire
	if err := json.Unmarshal(out, &w); err != nil {
		return Facts{}, fmt.Errorf("decode oracle output %q: %w", out, err)
	}
	return Facts{
		OK:         w.OK == 1,
		Ops:        w.Ops,
		OpCount:    w.OpCount,
		Srefgen:    w.Srefgen,
		Entersub:   w.Entersub,
		Prototypes: w.Prototypes,
		Stderr:     w.Stderr,
	}, nil
}

// scriptPath locates parse_facts.pl relative to this source file, so the
// oracle is callable from any working directory. Resolving it against the
// process working directory only ever worked for a test run from inside this
// package, and the corpus runner is not that.
func scriptPath() (string, error) {
	_, self, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("cannot locate the parseoracle package source")
	}
	script := filepath.Join(filepath.Dir(self), "testdata", "parse_facts.pl")
	if _, err := os.Stat(script); err != nil {
		return "", fmt.Errorf("parse_facts.pl not found at %s: %w", script, err)
	}
	return script, nil
}
