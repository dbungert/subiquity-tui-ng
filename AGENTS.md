# AGENTS.md

This file is a curated summary of assumptions and constraints for
`subiquity-ng` — a Bubble Tea rewrite POC of Ubuntu's subiquity
installer. **Maintain this file as you work.** If you discover or
decide something non-obvious, add it here so the next agent does not
have to re-derive it. Keep it terse; this is not a changelog.

## What this project is

Proof-of-concept rewrite of subiquity in Go using Bubble Tea +
Lipgloss. Upstream subiquity is Python + urwid and lives at
`/project/megademo/subiquity/`. We are reimplementing the client-side
TUI, not the server or the curtin integration (yet).

## What this project is not (yet)

- Not feature-complete with upstream. Screens land one at a time.
- Not a port — it's a rewrite. We don't need to mirror urwid
  abstractions; Bubble Tea / Lipgloss already give us scrolling,
  sized widgets, and styled text natively.

## Inherited UX rules (from upstream DESIGN.md, still apply)

- **80x24 minimum.** Must work in an 80x24 terminal. Smaller terminals
  may render badly but must not crash (SSH users resize at will).
- **Never block the UI > 0.1s.** Anything slower runs in the
  background. If it shows indication, show it for ≥ 1s to avoid
  flicker. In Bubble Tea this means `tea.Cmd` for I/O, never sync work
  in `Update`/`View`.
- **Up / Down / Space / Enter + occasional typing.** That's the entire
  interaction vocabulary. No mouse-required affordances.
