# Repository Guidelines

## Project Structure & Architecture
`uberlauncher` is a Go Bubble Tea launcher. The entry point is `cmd/uberlauncher/main.go`, which wires runtime services, registers built-in skills, initializes them synchronously, and starts the TUI.

Current internal layout:
- `internal/cache/`: per-skill filesystem cache helpers.
- `internal/config/`: per-skill config access.
- `internal/engines/`: ranking engine interface and fuzzy implementation.
- `internal/entry/`: launcher entry type and execution context.
- `internal/notifier/`: background-task messages/errors delivered to the UI.
- `internal/runtime/`: goroutine lifecycle helpers and command detection.
- `internal/skill/`: core `Skill` and `Context` interfaces.
- `internal/store/`: in-memory entry registry and query/ranking bridge.
- `internal/ui/`: Bubble Tea model and rendering.
- `internal/skills/`: built-in skills: `apps`, `search`, `system`, `todo`, `wifi`.

Initialization model:
- Skills are constructed in `main.go` and initialized synchronously via `Init(ctx skill.Context)`.
- Skills register launcher entries with `ctx.Store.UpsertEntry(...)`.
- Long-running or asynchronous work should go through `ctx.Runtime.Go(...)` and report user-visible status via `ctx.Notifier`.

## Build, Test, and Development Commands
Prefer `make` targets:
- `make build`: build `./uberlauncher` from `cmd/uberlauncher`.
- `make run`: run with `go run`.
- `make test`: run `go test ./...`.
- `make fmt`: format all Go files with `gofmt`.
- `make fmt-check`: verify Go formatting.
- `make vet`: run `go vet ./...`.
- `make tidy`: run `go mod tidy`.
- `make check`: run `fmt-check`, `vet`, `test`, and `lint`.

Current repo caveat:
- `make lint` is broken in the current `Makefile`; it invokes `golangci-lintrun` instead of `golangci-lint run`.
- As of May 20, 2026, `make vet` and `make test` pass, and `make check` fails only at that lint step.

## Coding Style & Naming Conventions
Follow `.editorconfig`: UTF-8, LF, trailing newline, no trailing whitespace. Go code uses tabs. Keep packages small and focused. Use exported `CamelCase` for public identifiers and unexported `camelCase` internally. Prefer constructor functions like `New()` at package scope, matching the existing codebase.

When adding skills:
- Create `internal/skills/<name>/<name>.go`.
- Implement `Id() string` and `Init(ctx skill.Context)`.
- Expose `func New() skill.Skill`.
- Register the skill in `cmd/uberlauncher/main.go`.

## Testing Guidelines
Use the standard `testing` package and keep tests adjacent to code as `*_test.go`. Focus first on behavior-heavy packages such as `internal/engines/`, `internal/store/`, runtime/notifier flows, and skill adapters.

Current state:
- The repository currently has no Go test files.
- `make test` succeeds because all packages report `[no test files]`.

If automated coverage is not practical, document manual verification steps in the PR.

## Git Hooks & Validation
`lefthook.yml` defines:
- `pre-commit`: `make fmt-check`, `make vet`, and `make lint` when `golangci-lint` is installed.
- `pre-push`: `make test`.

Keep in mind the current `make lint` typo affects the hook path too when `golangci-lint` is present.

## Commit & Pull Request Guidelines
Before creating a commit, read and follow `docs/sop/git-commit-message-format.md`. That SOP is the source of truth for title tense, summary paragraphs, reasoning paragraphs, and trailers.

In PRs, include:
- What changed and why.
- How you validated it.
- Screenshots or terminal captures for TUI changes.
- Linked issue/task when relevant.

## Security & Configuration Tips
Never commit secrets. The `todo` skill uses `TODOIST_API_TOKEN`, either from the environment or `~/.config/uberlauncher/todo.env`. Skill caches live under `~/.cache/uberlauncher/`.

## Design docs

The SOPs and GWIs below are active project instructions, not optional reference material. Before doing work covered by one of these documents, open the relevant file and follow it. In particular, check the commit SOP before every commit, the skill constructor SOP before adding or changing skills, and the ownership/data-flow GWIs before moving responsibilities between packages.

### SOPs
- `docs/sop/skill-constructor-pattern.md` — how to add a new skill
- `docs/sop/go-constructors-are-package-functions.md` — why `New()` on an interface doesn't work
- `docs/sop/git-commit-message-format.md` — commit message structure and style

### GWIs
- `docs/gwi/architecture-who-owns-what.md` — ownership table for every concern in the system
- `docs/gwi/architecture-trace-data-flow-first.md` — map producer→channel→consumer before designing
- `docs/gwi/channels-producer-owns.md` — the producer creates the channel, not main.go
