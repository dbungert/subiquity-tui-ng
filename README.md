# subiquity-ng

**⚠️ Proof of Concept — Not Product Direction**

This is an experimental rewrite of Ubuntu's [subiquity](https://github.com/canonical/subiquity) installer client as a Terminal User Interface (TUI) in Go, using [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

**This is not an official Ubuntu product.** It exists to explore alternative implementation approaches and is not intended to replace the production subiquity installer. Use the official [subiquity](https://github.com/canonical/subiquity) for real Ubuntu installations.

## About

The upstream subiquity is written in Python using the urwid TUI framework. This proof of concept reimplements the client-side TUI in Go to explore:

- Modern Go practices for terminal UI
- Bubble Tea + Lipgloss as an alternative to urwid
- A cleaner, more maintainable codebase for interactive terminal applications

## Current Status

**Implemented screens:**
- Language selection (live type-ahead filter, 33 languages)
- Source/mirror selection
- Storage/disk selection with guided partitioning
- User identity (username, realname, password with confirmation)
- Hostname selection
- Installation confirmation (destructive action warning)
- Install progress (long-polling with real-time journalctl log streaming)
- Reboot confirmation

**Server integration:**
- Unix socket HTTP client (`internal/client/client.go`)
- Full API integration: locale, source, storage, identity, install monitoring, reboot
- Password hashing (SHA-512 crypt(3) format)
- Long-polling for install state transitions with 5-minute timeout
- Real-time log streaming from `journalctl` during installation

**Architecture:**
- 80x24 terminal minimum with proper behavior on larger displays
- Screen-based navigation model with state-driven transitions
- Ubuntu orange header with title
- Keyboard-only interaction (no mouse)
- All I/O asynchronous (Bubble Tea commands, no blocking)

**Not implemented:**
- Keyboard layout configuration
- Network configuration
- Proxy configuration
- Autoinstall mode

## Building & Running

### Requirements
- Go 1.25+
- Standard POSIX tools

### Build
```bash
make build
```

Produces `./subiquity-client` binary.

### Run
```bash
make run
```

Or directly:
```bash
./subiquity-client
```

### Test & Lint
```bash
make test        # Run tests
make coverage    # Generate coverage report
make lint        # Run linters (pre-commit hooks)
make check       # Full pipeline: lint → test → build
```

## Design Principles

Inherited from upstream subiquity:

- **80x24 minimum**: Works on small terminals, scales gracefully to wide displays
- **Never block > 0.1s**: All I/O happens asynchronously (Bubble Tea commands)
- **Keyboard only**: Up/Down/Enter/typing for navigation and input
- **Prevent invalid input**: Filter keystrokes rather than reject after the fact
- **Common case first**: English pre-selected, sensible defaults throughout

On wide terminals (>120 columns), content is centered for better visual spacing.

## Code Organization

```
.
├── cmd/subiquity-client/    # Main entry point
│   ├── main.go              # App model and TUI loop
│   └── main_test.go         # Tests for top-level Model
├── internal/
│   ├── screens/             # Individual screens (Language, Keyboard, etc.)
│   │   ├── screen.go        # Screen interface
│   │   ├── language.go      # Language selection screen
│   │   └── *_test.go        # Screen tests
│   └── ui/                  # Shared UI chrome and styling
│       ├── chrome.go        # Header, body, centering
│       ├── style.go         # Ubuntu branding (colors, glyphs)
│       ├── capability.go    # Terminal capability detection
│       └── *_test.go        # UI tests
├── AGENTS.md                # Development notes & conventions
├── Makefile                 # Build & test targets
└── go.mod / go.sum          # Dependencies (Bubble Tea, Lipgloss)
```

## Development

Each screen implements the `Screen` interface:

```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View(width, height int) string
    Title() string
}
```

Add new screens to `internal/screens/` and wire them up in the top-level Model. See `internal/screens/language.go` and `internal/screens/language_test.go` for a complete example.

### Conventions

- **One logical change per commit**: See `AGENTS.md` for details
- **High test coverage on logic, not UI**: Tests focus on rendering correctness and interaction, not 100% coverage
- **Multibyte-aware**: Use `lipgloss.Width()` instead of `len()` for Russian/Chinese/emoji text
- **Terminal capability detection**: Graceful fallback from half-blocks (▀▄) to plain rendering on limited terminals

## Upstream Context

For comparison or integration context, see:

- [subiquity repository](https://github.com/canonical/subiquity)
- [Upstream design doc](https://github.com/canonical/subiquity/blob/main/DESIGN.md)
- [Language list](https://github.com/canonical/subiquity/blob/main/languagelist)

## License

Licensed under the GNU General Public License v3 (GPLv3). See [LICENSE](./LICENSE) for details.

This proof of concept is experimental work created to explore alternative implementation strategies for the Ubuntu installer. It is not a committed product and does not represent official product direction.

## Questions or Feedback?

This is a learning/exploration project. Issues and PRs are not expected or monitored. For real Ubuntu installation issues, use the official [subiquity](https://github.com/canonical/subiquity).
