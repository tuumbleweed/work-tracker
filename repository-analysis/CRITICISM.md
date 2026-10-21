# Review notes

## Remaining Issues

- **Daily rollover is absent (confirmed by static inspection).** `InitializeTrackerApp` sets the daily file and totals once; `flushChunkIfRunning` continues using that path indefinitely. An app left open overnight attributes later work to the startup day. Detect day changes, split intervals at midnight, and reload/reset daily counters (`src/pkg/tracker-app/initialize-app.go`, `tracker-app-methods.go`).
- **Idle-detector failure becomes active time (confirmed).** `tryXprintidle` returns -1 on failure; `refreshActivityState` treats it as input during the interval and credits the full interval. Missing xprintidle or an incompatible desktop session silently inflates activity. Represent unavailable measurements explicitly and surface detector failures (`helpers.go`, `tracker-app-methods.go`).
- **Shared UI state can race (static concern, not reproduced).** `updateInterface` retains the live `TimeByTask` map after releasing the mutex, then reads it in a queued UI callback while timer code can write it. It also reads some fields outside the lock. Copy a consistent snapshot under the mutex and use only that snapshot in UI callbacks. Validate with targeted race testing.
- **Email error logging can expose content (confirmed).** `sha256Short` in `src/pkg/email/provider-wrappers.go` copies raw content into its output rather than hashing it. On delivery errors its alleged hashes can contain content prefixes. Use `crypto/sha256` or omit the fields. The send-email command also logs the entire report at verbose level; document or remove that behavior.
- **Coverage is narrow.** The only automated test file covers missed-time insertion. Timing transitions, persistence, rollover, report aggregation, and provider dispatch have no local automated coverage. Prioritize deterministic clock and idle-detector tests over live desktop/provider tests.

## Simplification Opportunities

- Consolidate the JSONL schema, path construction, and validation currently repeated in tracker, report, and cheat packages into a small UI-independent storage package. Their malformed-record and active-duration policies differ; the report reader also inserts a synthetic one-nanosecond unassigned task to reserve a color. Reserve colors explicitly instead of altering totals.
- Validate credentials for the selected provider only. `src/cmd/send-email/main.go` checks variables for all providers before parsing flags, requiring dummy values for unused services. Remove that setup burden and consider retiring unused SES v1 helpers after checking external callers.
- Make report opening optional and use an explicit template path or embed the template. The report command always starts `google-chrome`; template loading depends on the project working directory and panics on errors (`src/pkg/util/chrome.go`, `src/pkg/report/render-html.go`). A file-only mode would make automation and headless use clearer.
