// ABOUTME: Caches perl's parse facts on a content hash so a corpus sweep can gate a commit.
// ABOUTME: B::Concise costs ~600ms/file and its output is byte-identical across runs, so it is cacheable.

package parseoracle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Cache stores what perl said about a file, keyed on everything that can
// change the answer.
//
// A full sweep costs ~6 minutes: B::Concise is ~600 ms on a real corpus file,
// which is what stops the ratchet from running on every commit. Concise output
// is byte-identical across runs (verified by md5 on repeated runs of the same
// input), so the answer for a given question is a constant and caching it is
// correct rather than a source of stale results.
//
// Only Facts are cached, never verdicts. Facts depend on perl, which does not
// change between commits; verdicts depend on our parser, which changes
// constantly, and a cached verdict would hide exactly the movement the ratchet
// exists to detect.
//
// There is no eviction, size cap, or TTL. The cache is bounded by the corpus:
// one entry per (file content, identity), and a fixed corpus produces a fixed
// number of entries.
type Cache struct {
	dir      string
	identity CacheIdentity

	// mu serialises writes. Reads need no lock — an entry is written once
	// via a rename, so a reader sees either the whole file or nothing.
	mu sync.Mutex
}

// CacheIdentity is everything other than the source bytes that can change
// perl's answer. All three components are load-bearing:
//
//   - Interpreter is perl's $]. A different perl parses differently; that is
//     the whole reason corpus.pin records it.
//   - Revision is the corpus checkout. It pins which files the answers are
//     about, so a re-baselined corpus does not inherit the old one's cache.
//   - Script is a hash of parse_facts.pl. The script is the measuring
//     instrument: changing what it measures invalidates every prior answer,
//     and a key that omits it silently returns facts the current script would
//     never have produced.
type CacheIdentity struct {
	Interpreter string
	Revision    string
	Script      string
}

// DefaultCacheDir is where a locally generated cache lives, relative to the
// repository root.
//
// It is gitignored rather than committed. The conformance plan assumed the
// opposite, on the reasoning that CI would then spawn no perl — but the whole
// oracle phase costs ~35 seconds cold across 620 files, which does not justify
// carrying 620 files of perl output in the repo forever and re-churning them
// on every re-baseline of the interpreter or the corpus pin.
const DefaultCacheDir = "internal/parseoracle/testdata/oracle_cache"

// OpenCache opens a cache at dir, deriving the identity from the interpreter
// on PATH, the pin's corpus revision, and the current parse_facts.pl.
func OpenCache(dir string) (*Cache, error) {
	id, err := currentIdentity()
	if err != nil {
		return nil, err
	}
	return OpenCacheFor(dir, id)
}

// OpenCacheFor opens a cache at dir under an explicit identity. Tests use it
// to prove a key component matters without needing a second perl installed.
func OpenCacheFor(dir string, id CacheIdentity) (*Cache, error) {
	if dir == "" {
		return nil, fmt.Errorf("oracle cache needs a directory")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating oracle cache at %s: %w", dir, err)
	}
	return &Cache{dir: dir, identity: id}, nil
}

// Ask returns perl's facts for src, from the cache when it has them.
func (c *Cache) Ask(ctx context.Context, src []byte, opts Options) (Facts, error) {
	return c.lookup(ctx, src, "-", opts, func() (Facts, error) {
		return Ask(ctx, src, opts)
	})
}

// AskFile returns perl's facts for the file at path, from the cache when it
// has them. The key is the file's CONTENT, not its name: an edited file is a
// different question, and two identical files are the same one.
func (c *Cache) AskFile(ctx context.Context, path string, opts Options) (Facts, error) {
	src, err := os.ReadFile(resolve(path, opts.Dir))
	if err != nil {
		return Facts{}, fmt.Errorf("reading %s for the oracle cache: %w", path, err)
	}
	return c.lookup(ctx, src, path, opts, func() (Facts, error) {
		return AskFile(ctx, path, opts)
	})
}

// askFile is AskFile with a nil receiver meaning "no cache, ask perl". It
// exists so the runner has one call site rather than a branch, and it is
// unexported because a nil *Cache is a runner detail, not an API.
func (c *Cache) askFile(ctx context.Context, path string, opts Options) (Facts, error) {
	if c == nil {
		return AskFile(ctx, path, opts)
	}
	return c.AskFile(ctx, path, opts)
}

