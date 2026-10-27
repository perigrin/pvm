// ABOUTME: Tests for the per-document owner goroutine: ordering, debounce, cancellation, parseSlots.
// ABOUTME: Timers run on a fake clock and concurrency is proved by barriers, so no wall time decides a test.

package psc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
)

// fakeClock fires its timers only when a test advances it. A fired timer's
// value stays in its channel after a Stop, as a time.Timer's did before Go
// 1.23, which is the case the Stop/drain/Reset idiom guards against.
type fakeClock struct {
	mu     sync.Mutex
	now    time.Duration
	timers []*fakeTimer
}

type fakeTimer struct {
	clk    *fakeClock
	c      chan time.Time
	at     time.Duration
	active bool
}

func (c *fakeClock) NewTimer(d time.Duration) timer {
	c.mu.Lock()
	defer c.mu.Unlock()
	t := &fakeTimer{clk: c, c: make(chan time.Time, 1), at: c.now + d, active: true}
	c.timers = append(c.timers, t)
	return t
}

func (t *fakeTimer) C() <-chan time.Time { return t.c }

func (t *fakeTimer) Stop() bool {
	t.clk.mu.Lock()
	defer t.clk.mu.Unlock()
	was := t.active
	t.active = false
	return was
}

func (t *fakeTimer) Reset(d time.Duration) bool {
	t.clk.mu.Lock()
	defer t.clk.mu.Unlock()
	was := t.active
	t.at, t.active = t.clk.now+d, true
	return was
}

// Advance moves the clock on by d and fires every timer now due.
func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now += d
	for _, t := range c.timers {
		if t.active && t.at <= c.now {
			t.active = false
			select {
			case t.c <- time.Time{}:
			default:
			}
		}
	}
}

// publication is one publish an owner made.
type publication struct {
	version int
	err     error
}

// rig is an LSPServer on a fake clock whose parses are compute, and which
// records every publish.
type rig struct {
	s   *LSPServer
	clk *fakeClock
	pub chan publication

	mu     sync.Mutex
	parsed []string // the text of each compute call, in call order
}

// newRig makes a rig; a nil compute succeeds on any text.
func newRig(t *testing.T, compute func(ctx context.Context, text string) error) *rig {
	t.Helper()
	r := &rig{s: NewLSPServer(), clk: &fakeClock{}, pub: make(chan publication, 1000)}
	r.s.clock = r.clk
	r.s.compute = func(ctx context.Context, uri string, text []byte) (*document, error) {
		r.mu.Lock()
		r.parsed = append(r.parsed, string(text))
		r.mu.Unlock()
		if compute != nil {
			if err := compute(ctx, string(text)); err != nil {
				return nil, err
			}
		}
		return &document{analysis: &infer.FileAnalysis{}}, nil
	}
	r.s.publish = func(doc *document, err error) { r.pub <- publication{doc.version, err} }
	t.Cleanup(r.s.Shutdown)
	return r
}

// next is the next publish, failing the test if none comes. The deadline
// only bounds a hang; it decides nothing.
func (r *rig) next(t *testing.T) publication {
	t.Helper()
	select {
	case p := <-r.pub:
		return p
	case <-time.After(10 * time.Second):
		t.Fatal("no publish")
		return publication{}
	}
}

// quiet syncs with uri's owner a few times and asserts nothing more was
// published.
func (r *rig) quiet(t *testing.T, uri string) {
	t.Helper()
	o := r.s.owner(uri)
	for range 5 {
		o.call(func() {})
	}
	assert.Empty(t, r.pub, "nothing more is published")
}

func (r *rig) computed() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.parsed...)
}

// await waits for ch to close, failing the test if it does not.
func await(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatal(what)
	}
}

const ownedURI = "file:///owned.pl"

