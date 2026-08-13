---
name: kozmo-feature-implement
description: Use when the user asks to implement, build, or continue a feature that has a spec capsule. Implements strictly from the spec.
---

## What you are doing

You are implementing a feature strictly from its spec capsule under `.agents/specs/<spec-name>/`.

If no spec exists for the requested work, switch to the `kozmo-feature-plan` skill and create one first. Do not implement without a spec.

Your job:

1. Follow the standards in `AGENTS.md` at the repo root and `.agents/steering/project-standards.md`. They are the single source of truth for repo conventions, stack, and patterns.
2. Read the full capsule (`overview.md`, `design.md`, `tasks.md`) and any `reference/` materials before writing code.
3. Implement the spec, keeping `tasks.md` updated continuously.
4. Keep the capsule state accurate so the work is resumable by any agent at any point.

---

## Capsule selection

If `<spec-name>` is provided, use that folder.

If not provided, pick the most relevant spec by scanning `.agents/specs/*/tasks.md`:

- Prefer status `in_progress`, then `ready`
- If multiple match, pick the most recently updated
- If still unclear, ask the user and stop

---

## Reference materials

If the spec folder contains a `reference/` directory, read the relevant files before implementing anything that touches that area. `design.md` will point to the specific files. Treat `design.md` plus `reference/` as the contract.

---

## Task notation (mandatory)

- `[ ]` not started
- `[-]` in progress
- `[x]` done

---

## Task updates are continuous, not batched

After each task:

- Update `tasks.md` immediately: mark `[x]`, add a one-line note under the task if it helps future continuation.

If you cannot finish a task:

- Leave it `[-]`
- Add a short blocker note under it
- Ask the user the minimum question(s) needed to unblock

Never batch task updates until the end. The capsule must reflect reality at every step.

---

## Scope discipline

- Implement what the spec says. If you discover the spec is wrong or incomplete, stop and update `design.md` / `tasks.md` first, then implement: do not silently diverge.
- If the user introduces a project-level change (a new convention, a stack decision, a pattern that affects more than this feature), update `.agents/steering/` rather than burying it in the feature spec. Repo-wide truth lives in AGENTS.md and steering; feature-specific detail stays in the capsule.
- **Public-repo hygiene is absolute**: no client names, no internal project references, no private paths, no prices, anywhere, including comments and test fixtures.

---

## Exit

When all tasks are `[x]`, run `make check` (see `AGENTS.md`: gofmt, golangci-lint, go test, kbf lint dogfood, conformance, schema-freshness, boundaries), summarize what was done, and note anything discovered that should inform future work.
