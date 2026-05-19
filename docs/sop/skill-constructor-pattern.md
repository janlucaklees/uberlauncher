# SOP: Adding a New Skill

## The rule

Every skill is a struct in its own package under `internal/skills/<name>/`. It exposes a package-level `New()` constructor and implements `Id() string` and `Init(ctx skill.Context)`.

## Steps

**1. Create the package and file**

```
internal/skills/myskill/myskill.go
```

**2. Define the struct and constructor**

```go
// internal/skills/myskill/myskill.go
package myskill

import "uberlauncher/internal/skill"

type MySkill struct {
    ctx skill.Context
}

func New() skill.Skill {
    return &MySkill{}
}

func (s *MySkill) Id() string { return "myskill" }

func (s *MySkill) Init(ctx skill.Context) {
    s.ctx = ctx
    // register entries, start background tasks
}
```

**3. Register in main.go**

```go
// cmd/uberlauncher/main.go
var skillList = []skill.Skill{
    // ... existing skills ...
    myskill.New(),
}
```

## Rules

- `Id()` must return a unique lowercase string — it is used to namespace the cache directory.
- Store the `skill.Context` on the struct if you need it after `Init` (e.g. inside `Run` closures).
- Background work goes inside `ctx.Runtime.Go(func() { ... })` so it is tracked by the WaitGroup.
- Surface errors via `ctx.Notifier.ReportError(err)`, never panic or log directly.
- Desktop notifications go via `ctx.Notifier.SendNotification(message)`.

## See also

- `docs/gwi/architecture-who-owns-what.md` — which context field is responsible for what
