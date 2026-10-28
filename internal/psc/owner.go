// ABOUTME: owner, the one goroutine that holds an open LSP document (spec §6.7.1).
// ABOUTME: Debounces parses and diagnostics, cancels superseded parses, and bounds them through parseSlots.

package psc

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"tamarou.com/pvm/internal/parse"
)

// The debounce intervals of §6.7.2: a parse waits out a burst of typing,
// and diagnostics wait for a pause, since an error mid-word is noise.
const (
	parseDebounce       = 15 * time.Millisecond
	diagnosticsDebounce = 300 * time.Millisecond
)

// parseSlots bounds the parses running at once across every document
// (§6.7.4): half the CPUs, at least two, so the editor keeps the rest.
var parseSlots = make(chan struct{}, max(2, runtime.NumCPU()/2))

// clock makes the owner's timers, so a test can drive them by hand.
type clock interface {
	NewTimer(d time.Duration) timer
}

// timer is the part of time.Timer an owner uses.
type timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

type realClock struct{}

func (realClock) NewTimer(d time.Duration) timer { return realTimer{time.NewTimer(d)} }

type realTimer struct{ *time.Timer }

func (t realTimer) C() <-chan time.Time { return t.Timer.C }

// stopTimer stops t and drains a fire already sitting in its channel, so
// a Reset after it cannot be followed by that stale fire (§6.7.2).
func stopTimer(t timer) {
	if !t.Stop() {
		select {
		case <-t.C():
		default:
		}
	}
}

// parseResult is what one parse of generation gen produced: doc carries
// the text it parsed and, when err is nil, its analysis.
type parseResult struct {
	gen uint64
	doc *document
	err error
}

// owner holds one open document. Every field below ops is confined to its
// goroutine, run; other goroutines reach them only through call.
type owner struct {
	uri                   string
	parseDelay, diagDelay time.Duration
	compute               func(ctx context.Context, uri string, text []byte) (*document, error)
	publish               func(doc *document, err error) // may be nil
	// With cache set, as parse.IncrementalRepair was when the owner was
	// made, reparse parses instead of compute, from the last tree that
	// landed, sharing cache, the document's module cache, across parses.
	reparse func(ctx context.Context, uri string, text []byte, prev *document, cache *parse.ModuleCache) (*document, error)
	cache   *parse.ModuleCache

	ops     chan func()
	results chan parseResult
	quit    chan struct{}
	done    chan struct{}
	stopped sync.Once

	text    *Document
	version int
	// gen counts the texts the owner has held; started and finished are
	// the generations of the last parse begun and the last one landed.
	gen, started, finished uint64
	cancel                 context.CancelFunc
	diagDue                bool        // publish when the current generation lands
	last                   parseResult // the last parse that landed
	snap                   *document   // the last tree that landed; nil before one does
	waiters                []chan *document
	parseT, diagT          timer
}

// newOwner starts the owner of uri holding text at version, and begins
// parsing it at once: an open has no burst to coalesce, so its
// diagnostics are published as soon as the parse lands.
func newOwner(s *LSPServer, uri string, version int, text []byte) *owner {
	o := &owner{
		uri:        uri,
		parseDelay: s.parseDelay,
		diagDelay:  s.diagDelay,
		compute:    s.compute,
		publish:    s.publish,
		reparse:    s.reparse,
		ops:        make(chan func()),
		results:    make(chan parseResult),
		quit:       make(chan struct{}),
		done:       make(chan struct{}),
		text:       NewDocument(text),
		version:    version,
		gen:        1,
		diagDue:    true,
		parseT:     s.clock.NewTimer(time.Hour),
		diagT:      s.clock.NewTimer(time.Hour),
	}
	if parse.IncrementalRepair {
		o.cache = parse.NewModuleCache()
	}
	stopTimer(o.parseT)
	stopTimer(o.diagT)
	o.startParse()
	go o.run()
	return o
}

func (o *owner) run() {
	defer close(o.done)
	for {
		select {
		case f := <-o.ops:
			f()
		case <-o.parseT.C():
			if o.started != o.gen {
				o.startParse()
			}
		case <-o.diagT.C():
			o.diagDue = true
			if o.finished == o.gen {
				o.land()
			} else if o.started != o.gen {
				o.startParse()
			}
		case r := <-o.results:
			// The generation check: a parse that finished just before an
			// edit cancelled it still never lands (§6.7.3).
			if r.gen == o.gen {
				o.finished, o.last = r.gen, r
				if r.err == nil {
					o.snap = r.doc
				}
				o.land()
			}
		case <-o.quit:
			if o.cancel != nil {
				o.cancel()
			}
			return
		}
	}
}

// land publishes the last parse if diagnostics are due, then serves every
// request waiting on the current generation the last tree that landed.
func (o *owner) land() {
	if o.diagDue && o.publish != nil {
		o.publish(o.last.doc, o.last.err)
	}
	o.diagDue = false
	for _, w := range o.waiters {
		w <- o.view()
	}
	o.waiters = nil
}