func TestDocumentOwnerOnePerURI(t *testing.T) {
	r := newRig(t, nil)
	r.s.parseDelay, r.s.diagDelay = 0, 0

	r.s.open(ownedURI, 1, []byte("v1"))
	first := r.s.owner(ownedURI)
	r.s.open(ownedURI, 1, []byte("v1")) // a reopen replaces the owner
	await(t, first.done, "the replaced owner stops")
	assert.Len(t, r.s.documents, 1, "one owner for the uri")

	for _, v := range []string{"v2", "v3", "v4", "v5"} {
		r.s.change(ownedURI, int(v[1]-'0'), []byte(v))
	}
	doc := r.s.snapshot(ownedURI)
	require.NotNil(t, doc)
	assert.Equal(t, 5, doc.version)
	assert.Equal(t, "v5", string(doc.text.Text))
	assert.Len(t, r.s.documents, 1, "edits make no second owner")

	// The snapshot landed after v5's publish. A superseded parse never
	// lands, so what was published is in version order and ends at v5.
	var published []int
	for len(r.pub) > 0 {
		published = append(published, (<-r.pub).version)
	}
	assert.IsNonDecreasing(t, published, "edits apply in order")
	assert.Equal(t, 5, published[len(published)-1], "%v", published)
}

func TestDiagnosticsDebounced(t *testing.T) {
	r := newRig(t, nil)
	r.s.open(ownedURI, 1, []byte("v1"))
	assert.Equal(t, 1, r.next(t).version, "an open publishes at once")

	for v := 2; v <= 5; v++ {
		r.s.change(ownedURI, v, []byte{'v', byte('0' + v)})
		r.clk.Advance(50 * time.Millisecond) // parses, but inside the diagnostics interval
	}
	r.clk.Advance(diagnosticsDebounce - 50*time.Millisecond - time.Millisecond)
	r.quiet(t, ownedURI)

	r.clk.Advance(time.Millisecond)
	assert.Equal(t, publication{version: 5}, r.next(t), "the burst publishes once, for the last version")
	r.quiet(t, ownedURI)
}

func TestSupersededParseNeverPublishes(t *testing.T) {
	started, cancelled := make(chan struct{}), make(chan struct{})
	r := newRig(t, func(ctx context.Context, text string) error {
		if text == "v2" {
			close(started)
			<-ctx.Done()
			close(cancelled)
			// It returns a tree, as a parser that never polls would.
		}
		return nil
	})
	r.s.parseDelay, r.s.diagDelay = 0, 0
	r.s.open(ownedURI, 1, []byte("v1"))
	assert.Equal(t, 1, r.next(t).version)

	r.s.change(ownedURI, 2, []byte("v2"))
	await(t, started, "v2's parse starts")
	r.s.change(ownedURI, 3, []byte("v3"))
	await(t, cancelled, "the newer edit cancels v2's parse")

	assert.Equal(t, 3, r.next(t).version)
	r.quiet(t, ownedURI)
	assert.Equal(t, 3, r.s.snapshot(ownedURI).version)
}

func TestCancelAfterParseCompletesNeverPublishes(t *testing.T) {
	proceed, returned := make(chan struct{}), make(chan struct{})
	r := newRig(t, func(ctx context.Context, text string) error {
		if text == "v2" {
			<-proceed
			defer close(returned)
		}
		return nil
	})
	r.s.parseDelay, r.s.diagDelay = 0, 0
	r.s.open(ownedURI, 1, []byte("v1"))
	assert.Equal(t, 1, r.next(t).version)

	r.s.change(ownedURI, 2, []byte("v2"))
	o := r.s.owner(ownedURI)
	// On the owner's goroutine, so v2's result cannot land first: v2's
	// parse finishes, uncancelled, and only then does v3 arrive.
	o.call(func() {
		close(proceed)
		<-returned
		o.applyEdit(3, []byte("v3"))
	})

	assert.Equal(t, 3, r.next(t).version, "only the current generation publishes")
	r.quiet(t, ownedURI)
}