- **Prevent invalid input** rather than reject after the fact (e.g.,
  filter keystrokes that can't appear in a unix username). When you
  must reject, explain what's valid.
- **Bias for the common case.** Set initial focus / default selection
  so the common path is one Enter away.

## Screen anatomy (from upstream)

A subiquity screen has:
1. Header — 3-line Ubuntu orange band, summary on left, `[ Help ]` on right.
2. Body — usually: excerpt (explains the screen) + scrollable content
   + button stack with done/cancel.

We replicate the header in `internal/ui/chrome.go`. Body layout is the
screen's own responsibility for now; if a body-layout helper emerges,
add it to `internal/ui/`.

## Architecture conventions for this repo

- **`Screen` interface** (`internal/screens/screen.go`): each screen
  implements `Init / Update / View(w, h) / Title`. The top-level
  `Model` in `main.go` owns the current screen and delegates. Screen
  swap-out is not yet implemented but is the next obvious extension.
- **Chrome is owned by the top-level Model**, not the screen. Screens
  must not render the orange header themselves. They receive a body
  rectangle of `(width, height - ui.HeaderHeight)`.
- **No `q` quit binding.** Future screens will have text input where
  `q` is a valid character. Only `ctrl+c` quits.
- **Ubuntu orange = `#E95420`.** Hardcoded in `internal/ui/style.go`.
  Linux-tty palette (`PIO_CMAP`) is still a future concern; the
  half-block trick assumes the terminal renders `▀` and `▄`
  contiguously and uses UTF-8. `internal/ui/capability.go` falls back
  to 3 solid orange rows when `TERM` is `linux` / `dumb` / empty or
  the locale isn't UTF-8. Override via `SUBIQUITY_NG_HEADER=blocks|plain`.
- **Header band is 3 rows but reads as 2 visual lines.** Rows 1 and 3
  are `▄` / `▀` (orange fg, terminal-default bg) so only half the
  cell is orange; row 2 is fully orange with the title vertically
  centered. Glyph constants and `HalfBlockStyle` live in
  `internal/ui/style.go`; `HeaderHeight = 3` is still the row budget.
- **Measure with `lipgloss.Width`, not `len`.** Titles contain
  multibyte runes (`Добро пожаловать!`); byte length is wrong.
- **Lint is gated by pre-commit.** `make` (default target) runs
  `lint → test → build`; `make init` installs the git pre-commit hook
  once per clone. Config in `.golangci.yml` and `.pre-commit-config.yaml`.
  Bumping the `golangci-lint` rev: `pre-commit autoupdate`.
- **Tests aim for high coverage, not 100%.** `make test` prints the
  per-package percentage; `make coverage` writes `.coverage.out`
  (gitignored) and prints the per-function report. Don't chase
  coverage for trivial getters, `main()`, or interface types — focus
  on the rendering and capability logic.
- **Commits represent one logical change.** Commas or "and" in the
  subject are a smell — usually a sign to split. Closely related
  housekeeping can ride together (e.g., adding several pre-commit
  hooks at once is one logical change even if the subject lists them),
  but distinct concerns don't bundle just because they happen to be
  chores.

## Screen flow and API integration

- **Server interaction.** HTTP client in `internal/client/client.go` communicates
  with subiquity server over unix socket at `/run/subiquity/socket` (or
  `.subiquity/socket` in test mode). Async commands (`tea.Cmd`) fetch data and
  post selections back.
- **Screen sequence:** Language → Source → Storage (disk selection & capability) →
  Confirm (destructive-action confirmation) → Keyboard → ... See `main.go` for
  handler routing and message types.
- **Subiquity API: body vs. query parameters.** Check `apidef.py` in upstream
  subiquity: parameters with `Payload` type in the handler signature go in the
  request body (JSON-encoded); everything else goes in query parameters.
  Query parameters must be JSON-encoded before URL encoding. E.g., `tty=/dev/tty1`
  becomes `?tty=%22%2Fdev%2Ftty1%22` (URL-encoded `"/dev/tty1"`). Arrays in query
  params: JSON-encode the whole array (e.g., `?endpoint_names=%5B%22network%22%5D`
  for `["network"]`). This is pervasive; consult apidef.py rather than guessing.

## Language screen (`internal/screens/language.go`)

- **First screen presented** in `main.go` on app start.
- **Live type-ahead filter.** Typing any character appends to the filter;
  the list narrows to entries whose native or English name contains the query
  (case-insensitive). Backspace/Ctrl+H removes the last rune; Esc clears
  the filter entirely.
- **Language data** is hardcoded from `/project/megademo/subiquity/languagelist`,
  sorted by English name, with both native and English names added. Upstream
  sorts by native name; we sort by English to make the ASCII-primary index more
  predictable.
- **Pointer receiver** (`*LanguageScreen`) is used because the screen owns mutable
  list/scroll state. It implements `Screen` via pointer.
- **`langOverhead = 6`** captures the fixed lines consumed outside the list
  (description, blank, search, blank, blank-before-hints, hints). `listH = height - 6`.
- **Body design diverges from upstream.** The body is open to new ideas per project
  rules. Current design: description → search line with `█` cursor and live match
  count → scrollable list with `> ` selection prefix and orange highlight row →
  key-hint bar. No button stack (Enter directly confirms inline).
- **On Enter,** posts locale to server and transitions to Source screen.

## Future scope worth knowing about

- **Autoinstall mode** in upstream bypasses the TUI entirely. Not in
  scope yet, but the `Screen` interface should not assume interactivity
  is mandatory forever.
- **Keyboard / Network / Proxy / Identity screens** come next in the upstream
  sequence. Consult `subiquity.client.client` in the upstream tree for
  ordering and interactions.
- **Error recovery.** Current screens transition linearly; no back-button or
  error-state handling. Real subiquity allows rework of previous screens.

## Where the upstream lives

- Code: `/project/megademo/subiquity/`
- Design doc: `/project/megademo/subiquity/DESIGN.md`
- Client controllers: `subiquity/client/controllers/` in that tree.
- Views: `subiquity/ui/views/` in that tree.

## Maintenance contract

- When you make a decision that affects future work (e.g., "we don't
  use lipgloss tables, we render manually because X"), record it here.
- When an assumption above turns out to be wrong, **edit it in
  place**. Don't append. This file shrinks and shifts; it does not
  grow monotonically.
- Don't paste large chunks of upstream docs here. Link out instead.
- If a section is no longer true and has no replacement, delete it.
