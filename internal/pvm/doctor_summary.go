// ABOUTME: Summary tip generation for the doctor command output
// ABOUTME: Picks the right wording for the --fix suggestion based on whether anything is actually fixable

package pvm

import (
	"os"

	"tamarou.com/pvm/internal/backup"
	"tamarou.com/pvm/internal/cli/ui"
	"tamarou.com/pvm/internal/shell"
	"tamarou.com/pvm/internal/xdg"
)

// doctorSummaryTip returns the trailing tip line shown after the doctor
// surfaces issues or warnings. When the auto-fix engine has at least one
// applicable fix, the tip pitches `--fix` directly. When nothing is
// fixable (typically because the user just hasn't installed any Perl yet),
// the tip is reworded so it does not promise auto-resolution that won't
// happen, but still points at the command for discoverability.
func doctorSummaryTip(fixable bool) string {
	if fixable {
		return "Run 'pvm self doctor --fix' to automatically resolve common issues " +
			"(or 'pvm self doctor --dry-run' to preview)"
	}
	return "No automatic fixes are available for the items above — see the " +
		"per-issue suggestions, or run 'pvm self doctor --dry-run' to confirm"
}

// hasApplicableAutoFixes inspects the auto-fix engine against the current
// environment and reports whether any fix would actually run. Returns false
// if the environment cannot be probed (we'd rather suppress a misleading
// pitch than point the user at a command we can't validate).
func hasApplicableAutoFixes(out *ui.Output) bool {
	engine := NewAutoFixEngine()

	backupMgr, err := backup.NewManager()
	if err != nil {
		return false
	}
	shellMgr, err := shell.NewConfigManager()
	if err != nil {
		return false
	}
	shellConfig, err := shellMgr.DetectShellConfig()
	if err != nil {
		return false
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	xdgDirs, err := xdg.GetDirs()
	if err != nil {
		return false
	}

	ctx := &FixContext{
		UI:        out,
		DryRun:    true,
		Session:   backupMgr.CreateSession("doctor-summary-probe"),
		Shell:     shellConfig,
		HomeDir:   homeDir,
		XDGDirs:   xdgDirs,
		BackupMgr: backupMgr,
		ShellMgr:  shellMgr,
	}
	return engine.HasApplicableFixes(ctx)
}
