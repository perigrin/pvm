# Issue: Make Tool Mappings Configurable

## Problem
Currently, tool name mappings (e.g., `perltidy` → `Perl-Tidy`) are hardcoded in `internal/tool/mapping.go`. Users cannot customize or extend these mappings for their own tools or preferences.

## Proposed Solution
Add configuration support for custom tool mappings while keeping built-in defaults.

## Design

### Configuration Format
Add a `tool_mappings` section to PVM config:

```toml
[tool_mappings]
# Custom user mappings
mytool = "My::Custom::Tool"
pt = "Perl-Tidy"  # Alias for perltidy

# Override built-in mappings if needed
perltidy = "Alternative::Tidy::Tool"
```

### Implementation Plan

1. **Extend Config Structure**
   - Add `ToolMappings map[string]string` to config types
   - Update config loading/saving to handle tool mappings

2. **Update ToolMapping Class**
   - Already has `configMappings` support (implemented but unused)
   - Load user mappings from config during initialization
   - Priority: config mappings override built-in mappings

3. **Configuration Commands**
   - `pvm config tool-mapping add <tool> <module>`
   - `pvm config tool-mapping remove <tool>`
   - `pvm config tool-mapping list`

4. **Documentation**
   - Update user guide with tool mapping configuration
   - Add examples for common use cases

### Benefits
- **User Flexibility**: Custom tools can be mapped without code changes
- **Project-Specific Mappings**: Different projects can use different tool configs
- **Override Built-ins**: Users can fix incorrect built-in mappings
- **Extensibility**: Easy to add new tools without touching core code

### Backward Compatibility
- All existing built-in mappings continue to work
- No breaking changes to existing workflows
- Config is optional - defaults work without configuration

## Related Work
- Consolidation of duplicate mappings (completed)
- Single source of truth in `internal/tool/mapping.go` (completed)

## Priority
**Low** - Enhancement for power users. Current system works for common use cases.

---
*Issue created during tool mapping consolidation work*