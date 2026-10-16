# Evidence and review record

## Scope and screening

Full analysis: the repository contains substantive task tracking, persistence, reporting, desktop integration, and a meaningful feature history. Initial scaffolding and reusable helpers are not counted as original product features.

Inspected local `main`, `origin/main`, `origin/HEAD` (all at `bf7551cce4925eeabea9611923cc65e2b85c5da0`) and tag `v0.0.1` (`a5aa5735705d05a20652f9eac18cd37dcf9f55f3`). The tag is on the available main history. No additional local branches or tags were present; no remote fetch was performed. Reviewed all 152 commit subjects, initial-layout history, selected milestone changes, root README, INSTALL, TODO, module manifest, configuration/environment examples, scripts, tracker state/UI/persistence, report aggregation/rendering, missed-time implementation/tests, and email dispatch/provider code. No prior analysis directory existed and the initial working tree was clean. No applicable AGENTS.md was found in the repository or its inspected parent subtree.

Analysis is static. Repository programs, GUI, tests, and network delivery were not executed, respecting the instruction against running untrusted programs without authorization. Build success, runtime stability, graphical appearance, and live provider compatibility remain unverified. The four distinct screenshots displayed by the root README were visually inspected during the image-requirements update. Both generated public documents now embed all four using verified `../pictures/` paths; the repeated running screenshot in the root README is displayed once per output. No external research was necessary; no claim is made about present provider SDK support or vulnerability status.

## Claim ledger

| Claim | Local evidence | Confidence |
|---|---|---|
| Personal utility with intermittent maintenance | Owner's supplementary context; tracking features and commit distribution independently support purpose and update pattern | High for purpose/pattern; actual usage frequency is owner-reported |
| Substantive application built after a reusable project layout | `7e3823d` initial layout, `2093de3` tracker start, `eec4b0e` project rename; placeholder entrypoints and generic config helpers remain | High for development sequence; provenance of reusable scaffolding is uncertain |
| Task controls and daily totals in Fyne | `src/pkg/tracker-app/initialize-interface.go`, `tasks-ui.go:makeTasksUI`, `tracker-app-methods.go:refreshUIState`; `760c9d8`, `6c54525` | High: implemented paths inspected |
| Independent UI/activity/save loops and UI dispatch | `tracker-app-methods.go:Start`, `uiTickLoop`, `activityTickLoop`, `flushTickLoop`, `updateInterface`; `415e9da` | High for architecture; synchronization correctness not verified |
| Input activity is an all-or-nothing interval heuristic | `helpers.go:tryXprintidle`, `tracker-app-methods.go:refreshActivityState`; `f834d6e` | High: explicit threshold comparison; failure returns -1 and takes the active branch |
| JSONL interval persistence and startup restoration | `flush-chunk.go:flushChunk`, `appendChunk`; `load-todays-data.go:loadFileActivityAndDuration`; `initialize-app.go`; `5aa8410`, `57486eb` | High: timestamps, nanosecond duration, per-task totals and year/month paths are explicit |
| Task transitions preserve old-task attribution through a flush | `tasks-ui.go:rowPlayButton.OnTapped`, `tracker-app-methods.go:onButtonTapped`; `6c54525` | High for intended implemented sequence; no timing-boundary tests |
| Report aggregation and adaptive HTML charts | `src/pkg/report/build-report.go:BuildReport`, `ResolveRange`; `read-day-file.go`; `render-html.go:renderHTMLReport`; `cfg/report-template.html`; `6b8fe93`, `9a92bf4`, `bf7551c` | High: inspected implementation, no visual runtime verification |
| Optional provider-based email sending | `src/cmd/send-email/main.go:sendReport`, `src/pkg/email/provider-wrappers.go:SendMessage`; provider implementations; `9c4aad5`, `61525cc` | High for integration code; live delivery unverified. SDKs implement transport, not original infrastructure |
| Tray and Linux desktop launchers | `src/pkg/tracker-app/tray.go:initTray`; `scripts/install.sh`, `shared.sh`, desktop templates; `bc4fd89`, `295fd35` | High for code; driver support and desktop behavior unverified |
| Validated missed-time appends and tests | `src/pkg/cheat/append.go:Append`, `append_test.go:TestAppend`, `TestInvalidDoesNotModifyFile`, `TestBadFilesAndTaskOverride`; `453cdd5` | High for implementation/test presence; tests not executed |
| Recovery from malformed JSON differs across consumers | Tracker loader skips malformed JSON but rejects invalid intervals/active duration; report reader skips malformed records and clamps activity; cheat rejects malformed records; `3be5a0b` | High: direct inspection; avoid claiming uniform recovery |

