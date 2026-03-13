Below is a **self-contained technical specification** for implementing the **UberLauncher** project.
It is written so that an **AI implementation agent** can execute it without needing prior discussion context.

---

# UberLauncher — Technical Specification (MVP)

## 1. Project Overview

UberLauncher is a **terminal-based launcher** that allows users to quickly execute actions using **fuzzy search over aggregated entries provided by modular skills**.

It behaves like a hybrid of:

* a **fuzzy command launcher**
* a **CLI with extensible commands**

The system aggregates entries from multiple **skills** and presents them in a **fuzzy searchable list**.
Users select entries or provide free-text input to trigger actions.

---

# 2. Design Goals

The system must satisfy the following goals:

### Usability

* Keyboard-driven interface
* Minimal latency
* Fuzzy search over all available entries
* Immediate execution via `Enter`

### Extensibility

* New skills must be easy to implement
* Skills must be isolated in their own packages
* Skills communicate with the core through a defined interface

### Installation Simplicity

* The system should compile into a **single binary**
* No required setup or background daemon

### Architecture Simplicity

* Skills are compiled into the binary
* No external plugin system in the MVP
* Communication between core and skills uses **DTOs**

---

# 3. Technology Stack

Language:

```
Go
```

UI framework:

```
Bubble Tea (TUI)
```

Other libraries allowed:

* fuzzy matching library similar to `fzf`
* standard Go libraries

---

# 5. Core Concepts

## 5.1 Skills

Skills provide functionality.

Each skill:

* publishes entries
* executes commands
* may update entries asynchronously
* may run background tasks

Skills live in:

```
internal/skills/<skill_name>/
```

Each skill must implement the **Skill interface**.

---

# 5.2 Entries

Entries represent **selectable items shown in the launcher**.

Examples:

```
shutdown
restart
wifi on
wifi HomeNetwork
todo
```

Entries are searchable and executable.

---

# 5.3 Free-Text Skills

Some skills accept raw user input.

Example:

```
todo buy milk tomorrow
```

Only the **first word** determines the skill.

Free-text mode activates when:

* first token exactly matches the skill name
* followed by a space
* match is **case-sensitive**
* no leading whitespace allowed

Examples:

Valid:

```
todo buy milk
todo something
```

Invalid:

```
Todo buy milk
 todo buy milk
some text todo
todo
```

---

# 6. Data Transfer Objects

DTOs are used for communication between core and skills.

---

# 6.1 EntryDTO

Represents a selectable entry.

```
type EntryDTO struct {
    SkillName string
    EntryID   string
    DisplayText string
    IsFreeText bool
}
```

Notes:

* `DisplayText` must include the skill prefix
* `EntryID` must be unique **within the skill**
* Global identity is `(SkillName, EntryID)`

---

# 6.2 RunCommandDTO

Represents an execution request.

```
type RunCommandDTO struct {
    SkillName   string
    EntryID     string
    RawInput    string
    TriggerType TriggerType
}
```

```
type TriggerType int

const (
    TriggerEntry TriggerType = iota
    TriggerRawInput
)
```

Rules:

* `TriggerEntry` → execution via selected entry
* `TriggerRawInput` → raw user input execution

---

# 7. Skill Interface

All skills must implement the following interface.

```
type Skill interface {

    Manifest() SkillManifest

    Start(runtime SkillRuntime) error

    Execute(cmd RunCommandDTO) error

    Stop(ctx context.Context) error
}
```

---

# 7.1 SkillManifest

```
type SkillManifest struct {
    Name string
    SupportsFreeText bool
}
```

---

# 8. Skill Runtime Services

Skills receive a runtime interface providing core services.

```
type SkillRuntime interface {

    PublishEntries(entries []EntryDTO)

    UpsertEntries(entries []EntryDTO)

    RemoveEntries(ids []string)

    Notify(message string)

    ReportError(err error)

    CacheDir() string

}
```

Responsibilities:

