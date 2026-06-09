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

## Go Conventions

### Service Object Pattern
Skills store `skill.Context` on the struct in `Init` and implement all behavior as methods on the receiver. Never thread context through as a function parameter after initialization.

```go
type MySkill struct {
    ctx skill.Context
}

func (s *MySkill) Init(ctx skill.Context) {
    s.ctx = ctx
    // all further logic uses s.ctx
}
```

### Rich Domain Model
Define one struct per domain concept with semantic boolean fields. Do not split a concept across multiple thin data-transfer types and merge them at call sites. A unified type with meaningful fields (`isConnected`, `isSaved`, `isSecured`) is always preferred over separate `savedX` and `scannedX` structs.

### Ubiquitous Language
Pick one name per concept and use it consistently across all files, types, methods, and parameters. Mixing synonyms (`network`, `connection`, `saved`, `scanned`) for the same concept signals a modelling problem, not a naming problem.

### Eliminate Accidental Complexity
Do not build infrastructure for features that cannot work end-to-end. If a feature is incomplete (e.g. connecting to unsaved secured networks requires a password prompt that does not exist), remove the scaffolding entirely rather than leaving dead code.

### Single Level of Abstraction
Orchestrating functions read like a table of contents — they name steps and delegate everything. Implementation functions do one concrete thing and know nothing about the bigger picture. Never mix the two in the same function.

## Performance & Snappiness

Snappiness is a core requirement of this launcher — not a nice-to-have. The launcher is a keystroke-triggered tool and must feel instant.

- **Known data is shown immediately.** Never gate fast data behind slow data. If saved connections are known, upsert them before waiting for a network scan.
- **Slow operations run in the background.** Subprocess calls, network I/O, and disk reads belong in `ctx.Runtime.Go()` goroutines, not in the synchronous part of `Init()`.
- **Upsert as soon as data is available.** Do not batch or defer — call `ctx.Store.UpsertEntry()` the moment each piece of data is ready.
- **Actions should trigger immediate UI updates.** When a user-initiated action changes known state (e.g. connecting to a network), re-upsert the affected entries right away rather than waiting for the next poll cycle.

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
    Notifier Notifier  // UI feedback: Debug() + ReportError() + ReportWarning() + ReportMessage() + SendNotification()
    Config   ConfigMap // map[string]any — skill's section from config.toml
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
| `bluetooth` | `bluetooth` | Power on/off + connect to paired devices via `bluetoothctl` |
| `custom` | `custom` | User-defined shell commands from config; each entry has a `label` and `command` |
| `keyboard` | `keyboard` | Switch keyboard layout via `hyprctl switchxkblayout`; reads configured layouts from `hyprctl getoption input:kb_layout` |
| `notifications` | `notifications` | Enable/disable notifications via `makoctl` (do-not-disturb mode) |
| `power` | `power` | Set performance profile (power save / balanced / performance) via `powerprofilesctl` |
| `debug` | `debug` | Test entries for development; only loaded when `-v`/`--verbose` flag is set |
| `search` | `search` | Free-text; opens a search via `xdg-open` |
| `system` | `system` | Shutdown / reboot via `systemctl` or `shutdown` |
| `todoist` | `todoist` | Todoist quick-add via API; reads token from `ctx.Config["token"]` |
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
- `docs/gwi/domain-model-prefer-rich.md` — one semantic struct per concept; no anemic data bags
