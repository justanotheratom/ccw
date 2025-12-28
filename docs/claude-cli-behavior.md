# Claude CLI Behavior Notes (T1.7)

Assumptions to validate for future implementation:

- `claude --resume <session_name>` resumes an existing Claude Code session without prompting.
- Session rename support: Claude Code accepts rename commands triggered externally after a short delay; start with a 5 second delay as default and adjust if the CLI exposes a safer flag.
- Launch command uses `claude` binary from PATH; no additional flags are required for basic interactive mode.
- If `--resume` fails (unknown session), fallback should allow creating/selecting a session manually.

Open verification items:
- Confirm whether `--resume` creates a new session when the name does not exist or exits with an error.
- Confirm if there is a dedicated flag for setting the session name at launch to avoid sending rename keystrokes.
- Measure minimum safe delay before issuing a rename to avoid race conditions.

These assumptions will be revisited before building session management in Phase 4 to ensure the behavior matches the real CLI.