| Method         | Purpose                |
| -------------- | ---------------------- |
| PublishEntries | initial entry list     |
| UpsertEntries  | add/update entries     |
| RemoveEntries  | delete entries         |
| Notify         | send user notification |
| ReportError    | display inline error   |
| CacheDir       | skill cache directory  |

---

# 9. Entry Store

The core maintains an entry store.

Key:

```
(skillName, entryID)
```

Behavior:

* duplicate `(skill,id)` replaces the entry
* duplicate IDs inside a skill cause error

---

# 10. Search and Ranking

Search is performed using **fuzzy matching**.

Ranking uses:

```
score = fuzzy_score
      + usage_weight
      + recency_weight
```

Tracked per:

```
(skillName, entryID)
```

Free-text usage only tracks the skill entry.

---

# 11. UI Behavior

The UI uses **Bubble Tea**.

Components:

```
Input Field
Entry List
Selection State
Error Display
```

---

# 12. Default Behavior

When launcher opens:

```
input = ""
all entries shown
best match selected
```

---

# 13. Typing Behavior

On each keystroke:

* fuzzy filter updates
* ranking recalculated
* best match auto-selected

Unless user manually changed selection.

---

# 14. Free-Text Mode

Activates when:

```
<skillname> + space
```

Behavior:

```
selection moves to input field
Enter submits raw input
```

Navigation:

```
Up/Down → move into result list
Enter → execute selected entry
User must manually return to input
```

---

# 15. Async Entry Updates

Skills may update entries while the launcher is running.

Example:

WiFi scanning.

When updates occur:

Rules:

```
if user has not manually selected entry:
    auto-select best match

if user manually selected entry:
    keep selected entry if still exists

if selected entry disappeared:
    fallback to best match
```

---

# 16. Execution Behavior

When `Enter` is pressed:

### Case 1 — Entry selected

```
TriggerEntry
EntryID provided
```

### Case 2 — Free-text mode active

```
TriggerRawInput
RawInput provided
```

---

# 17. Error Handling

Errors must be visible to the user.

Mechanisms:

```
runtime.ReportError()
runtime.Notify()
```

Inline errors should remain visible until user dismisses them.

---

# 18. Notifications

Skills may emit notifications for:

* async success
* async failure

Example:

```
todo buy milk
```

Execution:

```
launcher closes
API request executes
notification appears
```

---

# 19. Caching

Each skill receives a cache directory:

```
~/.cache/uberlauncher/<skill>/
```

Skills manage cache data themselves.

---

# 20. Background Tasks

Background tasks must run as **detached subcommands**.

Example:

```
uberlauncher __internal refresh-app-cache
```

Core may spawn detached tasks when needed.

---

# 21. Suggested Project Layout

```
cmd/
  uberlauncher/

internal/
  core/
  ui/
  ranking/
  runtime/
  store/

  skills/
    apps/
    todo/
    wifi/
    system/

  jobs/
    refresh_apps_cache/
```

---

# 22. MVP Skill Set

Implement the following initial skills.

### apps

Lists installed desktop applications.

### todo

Free-text skill adding todos via API or file.

### shutdown

Shuts down the system immediately

### restart

Restarts the system immediately

### wifi (optional)

Provides:

```
wifi on
wifi off
wifi <network> (to connect / disconnect)
```

---

# 23. Implementation Phases

## Phase 1

Core foundations:

* skill registry
* entry store
* ranking engine
* basic UI
* command execution

---

## Phase 2

Free-text mode:

* detection
* navigation
* raw input execution

---

## Phase 3

Async entry updates:

* runtime entry API
* dynamic updates
* selection preservation

---

## Phase 4

Cache support:

* cache directory
* internal background jobs

---

## Phase 5

Initial skill implementations.

---

## Phase 6

Hardening:

* duplicate detection
* structured logging
* configuration file

---

# 24. Out of Scope for MVP

The following features must NOT be implemented yet:

* external plugins
* alias system
* hidden search keywords
* daemon mode
* complex configuration UI
* massive indexing

---

# End of Specification
