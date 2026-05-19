# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build       # compile to ./uberlauncher
make run         # go run (no binary)
make test        # go test ./...
make fmt         # format all Go files
make fmt-check   # verify formatting (used in CI)
make vet         # go vet ./...
make lint        # golangci-lint (must be installed)
make check       # fmt-check + vet + test + lint (full local gate)
```

Run a single test package: `go test ./internal/engines/...`

Git hooks (Lefthook): `pre-commit` runs fmt-check, vet, lint; `pre-push` runs tests.

## Architecture

UberLauncher is a BubbleTea TUI launcher. Skills push entries into a store during initialization; the UI renders them and re-ranks on every keystroke.

### Data flow

1. `cmd/uberlauncher/main.go` — creates all services, initializes each skill with a `skill.Context`, then starts the BubbleTea program
2. Skills call `ctx.Store.UpsertEntry()` during `Init()` to register entries (synchronous, before UI starts)
3. `internal/store/store.go` — thread-safe entry map; delegates ranking to the injected `Engine`
4. `internal/ui/model.go` — reads the store once on startup; re-queries on every keystroke; listens to `notifier.Events` for background task messages/errors
5. Background tasks started via `ctx.Runtime.Go()` report results back via `ctx.Notifier`

### Key interfaces (`internal/skill/skill.go`)

```go
type Skill interface {
    Id() string
    Init(ctx Context)
}

type Context struct {
    Runtime  Runtime   // goroutine lifecycle: Go() + HasCommand()
    Cache    Cache     // namespaced file cache: ReadFile() + WriteFile()
    Notifier Notifier  // UI feedback: ReportError() + ReportMessage() + SendNotification()
    Config   Config    // key-value config: Get() + GetInt() + GetBool() + Set()
    Store    Store     // entry registration: UpsertEntry()
}
```

### Adding a new skill

See `docs/sop/skill-constructor-pattern.md` for the full procedure. In short:
- Create `internal/skills/<name>/<name>.go`
- Implement `Id() string` and `Init(ctx skill.Context)`
- Expose `func New() skill.Skill`
- Register in `main.go`'s `skillList`

### Initialization order

Skills are initialized **synchronously** in `main.go` before BubbleTea starts. This means:
- `UpsertEntry` calls in `Init()` are safe (nobody is reading from any channel yet)
- Do not use blocking channel sends from within `Init()` — use `ctx.Runtime.Go()` for anything that needs a channel

### Event architecture

| Channel | Owner | What flows through it |
|---|---|---|
| `notifier.Notifier.Events` | `internal/notifier` | errors and messages from background tasks → UI |

Dynamic entry updates (store events) are not yet implemented. When a skill needs them, `store.Store` will own an `Events` channel following the same pattern. See `docs/gwi/channels-producer-owns.md`.

### Ranking (`internal/engines/`)

- `Engine` interface: `Rank(entries []entry.Entry, query string) []entry.Entry`
- `FuzzyEngine` (default): fuzzy match via `github.com/sahilm/fuzzy` on `Entry.Label`

### Built-in skills (`internal/skills/`)

| Skill | Id | Notes |
|---|---|---|
| `apps` | `apps` | Parses `.desktop` files; caches to `~/.cache/uberlauncher/apps/`; launches via `hyprctl` if available |
| `search` | `search` | Free-text; opens Google search via `xdg-open` |
| `todo` | `todo` | Todoist quick-add via API; reads token from `ctx.Config` |
| `system` | `system` | Shutdown / reboot via `systemctl` or `shutdown` |
| `wifi` | `wifi` | Known WiFi connections via `nmcli` |

### Linters enabled (`.golangci.yml`)

`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`

## Design docs

### SOPs
- `docs/sop/skill-constructor-pattern.md` — how to add a new skill
- `docs/sop/go-constructors-are-package-functions.md` — why `New()` on an interface doesn't work
- `docs/sop/git-commit-message-format.md` — commit message structure and style

### GWIs
- `docs/gwi/architecture-who-owns-what.md` — ownership table for every concern in the system
- `docs/gwi/architecture-trace-data-flow-first.md` — map producer→channel→consumer before designing
- `docs/gwi/channels-producer-owns.md` — the producer creates the channel, not main.go
