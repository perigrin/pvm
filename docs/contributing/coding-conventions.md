---
stability: 1
covers: []
---

## Code Style

Follow the conventions of the surrounding code. Consistency within a file
matters more than strict adherence to an external style guide.

## Naming

Use clear, descriptive names. Avoid abbreviations unless they are universally
understood in context.

## Error Handling

Return errors rather than panicking. Wrap errors with context using
fmt.Errorf("context: %w", err).

## Testing

Write a failing test before implementation (TDD). Tests use real data — no
mocks or stubs.
