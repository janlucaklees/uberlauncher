# GWI: Who Owns What in the Architecture

## The question to ask

> "Which component is responsible for X?"

## The split

| Concern | Owner | Notes |
|---|---|---|
| Background goroutine lifecycle | `runtime.Runtime` | `Go()` + `Wait()` |
| User-facing errors and messages → UI | `notifier.Notifier` | via `Events chan Event` |
| Desktop (system) notifications | `notifier.skillNotifier` | calls `send()` directly, scoped per skill |
| Entry storage and ranking | `store.Store` | thread-safe map + Engine |
| Per-skill file cache | `cache.SkillCache` | namespaced under `~/.cache/uberlauncher/<skillId>/` |
| Per-skill config | `config.Config` | stub; returns zero values |
| Skill initialization order | `cmd/uberlauncher/main.go` | synchronous loop before BubbleTea starts |
| UI event loop | `internal/ui/model.go` | BubbleTea model; listens to `notifier.Events` |

## The test

If you are unsure where a piece of logic belongs, ask:

- Does it move data from a skill to the UI? → notifier (errors/messages) or store (entries)
- Does it run in the background and needs tracking? → `ctx.Runtime.Go()`
- Does it read or write files specific to one skill? → `ctx.Cache`
- Does it need a config value? → `ctx.Config`

## See also

- `docs/gwi/architecture-trace-data-flow-first.md` — how to reason about data movement before designing
- `docs/sop/skill-constructor-pattern.md` — how to wire a new skill into this ownership model
