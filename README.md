# UberLauncher

UberLauncher is a Bubble Tea based terminal launcher with modular skills.

It aggregates entries from built-in skills, applies fuzzy search with usage/recency ranking, and executes the selected result. It also supports free-text skills (for example `todo buy milk`).

## MVP Features

- Fuzzy search across all skill entries
- Usage + recency weighted ranking persisted in cache
- Free-text mode for supported skills
- Async skill entry updates via runtime events
- Built-in skills:
  - `apps`
  - `todo` (Todoist quick add)
  - `shutdown`
  - `restart`
  - `wifi`

## Requirements

- Go 1.22+
- Linux environment
- Optional runtime tools depending on skill usage:
  - `hyprctl` for app launching on Hyprland
  - `nmcli` for WiFi skill

## Build

```bash
go build ./cmd/uberlauncher
```
This creates the `uberlauncher` binary in the repository root.

## Run

```bash
./uberlauncher
```

## Todo Skill Setup

`todo` uses Todoist quick add and requires:

- `TODOIST_API_TOKEN` environment variable

Example:

```bash
export TODOIST_API_TOKEN="<your-token>"
./uberlauncher
```

Optional fallback file supported by the skill:

- `~/.config/uberlauncher/todo.env`

with content like:

```bash
TODOIST_API_TOKEN=<your-token>
```

## Internal Job

The apps skill can refresh its app cache using:

```bash
./uberlauncher __internal refresh-app-cache
```

## Cache Paths

- Usage ranking: `~/.cache/uberlauncher/usage.json`
- Skill caches: `~/.cache/uberlauncher/<skill>/`

## Project Structure

- `cmd/uberlauncher/` - entrypoint
- `internal/core/` - launcher orchestration and skill registry
- `internal/skill/` - skill interfaces
- `internal/types/` - shared DTOs
- `internal/ui/` - Bubble Tea UI model
- `internal/ranking/` - fuzzy ranking and usage store
- `internal/runtime/` - runtime services + event bus
- `internal/store/` - in-memory entry store
- `internal/skills/` - built-in skills
- `internal/jobs/` - internal background jobs

## Notes

- No fallback shell command execution is enabled in MVP.
- Aliases, external plugins, and daemon mode are out of scope for MVP.

## Developer Tooling

- `make help` shows available tasks
- `make fmt` formats Go files
- `make fmt-check` validates formatting
- `make vet` runs static vet checks
- `make test` runs all tests
- `make lint` runs `golangci-lint` (requires local install)
- `make check` runs `fmt-check`, `vet`, `test`, and `lint`

CI is configured at `.github/workflows/ci.yml` and runs format, vet, test, build, and lint on pushes/PRs.

## Git Hooks (Lefthook)

This repo uses Lefthook for Git hooks:

- `pre-commit`: runs `make fmt-check`, `make vet`, and `make lint` (when `golangci-lint` is installed)
- `pre-push`: runs `make test`
