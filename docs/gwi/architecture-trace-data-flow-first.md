# GWI: Trace the Full Data Flow Before Designing Event Architecture

## The question to ask

> "I need component A to cause component B to update. How?"

Before writing any code, draw the arrow:

```
producer → channel/callback → consumer
```

If any part of that arrow is missing or assumed, that is the design gap — fill it explicitly before proceeding.

## Why

Skipping this step leads to incomplete event architectures where a channel is defined but nothing reads it, or a consumer is assumed to poll when in fact it needs a push. The gap is invisible until runtime.

## The test

For every event type in the system, answer all three questions:

1. **Who produces it?** (which function calls the send)
2. **Who consumes it?** (which function blocks on the receive)
3. **What is between them?** (the channel, and who owns it)

If you cannot answer all three without looking at code, go read the code first.

## Example: entry events

| Question | Answer |
|---|---|
| Who produces? | Skills via `ctx.Store.UpsertEntry()` |
| Who consumes? | The UI, to trigger a re-render |
| What is between them? | Nothing yet — deferred until a skill needs it |

Because the consumer and channel are not yet implemented, dynamic entry updates are intentionally absent. When the time comes, `store.Store` will own an `Events` channel and the UI will listen to it via a BubbleTea wait command — the same pattern used for `notifier.Events`.

## See also

- `docs/gwi/architecture-who-owns-what.md` — current ownership map
- `docs/gwi/channels-producer-owns.md` — who should own the channel
