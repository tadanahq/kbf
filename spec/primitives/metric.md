---
type: spec-doc
---

# Metric

A metric is a number computed over entities, with enough declared shape
(grain, additivity, unit) that it can be rolled up, joined, or compared
without an agent guessing whether that is safe.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `kind` | yes | `metric` | Discriminates this document from other primitives. |
| `name` | yes | kebab-case string | Unique within the namespace. |
| `formula` | yes (root only) | free text | Opaque expression string in v0: the linter checks that it is present, not that it parses. See `spec/versioning.md`. |
| `grain` | yes (root only) | list of entity names and/or dimensions | The level at which one row of this metric is true. `business-date` is the standard time dimension. |
| `additivity` | yes (root only) | `additive \| semi-additive \| non-additive` | Whether the metric can be summed across its grain. Empty or invalid fails `KBF005` / `KBF002`. |
| `unit` | yes (root only) | free text | Opaque string in v0 (common values: `currency`, `count`, `ratio`), not a closed vocabulary. |
| `thresholds` | no | map of named bounds | Glossary tier by definition: the one field a composing playbook may set without forking the metric. |
| `tier` | yes (root only) | `structural \| glossary \| instance` | Governance tier for the metric definition itself (separate from a threshold's own glossary status). Empty fails `KBF010`. |

"Root only" fields are required when a metric is defined for the first
time. `thresholds` is the exception: it is the one field a composing
playbook may set on its own, in a fragment that leaves everything else
zero-valued (see "Common mistakes").

## Example

Copied from `playbooks/core-business/ontology/metrics.yaml`:

```yaml
kind: metric
name: gross-margin
formula: (revenue - purchase-cost) / revenue
grain: [organization, business-date]
additivity: non-additive
unit: ratio
thresholds: {warn-below: 0.55}
tier: structural
```

`core-business`'s own metrics stop at `[organization, business-date]`
grain: `location` is not universal (`playbooks/core-operations` adds it).
A core playbook that wants a location-grain sibling of an existing metric
declares a new metric with its own name at that grain
(`playbooks/core-operations`'s `average-ticket`, next to core-business's
`average-transaction-value`): grain is not glossary-eligible, so
redeclaring an existing metric name at a different grain is a fork, not
an override.

Metric formulas may reference other metric names directly (`revenue`,
`purchase-cost` above); this is a documentation convention, not something
`kbf lint` parses in v0.

## Common mistakes

- **Missing grain or additivity.** A metric without both fails `KBF005`.
  These are the two fields that make a metric safe to roll up or join; a
  metric is not "done" without them even though `formula` alone might look
  complete.
- **Treating `unit` as a closed enum.** It is an opaque string in v0. `kbf
  lint` will not reject an unusual value, but an unfamiliar unit makes the
  metric harder for the next agent to trust. Stay inside the common set
  unless there is a real reason not to.
- **Forking instead of overriding a threshold.** A playbook that wants to
  tighten or loosen a threshold should declare a fragment with only `kind`,
  `name`, and `thresholds` set, exactly like the entity synonym pattern in
  `spec/primitives/entity.md`:

  ```yaml
  # Right: repeats nothing but the glossary-eligible field. See
  # examples/cafe-demo/ontology/metrics.yaml for the real file.
  kind: metric
  name: location-labor-cost-ratio
  thresholds: {warn-above: 0.32}
  ```

  Repeating `formula`, `grain`, `additivity`, `unit`, or `tier`, even with
  identical values, is a fork and fails `KBF008`.
