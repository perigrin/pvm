// ABOUTME: ModuleCache, the module facts one LSP document keeps across its re-parses, keyed by content hash.
// ABOUTME: It saves parsing, not reading: every Loader call still runs, and an entry replays only when every module it reached reads the same bytes.

package parse

import (
	"crypto/sha256"
	"encoding/binary"
	"maps"
	"sync"
)

// ModuleCache holds what parsing each module said, for the fresh Sessions
// of one document's re-parses (01a17969-4629). Resolution is never cached:
// each parse calls the Loader for every module, so a module created,
// deleted or shadowed in an earlier root is seen on the next parse. What is
// saved is parsing the bytes the Loader returned, when they and the bytes
// of every module they transitively used are the ones an entry was parsed
// from. No mtime is read, so a save inside one mtime tick is not missed.
//
// A replayed entry gives the parse the LoadedModules, DeclarationErrors and
// DeclarationWarnings a fresh uncached parse would, in its order. A module
// on a `use` cycle is never cached, since what it says depends on which
// module of the cycle was entered first, and nor is one that uses it or
// requires a `.pl` helper. It is safe for concurrent use.
//
// shortcut: entries are never evicted, so every version of a module the
// document's parses read is held until the document closes; evict by
// last use if a long session edits modules often.
type ModuleCache struct {
	mu      sync.Mutex
	entries map[cacheKey]*cacheEntry
	parses  int // modules parsed rather than replayed
}

// NewModuleCache returns a ModuleCache holding nothing.
func NewModuleCache() *ModuleCache {
	return &ModuleCache{entries: map[cacheKey]*cacheEntry{}}
}

// NewSessionWithCache is NewSession whose parses share c, which may be nil
// for none. Its ReparseIncremental and Repair hand c to the fresh Session
// they parse in.
func NewSessionWithCache(c *ModuleCache) *Session {
	s := NewSession()
	s.cache = c
	return s
}

// cacheKey is a module's name, the hash of what it resolved to, and how the
// parse reading it reads.
type cacheKey struct {
	name                   string
	sum                    [sha256.Size]byte
	passes, inRequiredFile bool
}

// cacheEntry is what parsing one module's bytes did.
type cacheEntry struct {
	facts moduleFacts
	// events are its parse's own steps in order: each resolve it called,
	// whether or not that module was already read, and each declaration
	// error and warning it added itself.
	events []cacheEvent
	// deps is every module it reached, transitively, and what each
	// resolved to then.
	deps map[string]moduleRead
}

// cacheEvent is one step: a resolve of dep, or the error err or warning
// warn appended.
type cacheEvent struct {
	dep       string
	err, warn error
}

func (c *ModuleCache) lookup(k cacheKey) *cacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.entries[k]
}

func (c *ModuleCache) store(k cacheKey, e *cacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[k] = e
}

func (c *ModuleCache) parsed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parses++
}

// withCache has r replay modules from c, when c is set, and returns r.
func (r *resolver) withCache(c *ModuleCache) *resolver {
	if c != nil {
		r.cache, r.track = c, &cacheTrack{
			sources: map[string]*moduleSource{},
			entries: map[string]*cacheEntry{},
		}
	}
	return r
}

// moduleSource is what one name resolved to in this parse: its source, or
// its declaration, and its XS source; read is whether it was found and the
// hash of them all.
type moduleSource struct {
	src, xs        []byte
	declared, xsOK bool
	read           moduleRead
}

// cacheTrack is one parse's record for its cache: what each name resolved
// to, the entry of each module it reached through the cache, and the
// modules being read now, innermost last.
type cacheTrack struct {
	sources map[string]*moduleSource
	entries map[string]*cacheEntry
	frames  []*cacheFrame
}

// cacheFrame is the entry a module being read is building.
type cacheFrame struct {
	events            []cacheEvent
	deps              map[string]moduleRead
	errMark, warnMark int  // how many errors and warnings r held at the last step
	tainted           bool // never stored
}

// step records, as f's own, the errors and warnings r gained since f's
// last step.
func (f *cacheFrame) step(r *resolver) {
	for _, err := range r.declErrs[f.errMark:] {
		f.events = append(f.events, cacheEvent{err: err})
	}
	for _, warn := range r.declWarns[f.warnMark:] {
		f.events = append(f.events, cacheEvent{warn: warn})
	}
	f.errMark, f.warnMark = len(r.declErrs), len(r.declWarns)
}

