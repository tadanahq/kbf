---
type: spec-design
---
# kbf-v0-foundation – Design

Architecture authority: `.agents/steering/project-architecture.md` (layout,
meta-model table, tooling packages). This design adds the feature-level detail;
it does not restate steering.

## The meta-model (authoring shapes)

One YAML document per element, one file per entity plus its owned elements, or
grouped files per kind: the linter accepts both; `kind:` discriminates.

```yaml
kind: entity
name: dish            # kebab-case, unique within namespace
meaning: >-           # one sentence, required
  A sellable menu item.
identity: [dish_id]
resolution: identity-mapping   # free text in v0, named strategies later
tier: structural
synonyms: {en: [item], es: [plato]}
attributes:
  - {name: category, type: text, slot: pos.catalog}
states: [active, retired]      # optional
```

```yaml
kind: relation
name: contains
from: dish
to: ingredient
cardinality: many-to-many      # enum: one-to-one|one-to-many|many-to-many
join: [dish_id, ingredient_id]
tier: source-synced            # enum: source-synced|client-configured
temporal: false
```

```yaml
kind: metric
name: dish-margin
formula: (revenue - ingredient-cost) / revenue   # v0: opaque expression string
grain: [dish, location, business-date]
additivity: non-additive       # enum: additive|semi-additive|non-additive
unit: ratio
thresholds: {warn-below: 0.60} # glossary tier by definition
tier: structural
```

```yaml
kind: action
name: flag-for-review
on: dish
risk: auto                     # enum: auto|confirm
writes: finding
```

```yaml
kind: competency-question
question: Which dishes lost margin this month, per location?
expects: [dish-margin]         # elements the answer must use
```

Slot mappings live in `install/slots.yaml`: `{slot: pos.catalog, source: ""}`
rows; static coverage = share of declared slots with a non-empty source.
`manifest.yaml`: `{name, version, spec: v0, extends: universal-core}`
(universal-core itself: `extends: null`).

v0 keeps `formula` and `resolution` as opaque strings deliberately: parsing
expressions is gate-era work; the linter checks presence and cross-references
(`expects`, `grain`, `slot` targets must exist), not expression semantics.

## Lint rule set (rule ids are public API)

`KBF001` unknown field · `KBF002` bad enum value · `KBF003` duplicate name in
namespace · `KBF004` entity without identity · `KBF005` metric without
grain/additivity · `KBF006` relation endpoint not declared · `KBF007` verb
outside controlled vocabulary · `KBF008` fork of a core element (redefinition
in an extending package) · `KBF009` dangling cross-reference · `KBF010`
missing governance tier · `KBF011` manifest missing/invalid · `KBF012` slot
reference without declaration. Human render: lipgloss table grouped by file.
`--format json`: `{rules: [{id, file, line, element, message, fix}]}`.

## Controlled verb vocabulary (v0 seed, lives in universal-core)

`contains, belongs-to, located-at, works-at, supplies, sells, places, staffed-by,
billed-to, derived-from, supersedes, responsible-for`. Additions via RFC.

## universal-core content (v0)

Entities: organization, location, order, product, customer, employee, shift,
supplier-purchase (split as supplier + purchase, 9 files ok if cleaner).
Metrics: revenue, order-count, average-ticket, labor-cost-ratio,
purchase-cost, gross-margin. Actions: flag-for-review, annotate,
request-sync, propose-extension. Competency questions: 1+ per entity.

## cafe-demo content

Fictional single-location cafe ("Demo Cafe"): extends universal-core with a
`menu-item` glossary rename (synonym, not fork), one client-configured relation
(location grouping), one threshold override, one vertical metric
(waste-ratio), slots mapped to fictional sources (`demopos`, `demobooks`).

## Key flows

- `kbf schema`: model structs → JSON Schema → YAML files in `schema/`;
  `--check` mode diffs against committed files (CI freshness job).
- `kbf lint <path...>`: load manifest → resolve extends chain (v0: single
  parent) → parse files position-aware → structural validation (against model)
  → semantic rules → render. Exit 1 on any rule hit.
- `kbf coverage <path>`: table of slots by entity: declared / mapped / unmapped.
- `kbf compile --to mermaid <path>`: entities as nodes, relations as labeled
  edges, actions as annotations. Deterministic output (sorted), golden-tested.
- Boundaries: `scripts/boundaries.go` walks the tree, fails on blocklist terms;
  blocklist named in-file; self-test plants a violation in a temp tree.

## Edge cases & constraints

- Lint must not panic on arbitrary YAML: fuzz the loader with the invalid
  fixtures; unknown `kind` is KBF002 not a crash.
- Windows paths and CRLF tolerated; output paths always forward-slash.
- Everything deterministic: sorted iteration everywhere (maps never range
  unsorted into output).

## Decisions

- Cross-file resolution happens per package after full load; no incremental
  mode in v0.
- `extends` chain depth 1 in v0 (universal-core only); deeper chains are a
  spec-versioning question, not a linter feature.
- Mermaid over any richer render: it previews natively in editors and GitHub.