// entry is one cached answer. It is stored as indented JSON with the source
// path alongside the facts, because the cache is committed and a reviewer
// reading the diff needs to see WHICH file's answer moved, not just that some
// hash's did.
type entry struct {
	// Source names the file this answers for. Informational: the key is the
	// content hash, so a renamed file with identical bytes reuses this entry
	// and the recorded name is simply whichever asked first.
	Source string `json:"source"`
	// Interpreter, Revision and Script restate the identity so an entry is
	// self-describing when read on its own.
	Interpreter string `json:"interpreter"`
	Revision    string `json:"revision"`
	Script      string `json:"script"`
	Facts       Facts  `json:"facts"`
}

func (c *Cache) lookup(ctx context.Context, src []byte, source string, opts Options, ask func() (Facts, error)) (Facts, error) {
	path := filepath.Join(c.dir, c.key(src, opts)+".json")

	if facts, ok := readEntry(path); ok {
		return facts, nil
	}

	facts, err := ask()
	if err != nil {
		return facts, err
	}
	c.store(path, entry{
		Source:      source,
		Interpreter: c.identity.Interpreter,
		Revision:    c.identity.Revision,
		Script:      c.identity.Script,
		Facts:       facts,
	})
	return facts, nil
}

// readEntry returns a cached answer, or false when there is not a usable one.
//
// A corrupt entry is a miss, not an error. The cache is a throwaway artifact
// and hand-editing or truncating one must never be able to break a build; the
// worst it can cost is one perl invocation.
func readEntry(path string) (Facts, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Facts{}, false
	}
	var e entry
	if err := json.Unmarshal(data, &e); err != nil {
		return Facts{}, false
	}
	return e.Facts, true
}

// store writes an entry via a temp file and a rename, so a concurrent reader
// sees either the previous entry or the complete new one. The sweep runs 24
// workers and several may miss on the same content.
func (c *Cache) store(path string, e entry) {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return // An unencodable fact is not worth failing a run over.
	}
	data = append(data, '\n')

	c.mu.Lock()
	defer c.mu.Unlock()

	tmp, err := os.CreateTemp(c.dir, ".entry-*")
	if err != nil {
		return
	}
	name := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(name)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return
	}
	// CreateTemp makes 0o600, and a committed cache is meant to be readable
	// by whoever checks the tree out.
	if err := os.Chmod(name, 0o644); err != nil {
		os.Remove(name)
		return
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
	}
}

// key hashes everything that can change perl's answer. Each component is
// length-prefixed by a NUL separator so no two different questions can
// concatenate to the same bytes.
func (c *Cache) key(src []byte, opts Options) string {
	h := sha256.New()
	for _, part := range []string{
		c.identity.Interpreter,
		c.identity.Revision,
		c.identity.Script,
		// Dir changes what @INC finds, so the same bytes compiled from a
		// different directory is a different question.
		opts.Dir,
		// Taint mode changes what perl accepts outright.
		fmt.Sprintf("taint=%t", opts.Shebang),
	} {
		h.Write([]byte(part))
		h.Write([]byte{0})
	}
	h.Write(src)
	return hex.EncodeToString(h.Sum(nil))
}

// currentIdentity reads the three identity components from the environment
// this process is actually measuring in.
func currentIdentity() (CacheIdentity, error) {
	id := CacheIdentity{Revision: "unpinned"}

	out, err := exec.Command("perl", "-e", "print $]").Output()
	if err != nil {
		return id, fmt.Errorf("asking perl for its version: %w", err)
	}
	id.Interpreter = strings.TrimSpace(string(out))

	script, err := scriptPath()
	if err != nil {
		return id, err
	}
	data, err := os.ReadFile(script)
	if err != nil {
		return id, fmt.Errorf("hashing %s: %w", script, err)
	}
	sum := sha256.Sum256(data)
	id.Script = hex.EncodeToString(sum[:])

	// The pin's revision joins the key when there is a pin to read. Its
	// absence is not fatal: Ask on a scratch buffer has no corpus.
	if pin, err := ReadPin(filepath.Join(filepath.Dir(script), "corpus.pin")); err == nil {
		id.Revision = pin.Revision
	}
	return id, nil
}
