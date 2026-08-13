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

## Implementation clarifications (resolved during Batch 1/2, binding)

Gaps found while building the model/linter. Recorded here per "don't silently
diverge"; binds Batch 3/4 content authoring too.

- **`tier` vocabulary is kind-specific.** Entity/Metric `tier` uses the
  steering governance vocabulary (structural/glossary/instance), inherited
  silently since design.md examples don't restate steering. Relation `tier`
  is a distinct axis (source-synced/client-configured), explicitly annotated
  in its example because it deviates from the default. Action has no `tier`
  field; its governance-equivalent is `risk` (auto/confirm) per
  project-architecture's "risk tier" framing. Competency-question and slot
  mapping (operational elements) carry neither: KBF010 does not apply to them.
- **KBF010 scope**: fires on empty Entity.tier, Relation.tier, Metric.tier, or
  Action.risk. **KBF002 scope**: fires on any non-empty enum field whose value
  is outside its kind-specific allowed set (tier, cardinality, additivity,
  risk, and `kind` itself). Only fields design.md marks `# enum:`, plus tier/
  risk per above, are closed vocabularies; `resolution`, `formula`, `unit`,
  `attributes[].type` stay opaque strings in v0 (matches the documented
  "formula and resolution are opaque strings" intent, generalized).
- **KBF007 vocabulary source**: no separate declaration file. The controlled
  verb vocabulary is the *set of distinct Relation.name values already
  declared in the extends-root package* (universal-core, or self when linting
  universal-core itself). Adding a verb means adding a relation to
  universal-core (RFC), which then unlocks it for extending packages. No new
  primitive, no `kind: verb-vocabulary`.
- **KBF008 fork detection**: an extending package redeclaring an element with
  a name+kind that exists in its extends-root is a fork UNLESS every
  non-glossary field on the child's copy is zero-valued, i.e. the child only
  layers glossary-tier fields. Glossary-eligible fields in v0: Entity.synonyms,
  Metric.thresholds (matches "thresholds: glossary tier by definition").
  Relation and Action have no glossary carve-out: any redeclaration forks.
- **`extends` resolution**: `kbf lint` takes one or more package paths as
  positional args and loads all of them into one in-memory set keyed by
  manifest `name`. Each package's `extends` is resolved by name against that
  set (not by filesystem convention: kbf lints arbitrary directories, not just
  this repo's `packages/`). A package whose `extends` isn't among the supplied
  paths fails KBF011 with a fix hint to pass the parent's path too. Linting a
  single root package (`extends: null`) needs only its own path.
- **KBF009 vs KBF006 vs KBF012 split**: KBF006 is relation.from/to only.
  KBF012 is attribute slot references against `install/slots.yaml` only.
  KBF009 (generic dangling cross-reference) covers everything else that names
  another element: metric.grain entries, action.on, competency-question.expects.
  Manifest.extends failing to resolve is KBF011 (manifest invalid), not KBF009.
- **Schema file scope**: exactly two files as tasked. `SlotMapping` (the
  `install/slots.yaml` row shape) is reflected into `manifest.schema.yaml`'s
  `$defs` (installation-config family) but is not itself a file root schema in
  v0, so `install/slots.yaml` doesn't get direct yaml-language-server root
  validation yet. Structural validation of slots still happens via the Go
  struct in the loader; this is a known, deliberate v0 gap, not a silent one.
- **Relation identity for KBF003/KBF008 is (name, from, to), not bare name**
  (found authoring content; binds Batch 2 as much as the entries above).
  Relation.name holds a controlled-vocabulary *verb*, meant to recur across
  unrelated entity pairs (that's the whole point of a 10-20 word vocabulary
  covering dozens of relations: project-standards.md, and KBF007's own
  wording, "the *set of distinct* Relation.name values", presupposes
  repeats feeding that set). So (1) KBF003 duplicate-name-in-namespace, for
  `kind: relation` only, keys on the (name, from, to) triple, not name
  alone: universal-core legitimately declares `contains` and `supplies` and
  `located-at` twice each, over different pairs. (2) KBF008 fork-matching
  for `kind: relation` also keys on (name, from, to): a child package
  declaring a relation whose triple doesn't already exist in the parent is
  a new relation, full stop, not a fork candidate at all, even when the verb
  is already used elsewhere in the parent under a different pair. This has
  to be true or cafe-demo's own required extension (a new client-configured
  `belongs-to` between two locations, reusing a verb the parent already
  spends on location→organization) would be structurally impossible to
  express: no package could ever add a relation without first winning an
  RFC for a brand-new, never-before-used verb. "Relation... any
  redeclaration forks" (above) still holds for the case this note carves
  out from it: same (name, from, to) triple as the parent, re-stated by a
  child, with or without field changes, is always a fork, no glossary
  carve-out. Entity/Metric/Action/competency-question identity stays plain
  name (already globally unique by construction), unaffected by this note.
