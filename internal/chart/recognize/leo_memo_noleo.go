//go:build chart_noleo

// ABOUTME: The Leo memo step as a no-op, under the chart_noleo build tag, whatever Options.NoLeo says.
// ABOUTME: It exists only so that TestLeoLinearAtLookaheadZero can be shown to fail when Leo does nothing.

package recognize

// memoAt makes no memo.
func (c *Chart) memoAt(at, sym int) *memo { return nil }
