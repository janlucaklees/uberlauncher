# GWI: Prefer a Rich Domain Model Over Anemic Data Bags

## The question to ask

> "Should I make one struct or two?"

If two types represent different states of the same concept, the answer is one.

## The pattern

An **Anemic Domain Model** splits a concept into multiple thin structs with no behavior:

```go
// smell: two types for one concept
type savedConnection struct { ssid string }
type scannedNetwork  struct { ssid string; signal int; secured bool }
```

Call sites then carry the burden of merging them, and every function that needs both must accept both.

A **Rich Domain Model** defines one struct per concept with semantic fields that capture all relevant state:

```go
type connection struct {
    ssid        string
    signal      int
    isSecured   bool
    isConnected bool
    isSaved     bool
}
```

Fields that are unknown at construction time are left at their zero value and populated later. A `connection` with `signal == 0` is a saved-but-out-of-range connection — the zero value carries meaning.

## Ubiquitous Language

Once the model is defined, its name becomes the only name used across all files, functions, and parameters. Mixing synonyms (`network`, `connection`, `saved`, `scanned`) for the same concept is a modelling problem, not a naming problem.

## The test

If you find yourself writing a function that accepts two structs describing the same thing and merges them into a third, stop. That merge belongs in the model, not the call site.

## See also

- `docs/sop/skill-constructor-pattern.md` — how skills own their context and data
