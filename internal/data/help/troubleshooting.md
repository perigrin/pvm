# Troubleshooting PVM Issues

💡 For command list: pvm -h
💡 For contextual help: pvm help

## Common Issues and Solutions

### Project Detection Issues
**Problem:** PVM doesn't recognize your project
**Solution:** Add a .perl-version file or cpanfile to your project root
**Command:** `echo "5.38.0" > .perl-version`

### Module Installation Issues
**Problem:** Modules fail to install
**Solution:** Check if you're in a project and have proper permissions
**Commands:** `pvm workspace status`, `pvm workspace doctor`

### Build Issues
**Problem:** Build fails with type errors
**Solution:** Check PSC configuration and fix type annotations
**Commands:** `pvm build --check-only`, `pvm workspace doctor`

### Environment Issues
**Problem:** Wrong Perl version being used
**Solution:** Check version resolution and shell integration
**Commands:** `pvm perl resolve`, `pvm shell setup`

## Diagnostic Commands

- Check overall workspace health: `pvm workspace doctor`
- View detailed workspace status: `pvm workspace status --json`
- Check dependency status: `pvm module status`
- Verify Perl version resolution: `pvm perl resolve`

## Getting Help

If you're still having issues:

1. Check the workspace status: `pvm workspace status`
2. Run the doctor: `pvm self doctor --fix`
3. Check verbose output: `pvm --verbose [command]`
4. Report issues at: https://github.com/anthropics/claude-code/issues
