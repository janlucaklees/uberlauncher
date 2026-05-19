# SOP: Constructors Are Package-Level Functions, Not Interface Methods

## The rule

Go interfaces describe instance methods only. Constructors are always plain package-level functions named `New()`, never methods on an interface.

## Bad

```go
// internal/skill/skill.go  ❌
type Skill interface {
    New() Skill   // requires an instance to call — defeats the purpose
    Init(ctx Context)
}
```

## Good

```go
// internal/skill/skill.go  ✓
type Skill interface {
    Id() string
    Init(ctx Context)
}
```

```go
// internal/skills/myskill/myskill.go  ✓
func New() skill.Skill {
    return &MySkill{}
}
```

## Why

You need an existing instance to call an interface method. A constructor that requires an instance to construct an instance is circular. In Go, the answer is always a package-level function. The compiler enforces the signature at the call site — if `myskill.New` has the wrong return type, it won't compile into a `[]skill.Skill` slice.

## The TypeScript analogy

TypeScript can express `new()` in an interface. Go cannot. Where TypeScript would use a static factory method or `new ClassName()`, Go uses `package.New()` — a named export with a fixed signature.