// view is the last tree that landed, with a Snapshot of its text: taken
// here, on the owner's goroutine, so no edit of the text runs under it.
func (o *owner) view() *document {
	if o.snap == nil {
		return nil
	}
	v := *o.snap
	v.text = o.snap.text.Snapshot()
	return &v
}

// startParse cancels any parse in flight and parses the current text on a
// goroutine of its own, so the owner keeps taking edits while it runs.
func (o *owner) startParse() {
	stopTimer(o.parseT)
	if o.cancel != nil {
		o.cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel
	o.started = o.gen
	gen, version, text, prev := o.gen, o.version, o.text.Snapshot(), o.snap
	go func() {
		r := parseResult{gen: gen}
		r.doc, r.err = o.parse(ctx, version, text, prev)
		select {
		case o.results <- r:
		case <-o.done:
		}
	}()
}

// parse runs compute, or reparse from prev, in a parse slot. A panic in it
// is this version's error, so the owner keeps serving and the next edit
// parses afresh.
func (o *owner) parse(ctx context.Context, version int, text *Document, prev *document) (doc *document, err error) {
	doc = &document{uri: o.uri, version: version, text: text}
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("psc: internal error analysing version %d: %v", version, p)
		}
	}()
	select {
	case parseSlots <- struct{}{}:
	case <-ctx.Done():
		return doc, ctx.Err()
	}
	defer func() { <-parseSlots }()
	src := text.Text[:len(text.Text):len(text.Text)] // an append copies, never writing the snapshot
	var got *document
	if o.cache != nil {
		got, err = o.reparse(ctx, o.uri, src, prev, o.cache)
	} else {
		got, err = o.compute(ctx, o.uri, src)
	}
	if err != nil {
		return doc, err
	}
	got.uri, got.version, got.text = doc.uri, doc.version, doc.text
	return got, nil
}

// call runs f on the owner's goroutine and returns once it has, or
// returns false if the owner has stopped.
func (o *owner) call(f func()) bool {
	ran := make(chan struct{})
	select {
	case o.ops <- func() { f(); close(ran) }:
		<-ran
		return true
	case <-o.done:
		return false
	}
}

// edit applies changes to the text at version, unless version is not newer
// than the owner's, and debounces the parse and the diagnostics.
func (o *owner) edit(version int, changes []lspTextDocumentContentChangeEvent) {
	o.call(func() { o.applyChanges(version, changes) })
}

// applyEdit replaces the text with text at version, as edit does.
func (o *owner) applyEdit(version int, text []byte) {
	o.applyChanges(version, []lspTextDocumentContentChangeEvent{{Text: string(text)}})
}

// applyChanges applies each change in order, to the text the one before it
// left: a change with no range replaces the text, and one with a range has
// its UTF-16 positions clamped through the line table (§6.6) and replaces
// the bytes between them. A reversed range, which no clamp produces from an
// ordered one, is the client's error, and only that change is dropped.
func (o *owner) applyChanges(version int, changes []lspTextDocumentContentChangeEvent) {
	if version <= o.version {
		return
	}
	for _, c := range changes {
		if c.Range == nil {
			o.text = NewDocument([]byte(c.Text))
			continue
		}
		start, end := o.text.PositionToByte(c.Range.Start), o.text.PositionToByte(c.Range.End)
		_, _ = o.text.ApplyEdit(int(start), int(end), []byte(c.Text))
	}
	o.version = version
	o.gen++
	if o.cancel != nil {
		o.cancel() // supersede the parse in flight
		o.cancel = nil
	}
	o.diagDue = o.diagDelay == 0
	if o.diagDelay > 0 {
		stopTimer(o.diagT)
		o.diagT.Reset(o.diagDelay)
	}
	if o.parseDelay == 0 {
		o.startParse()
		return
	}
	stopTimer(o.parseT)
	o.parseT.Reset(o.parseDelay)
}

// current is the tree of the owner's latest text. A request jumps the
// parse debounce (§6.7.1): it forces the parse and waits for it to land.
// It returns the last tree that landed if that parse failed, and nil if
// none has or the owner has stopped.
func (o *owner) current() *document {
	reply := make(chan *document, 1)
	ok := o.call(func() {
		if o.started != o.gen {
			o.startParse()
		}
		if o.finished == o.gen {
			reply <- o.view()
			return
		}
		o.waiters = append(o.waiters, reply)
	})
	if !ok {
		return nil
	}
	select {
	case d := <-reply:
		return d
	case <-o.done:
		return nil
	}
}

// stop ends the owner and cancels its parse in flight. It returns once
// the owner's goroutine has, so nothing is published after it.
func (o *owner) stop() {
	o.stopped.Do(func() { close(o.quit) })
	<-o.done
}
