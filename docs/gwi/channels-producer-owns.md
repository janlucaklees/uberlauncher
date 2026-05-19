# GWI: The Producer Owns the Channel

## The rule

The component that sends events owns the channel. Consumers receive a reference to it (or a read-only view `<-chan`). Channels are never created in `main.go` and passed around.

## Bad

```go
// cmd/uberlauncher/main.go  ❌
events := make(chan Event, 64)
notifier := notifier.New(events)  // notifier doesn't own it
ui := ui.New(store, events)       // UI holds same shared channel
```

## Good

```go
// internal/notifier/notifier.go  ✓
type Notifier struct {
    Events chan Event   // notifier creates and owns this
}

func New() *Notifier {
    return &Notifier{Events: make(chan Event, 64)}
}
```

```go
// cmd/uberlauncher/main.go  ✓
n := notifier.New()
ui := ui.New(store, n)  // UI reads n.Events, notifier owns it
```

## Why

Ownership determines lifecycle. If `main.go` creates the channel, it becomes a coordination hub responsible for wiring producers to consumers — coupling unrelated components through a shared variable. When the producer owns the channel, adding a new consumer is a one-liner and producers remain self-contained.

## The test

Ask: "If I remove this component, does the channel go away too?" If yes, the channel belongs to that component.

## See also

- `docs/gwi/architecture-trace-data-flow-first.md`
