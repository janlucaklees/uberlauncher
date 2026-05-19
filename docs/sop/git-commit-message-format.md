# SOP: Git commit message format

## Structure

```
<title>

<summary>

<paragraph per group of changes>
```

Each section is separated by a blank line.

## Title

Short and precise. Past tense. States what was achieved at the level of intent — not what files were touched.

- 72 characters or fewer
- No period at the end

```
# Good
Replaced string role field with proper RBAC system
Removed ticket ID from ticket overview
Fixed ticket query returning closed tickets for MedOps users

# Bad
Updated user.py                  ← describes a file, not a change
Refactoring and some bug fixes   ← vague and combines unrelated things
```

## Summary

Always present. One or two sentences giving a high-level description of what was achieved. Its purpose is to let a reader skimming git history judge in seconds whether this commit is relevant to what they are looking for — without having to open the diff.

## Paragraphs

One paragraph per logical group of changes. Each paragraph explains what that group of changes did and **why** it was done. The diff already shows the mechanics; the paragraph captures the reasoning that would otherwise be lost.

Grouping is by shared reasoning, not by layer or file. If two sets of changes serve different purposes but happen to be in the same commit, they belong in separate paragraphs.

## Example

```
Replaced string role field with proper RBAC system

Replaced the flat role text column on User with a proper Role model,
Permission enum, and UserRole join table. All access control checks now
use User.has_permission() instead of comparing role name strings.

Introduced the Role and UserRole models and a Permission enum with
granular per-resource permissions. Added a role management blueprint
with full CRUD so roles and their permissions can be managed through
the UI rather than being hardcoded.

Moved data access queries out of model classmethods and into the
repository layer (study, ticket, role, user repositories), following
the project's layering conventions. Took the opportunity to also move
db.session.commit() out of the password reset service and into the
calling controllers.
```

## Trailers

Add a `Co-Authored-By` trailer when a commit was written with AI assistance.

The trailer should identify the actual AI tool/model that assisted with the commit, not a fixed placeholder example.

Format:
```text
Co-Authored-By: <AI tool name> <model name> <noreply@provider-domain>
```
