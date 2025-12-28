# Claude CLI Behavior Notes (T1.7)

## Verified Behavior (Claude Code 2.0.76)

- `claude --help` available; no `--session-name`/`--name` flag advertised.
- `claude --resume <uuid>` is accepted; invalid UUIDs error immediately.
- `claude --resume <nonexistent-uuid>` returns: `No conversation found with session ID: <uuid>`.
- Basic launch without flags opens an interactive session.

## Items to Validate (still TODO)

- Minimum safe delay if keystrokes are required for renaming (no flag found).
- Behavior when resuming valid, existing sessions (requires a real session ID).

## Fallback Strategy

1) Detect capabilities at runtime (implemented in code):
   - Parse `claude --help` for resume/name flags (resume supported; no name flag found).
   - Use flags when available; otherwise fall back to keystrokes with a configurable delay.
2) If resume fails, warn the user and offer manual session selection or fresh launch (future improvement).
3) Allow overrides via environment variable (e.g., `CCW_CLAUDE_MODE`) once detection is implemented.

## Compatibility Matrix (to fill as versions are tested)

| Claude CLI Version | Resume Support          | Rename Method            | Status             |
|--------------------|-------------------------|-------------------------|--------------------|
| 2.0.76 (Code)      | Yes (`--resume <uuid>`) | No name flag; keystroke | Verified (no name) |

Update this document as real versions are verified during release testing.