func (t *cacheTrack) top() *cacheFrame {
	if len(t.frames) == 0 {
		return nil
	}
	return t.frames[len(t.frames)-1]
}

// taint keeps the module being read out of the cache, and so every module
// reading it.
func (t *cacheTrack) taint() {
	if f := t.top(); f != nil {
		f.tainted = true
	}
}

// source is what name resolves to, read once per parse.
func (t *cacheTrack) source(r *resolver, name string) *moduleSource {
	if m, ok := t.sources[name]; ok {
		return m
	}
	m := &moduleSource{}
	m.src, m.declared = declaration(name)
	ok := m.declared
	if !ok {
		m.src, ok = r.load(name)
	}
	if ok {
		m.xs, m.xsOK = r.load(name + ".xs")
	}
	h := sha256.New()
	for _, b := range [][]byte{m.src, m.xs} {
		h.Write(binary.LittleEndian.AppendUint64(nil, uint64(len(b))))
		h.Write(b)
	}
	h.Write([]byte{b2u(m.declared), b2u(m.xsOK)})
	m.read = moduleRead{ok: ok}
	h.Sum(m.read.sum[:0])
	t.sources[name] = m
	return m
}

func b2u(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// resolve is resolver.resolve with a cache: module is a step of the module
// being read, and what it reached joins that module's deps.
func (t *cacheTrack) resolve(r *resolver, module string) (moduleFacts, bool) {
	parent := t.top()
	if parent != nil {
		parent.step(r)
		parent.events = append(parent.events, cacheEvent{dep: module})
	}
	facts, ok := t.reach(r, module)
	if parent == nil {
		return facts, ok
	}
	// What module and its own reads added is theirs, not the parent's.
	parent.errMark, parent.warnMark = len(r.declErrs), len(r.declWarns)
	m, e := t.sources[module], t.entries[module]
	switch {
	case m != nil && !m.read.ok:
		parent.deps[module] = m.read // not found, which a parse can depend on
	case e != nil:
		parent.deps[module] = m.read
		maps.Copy(parent.deps, e.deps)
	default:
		// On a cycle, uncacheable, or read before this parse's cache: what
		// it said depends on more than its bytes.
		parent.tainted = true
	}
	return facts, ok
}

// reach is resolver.resolve's body, replaying module's entry when one
// matches and storing one when module's parse is cacheable.
func (t *cacheTrack) reach(r *resolver, module string) (moduleFacts, bool) {
	if facts, ok := r.facts[module]; ok {
		return facts, true
	}
	if r.seen[module] {
		return moduleFacts{}, false // a cycle, or not found before
	}
	r.seen[module] = true
	m := t.source(r, module)
	if !m.read.ok {
		return moduleFacts{}, false
	}
	r.loaded = append(r.loaded, module)

	key := cacheKey{module, m.read.sum, r.useIfPasses, r.inRequiredFile}
	f := &cacheFrame{deps: map[string]moduleRead{}, errMark: len(r.declErrs), warnMark: len(r.declWarns)}
	t.frames = append(t.frames, f)
	e := r.cache.lookup(key)
	if e != nil && !t.matches(r, e) {
		e = nil
	}
	var facts moduleFacts
	if e != nil {
		// As the resolver ran: each resolve again, which skips a module
		// this parse already reached, and this module's own errors.
		for _, ev := range e.events {
			switch {
			case ev.dep != "":
				r.resolve(ev.dep)
			case ev.err != nil:
				r.declErrs = append(r.declErrs, ev.err)
			default:
				r.declWarns = append(r.declWarns, ev.warn)
			}
		}
		facts = e.facts
	} else {
		r.cache.parsed()
		facts = r.read(module, m.src, m.declared, func() ([]byte, bool) { return m.xs, m.xsOK })
		f.step(r)
		if !f.tainted {
			e = &cacheEntry{facts: facts, events: f.events, deps: f.deps}
			r.cache.store(key, e)
		}
	}
	t.frames = t.frames[:len(t.frames)-1]
	if e != nil {
		t.entries[module] = e
	}
	r.keep(module, facts)
	return facts, true
}

// matches reports that every module e reached resolves, in this parse, to
// what it did when e was stored.
func (t *cacheTrack) matches(r *resolver, e *cacheEntry) bool {
	for name, want := range e.deps {
		if t.source(r, name).read != want {
			return false
		}
	}
	return true
}