func TestEditsToDifferentDocumentsRunConcurrently(t *testing.T) {
	var arrived sync.WaitGroup
	arrived.Add(2)
	both := make(chan struct{})
	go func() { arrived.Wait(); close(both) }()
	met := make(chan bool, 2)
	r := newRig(t, func(ctx context.Context, text string) error {
		if text == "edit" {
			// A barrier both parses must reach: behind one lock, the
			// first would wait here for a second that never starts. The
			// deadline only bounds that hang.
			arrived.Done()
			select {
			case <-both:
				met <- true
			case <-time.After(10 * time.Second):
				met <- false
			}
		}
		return nil
	})
	r.s.parseDelay, r.s.diagDelay = 0, 0
	r.s.open("file:///a.pl", 1, []byte("a"))
	r.s.open("file:///b.pl", 1, []byte("b"))

	r.s.change("file:///a.pl", 2, []byte("edit"))
	r.s.change("file:///b.pl", 2, []byte("edit"))
	assert.Equal(t, 2, r.s.snapshot("file:///a.pl").version)
	assert.Equal(t, 2, r.s.snapshot("file:///b.pl").version)
	assert.True(t, <-met && <-met, "both documents parse at once")
}

func TestParseSlotsCapConcurrentParses(t *testing.T) {
	limit := cap(parseSlots)
	require.Equal(t, max(2, runtime.NumCPU()/2), limit)

	var mu sync.Mutex
	var fill sync.Once
	inside, most := 0, 0
	full, release := make(chan struct{}), make(chan struct{})
	r := newRig(t, func(ctx context.Context, text string) error {
		mu.Lock()
		inside++
		most = max(most, inside)
		if inside == limit {
			fill.Do(func() { close(full) }) // later parses refill the slots once they free
		}
		mu.Unlock()
		<-release
		mu.Lock()
		inside--
		mu.Unlock()
		return nil
	})
	uris := make([]string, limit+2)
	for i := range uris {
		uris[i] = fmt.Sprintf("file:///slot%d.pl", i)
		r.s.open(uris[i], 1, []byte("x"))
	}
	await(t, full, "the slots fill")
	assert.Len(t, parseSlots, limit, "every slot is held")
	close(release)
	for _, uri := range uris {
		require.NotNil(t, r.s.snapshot(uri))
	}
	assert.Equal(t, limit, most, "no more than the cap parse at once")
	assert.Len(t, r.computed(), len(uris), "the rest parse once a slot frees")
}

// change sends a didChange of uri to version holding text.
func change(t *testing.T, h *handler, uri string, version int, text string) {
	t.Helper()
	params, err := json.Marshal(lspDidChangeParams{
		TextDocument:   lspVersionedTextDocumentIdentifier{URI: uri, Version: version},
		ContentChanges: []lspTextDocumentContentChangeEvent{{Text: text}},
	})
	require.NoError(t, err)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didChange", Params: params}))
}

func TestDidCloseStopsOwner(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true
	openDoc(t, h, out, ownedURI, "my $x = 1;\n")
	o := h.server.owner(ownedURI)
	require.NotNil(t, o)

	params, err := json.Marshal(lspDidCloseParams{TextDocument: lspTextDocumentIdentifier{URI: ownedURI}})
	require.NoError(t, err)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didClose", Params: params}))
	select {
	case <-o.done:
	default:
		t.Fatal("didClose returns with the owner's goroutine still running")
	}
	assert.Nil(t, h.server.owner(ownedURI))

	out.Reset()
	change(t, h, ownedURI, 2, "my $y = 2;\n")
	assert.Nil(t, h.server.owner(ownedURI), "a change does not resurrect the document")
	assert.Nil(t, h.server.Source(ownedURI))
	assert.Zero(t, out.Len(), "nothing is published for a closed document")
}

func TestDidChangeForUnopenedDocumentIgnored(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true
	change(t, h, "file:///never.pl", 2, "my $x = 1;\n")
	assert.Nil(t, h.server.owner("file:///never.pl"), "no owner is made")
	assert.Nil(t, h.server.Source("file:///never.pl"))
	assert.Zero(t, out.Len())
}