## Timeline basis

Author-calendar dates from `git log --all`, with unique reachable commits counted once:

| Period | Commits | Interpretation |
|---|---:|---|
| 2025-10-30–2025-10-31 | 43 | Scaffold, tracking, persistence, task interface |
| 2025-11-01–2025-11-11 | 89 | Reports, email, installation, tracking refinements, documentation |
| 2026-01-14–2026-01-23 | 9 | Dependency/toolchain changes, template extraction, year/month file layout |
| 2026-02-03 | 5 | Date formatting fix, tray integration and cleanup, icons |
| 2026-09-11–2026-09-30 | 6 | Logging/recovery, dependency updates, missed-time command, chart colors |

Earliest observed commit: `7e3823d`, 2025-10-30. Latest: `bf7551c`, 2026-09-30. Main development is inferred to have largely concluded by 2025-11-11 (132/152 commits, about 87%); this is not a formal completion date. The densest four days, October 30–November 2, account for 112 commits. No observed commits in December 2025 or March–August 2026. Gaps do not establish non-use or lack of uncommitted development.

All commits use one author name; this supports a single-author local history but does not prove sole ownership of every line. Actual project start/end dates remain null. Metadata status `active` means maintained intermittently, based on the owner's recollection and recent commits, not continuous development. The repository URL is derived from the Git remote and its accessibility/visibility was not checked.

## Context cross-checks and uncertainties

- The owner's account of a useful personal desktop tracker with occasional updates agrees with implemented functionality and commit density. Personal use and value are supplementary testimony, not independently measured outcomes.
- The main development period appears finished by November 11, 2025, but later substantive additions include tray support and missed-time insertion. Avoid describing the whole project as frozen or formally completed.
- TODO still lists year/month separation as unfinished; `57486eb` and `390085a` implement that layout and its report reader. Treat this TODO item as stale. No legacy flat-file migration or fallback was found in current report path selection.
- Confirm authorship/provenance of shared email/config/scaffolding helpers before making exclusive originality claims. The initial layout includes existing logger/error implementations, later replaced with libraries at `9a5bf16`.
- Confirm desktop/session requirements and whether the owner restarts daily; X11 activity sampling is implemented, native Wayland activity support is not established by installing Wayland libraries.
- No usage metrics, productivity gains, user counts, benchmark results, commercial deployments, or verified delivery outcomes are available. Do not infer them from screenshots or commits.

## Publication risks

- Root LICENSE is MIT, but this does not independently establish authorship or provenance of all helpers or image assets. Review third-party dependencies/assets before redistribution.
- Environment examples contain placeholders, not evidence of live credentials. Avoid publishing real environment files, AWS configuration, local work logs, generated reports, task lists, or personal contact details. No credential values or personal addresses are copied here.
- Existing screenshots were visually reviewed and preserved in both public documents as required, without generating or copying media. They display task names, dates, hours, and activity percentages; the yearly report also contains a person-named meeting category. Review that personal reference and activity history before publication. No credential or contact details were apparent in the screenshots. Image inclusion does not change `review_required` publication status.
- Email errors can include raw content prefixes through `sha256Short`; verbose send-email logs include complete report HTML. Tracker malformed-line warnings can include record contents. Logs should be considered sensitive.
- `EVIDENCE.md` and `CRITICISM.md` are review-oriented but become visible if committed to a public repository. The website display flag does not make them private. Publication remains `review_required`, criticism display disabled, and no facts are marked owner-verified.
