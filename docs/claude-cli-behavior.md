# Claude CLI Behavior Notes (T1.7)

## Verified Behavior (as of manual checks)

- `claude --help` is present and the CLI can be invoked from PATH.
- Basic launch without flags opens an interactive session.

## Items to Validate (TODO)

- `claude --resume <session_name>` behavior when the session is missing (does it create a new session or error?).
- Whether a `--session-name`/`--name` flag exists to avoid keystroke-based renaming.
- Minimum safe delay, if keystrokes are required for renaming.

## Fallback Strategy

1) Detect capabilities at runtime (planned):
   - Parse `claude --help` to discover resume/name flags.
   - Use flags when available; otherwise fall back to keystrokes with a configurable delay.
2) If resume fails, warn the user and offer manual session selection or fresh launch.
3) Allow overrides via environment variable (e.g., `CCW_CLAUDE_MODE`) once detection is implemented.

## Compatibility Matrix (to fill as versions are tested)

| Claude CLI Version | Resume Support | Rename Method | Status      |
|--------------------|----------------|---------------|-------------|
| (TBD)              | (TBD)          | (TBD)         | Needs test  |

Update this document as real versions are verified during release testing.