func TestStaleVersionDidChangeIgnored(t *testing.T) {
	h, _ := newTestHandler(t)
	h.initialized = true
	params, err := json.Marshal(lspDidOpenParams{TextDocument: lspTextDocumentItem{URI: ownedURI, LanguageID: "perl", Version: 5, Text: "my $v5;\n"}})
	require.NoError(t, err)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: params}))

	change(t, h, ownedURI, 5, "my $same;\n")
	change(t, h, ownedURI, 3, "my $older;\n")
	assert.Equal(t, "my $v5;\n", string(h.server.Source(ownedURI)), "a version not newer is not applied")

	change(t, h, ownedURI, 6, "my $v6;\n")
	assert.Equal(t, "my $v6;\n", string(h.server.Source(ownedURI)), "control: a newer version is")
}

func TestDebounceResetNoSpuriousFire(t *testing.T) {
	r := newRig(t, nil)
	r.s.diagDelay = time.Hour
	r.s.open(ownedURI, 1, []byte("v1"))
	require.NotNil(t, r.s.snapshot(ownedURI))
	o := r.s.owner(ownedURI)

	r.s.change(ownedURI, 2, []byte("v2"))
	// The parse timer fires while the owner is busy, so its fire is sitting
	// in the channel when the next edit resets it.
	o.call(func() {
		r.clk.Advance(parseDebounce)
		o.applyEdit(3, []byte("v3"))
	})
	var started uint64
	for range 5 {
		o.call(func() { started = o.started })
	}
	assert.EqualValues(t, 1, started, "the stale fire starts no parse; v3's debounce is still running")

	r.clk.Advance(parseDebounce)
	doc := r.s.snapshot(ownedURI)
	assert.Equal(t, 3, doc.version)
	assert.Equal(t, []string{"v1", "v3"}, r.computed(), "one parse, of the post-edit text")
}

func TestRequestDuringDebounceServesCurrentVersion(t *testing.T) {
	r := newRig(t, nil)
	r.s.open(ownedURI, 1, []byte("v1"))
	require.NotNil(t, r.s.snapshot(ownedURI))

	r.s.change(ownedURI, 2, []byte("v2")) // the fake clock never fires its debounce
	assert.Equal(t, "v2", string(r.s.Text(ownedURI).Text), "the request forces the parse")
	assert.Equal(t, 2, r.s.snapshot(ownedURI).version)
}

func TestOwnerSurvivesParsePanic(t *testing.T) {
	r := newRig(t, func(ctx context.Context, text string) error {
		if text == "v2" {
			panic("boom")
		}
		return nil
	})
	r.s.parseDelay, r.s.diagDelay = 0, 0
	r.s.open(ownedURI, 1, []byte("v1"))
	assert.Equal(t, 1, r.next(t).version)

	r.s.change(ownedURI, 2, []byte("v2"))
	p := r.next(t)
	assert.Equal(t, 2, p.version)
	require.Error(t, p.err, "the panic is v2's error")
	assert.Contains(t, p.err.Error(), "boom")
	assert.Equal(t, 1, r.s.snapshot(ownedURI).version, "requests are served the last good tree")

	r.s.change(ownedURI, 3, []byte("v3"))
	assert.Equal(t, publication{version: 3}, r.next(t), "the next edit publishes")
}

// TestOwnerSurvivesParsePanic covers the owner; this covers the handler
// turning the panic into one error diagnostic over the whole document.
func TestOwnerParsePanicPublishesErrorDiagnostic(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true
	h.server.compute = func(context.Context, string, []byte) (*document, error) { panic("boom") }
	h.server.open(ownedURI, 1, []byte("a\nbc"))
	settle(h, ownedURI)
	var params lspPublishDiagnosticsParams
	require.NoError(t, json.Unmarshal(readResponse(t, out).Params, &params))
	require.Len(t, params.Diagnostics, 1)
	assert.Contains(t, params.Diagnostics[0].Message, "boom")
	assert.Equal(t, lspPosition{Line: 1, Character: 2}, params.Diagnostics[0].Range.End)
}

