---
type: spec-doc
---

# Action

An action is a verb an agent may execute against an entity: flagging a
record, adding a note, requesting a fresh sync, proposing a change to the
ontology itself. Every action declares a risk tier, so a runtime can decide
what an agent may do unattended and what needs a human to confirm first.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `kind` | yes | `action` | Discriminates this document from other primitives. |
| `name` | yes | kebab-case string | Unique within the namespace. |
| `on` | yes | entity name | The entity this action applies to. Must resolve to a declared entity, or fails `KBF009`. |
| `risk` | yes | `auto \| confirm` | `auto`: an agent may execute it unattended. `confirm`: a human confirms first. Empty or invalid fails `KBF010` / `KBF002`. |
| `writes` | yes | free text | What the action produces. In v0, always the findings layer or a proposal queue, never a mart directly. |

Action has no `tier` field. Its governance dimension is `risk`, not tier:
see the tier comparison table in `conventions.md`.

## Example

Copied from `packages/universal-core/ontology/actions.yaml`:

```yaml
kind: action
name: flag-for-review
on: order
risk: auto
writes: finding
```

## Common mistakes

- **Missing or invalid risk.** An empty `risk` fails `KBF010`; anything
  other than `auto` or `confirm` fails `KBF002`.
- **Adding a `tier` field.** Action does not have one. If a redesign needs
  action governance to track entity/metric's three-tier vocabulary instead
  of `risk`'s two, that is a spec change, not something to route around
  by adding an undocumented field.
- **Pointing `on` at something that is not an entity.** `on` names a
  declared entity, not a relation, metric, or attribute. A typo or a wrong
  kind of reference fails `KBF009`.
- **Treating a `confirm` action as safe to automate.** The risk tier is
  read by the runtime that executes actions (outside this repo's v0
  scope), not enforced by `kbf lint`. Getting it right in the ontology is
  what makes that enforcement possible later; it does nothing on its own
  in v0.