func TestShutdownWithParseInFlight(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true
	entered, gate, left := make(chan struct{}), make(chan struct{}), make(chan struct{})
	release := sync.OnceFunc(func() { close(gate) })
	t.Cleanup(release)
	h.server.compute = func(context.Context, string, []byte) (*document, error) {
		defer close(left)
		close(entered)
		<-gate // never polls ctx, the worst case
		return &document{analysis: &infer.FileAnalysis{}}, nil
	}
	h.server.open(ownedURI, 1, []byte("my $x;\n"))
	await(t, entered, "the parse starts")
	o := h.server.owner(ownedURI)

	rawID := json.RawMessage(`1`)
	began := time.Now()
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", ID: &rawID, Method: "shutdown"}))
	took := time.Since(began)
	if os.Getenv("PSC_LATENCY_BUDGET") == "1" {
		assert.Less(t, took, 100*time.Millisecond, "shutdown returns within 100 ms")
	}

	select {
	case <-o.done:
	default:
		t.Fatal("shutdown returns with the owner running")
	}
	resp := readResponse(t, out)
	assert.Equal(t, json.RawMessage("null"), resp.Result)
	release()
	await(t, left, "the parse returns")
	rest, err := io.ReadAll(out)
	require.NoError(t, err)
	assert.Empty(t, rest, "nothing is published after shutdown")
}

func TestReopenPublishesNoStaleDiagnostics(t *testing.T) {
	r := newRig(t, nil)
	r.s.parseDelay, r.s.diagDelay = 0, 0
	gate, publishing, v2 := make(chan struct{}), make(chan struct{}), make(chan struct{})
	r.s.compute = func(_ context.Context, _ string, text []byte) (*document, error) {
		if string(text) == "v1" {
			<-gate
		}
		return &document{analysis: &infer.FileAnalysis{}}, nil
	}
	var mu sync.Mutex
	var order []int
	r.s.publish = func(doc *document, _ error) {
		if doc.version == 1 {
			// The old owner is mid-publish as the reopen arrives: hold it
			// there, so a new owner already parsing would publish first.
			// The bound only ends the hold once the reopen is stopping
			// this owner instead; it decides nothing.
			close(publishing)
			select {
			case <-v2:
			case <-time.After(50 * time.Millisecond):
			}
		}
		mu.Lock()
		order = append(order, doc.version)
		mu.Unlock()
		if doc.version == 2 {
			close(v2)
		}
	}

	r.s.open(ownedURI, 1, []byte("v1"))
	close(gate)
	await(t, publishing, "v1's diagnostics are being published")
	r.s.open(ownedURI, 2, []byte("v2"))
	await(t, v2, "v2's diagnostics are published")

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []int{1, 2}, order, "the superseded owner's diagnostics never follow the new owner's")
}

func TestNoPublishAfterDidClose(t *testing.T) {
	h, out := newTestHandler(t)
	h.initialized = true
	gate, publishing := make(chan struct{}), make(chan struct{})
	h.server.compute = func(context.Context, string, []byte) (*document, error) {
		<-gate
		return nil, fmt.Errorf("stale")
	}
	var o *owner
	h.server.publish = func(doc *document, err error) {
		close(publishing)
		<-o.quit // publish only once the close has begun stopping the owner
		h.published(doc, err)
	}
	h.server.open(ownedURI, 1, []byte("my $x;\n"))
	o = h.server.owner(ownedURI)
	close(gate)
	await(t, publishing, "the owner is publishing")

	params, err := json.Marshal(lspDidCloseParams{TextDocument: lspTextDocumentIdentifier{URI: ownedURI}})
	require.NoError(t, err)
	require.NoError(t, h.handle(&jsonRPCMessage{JSONRPC: "2.0", Method: "textDocument/didClose", Params: params}))

	var last *lspPublishDiagnosticsParams
	for tr := newTransport(out, io.Discard); ; {
		msg, err := tr.readMessage()
		if err != nil {
			break // the end of what was written
		}
		var p lspPublishDiagnosticsParams
		require.NoError(t, json.Unmarshal(msg.Params, &p))
		if msg.Method == "textDocument/publishDiagnostics" && p.URI == ownedURI {
			last = &p
		}
	}
	require.NotNil(t, last, "the close clears the document's diagnostics")
	assert.NotNil(t, last.Diagnostics, "the clear is an empty array, not null")
	assert.Empty(t, last.Diagnostics, "the client's last word for a closed document is an empty clear")
}
