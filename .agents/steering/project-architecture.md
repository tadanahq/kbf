# kozmo-bf – Project Architecture

## Repository layout

```
spec/            # the prose spec: index, primitives/, package-format, conventions, versioning
schema/          # GENERATED meta-schemas (ontology.schema.yaml, manifest.schema.yaml)
packages/        # ontology content: universal-core (Playbook Zero); verticals later
examples/        # cafe-demo: fictional, fully valid, teaching-first
tools/           # Go module: the kbf CLI and libraries
conformance/     # language-agnostic fixture suite (YAML in, expected outcome out)
rfcs/            # public change process for the spec
scripts/         # repo gates (boundaries scan), invoked by make
```

## The seven primitives (meta-model)

| Primitive | Carries |
|---|---|
| Entity | meaning, identity keys + resolution rule, attributes (typed, each with a source slot), synonyms, states where real |
| Relation | verb (controlled vocabulary), from/to, cardinality, join keys, tier, temporal-validity flag |
| Metric | formula, grain, additivity, unit, canonical filters, thresholds (glossary-editable) |
| Action | per-entity verbs agents may execute, risk tier (auto/confirm), what it writes |
| Slot mapping | element → source system declaration; fill status (static in v0) |
| Competency question | question + answer oracle; the acceptance suite |
| Namespace | `universal-core` vs `<package>`; extension rules |

Grouped in prose as **semantic elements** (entity, relation, metric),
**action elements** (action), and **operational elements** (slot mapping,
competency question, namespace).

## Tooling architecture (v0)

- `tools/` is one Go module, package layout:
  - `internal/model/`: the canonical structs (the meta-model). Single source of truth.
  - `internal/schemagen/`: emits `schema/*.yaml` from the model.
  - `internal/lint/`: loads YAML (position-aware), validates structure against
    the model, then applies semantic rules (extension-not-fork, verb vocabulary,
    grain/additivity presence, tier presence, cross-references resolve).
  - `internal/coverage/`: static slot-mapping completeness report.
  - `internal/compile/`: emitters; v0 ships `mermaid` only.
  - `cmd/kbf/`: cobra wiring: `lint`, `coverage`, `compile`, `schema` (regen).
- Lint pipeline: parse → structural validation → semantic rules → render
  (lipgloss table or `--format json` for agent consumption). **JSON output is a
  stable interface**: authoring agents parse it.
- Conformance runner is a thin Go test that walks `conformance/`, but fixtures
  never depend on Go: each case is `input/` + `expect.yaml` (ok | errors with
  rule ids).

## Package anatomy (what the linter understands)

```
<package>/
  manifest.yaml        # name, version, spec version, extends
  ontology/            # *.yaml: entities, relations, metrics, actions
  evals/               # competency questions
  install/             # slot-mapping templates, defaults
```

Workflows, agents, surfaces folders are reserved in the package format spec but
not linted in v0 (config-phase scope).

## Boundaries

- `spec/`, `schema/`, `packages/`, `examples/` never reference engines,
  databases, vendors, clients, or tooling internals.
- `tools/` may not embed package content; packages are always inputs.
- Public hygiene scan covers every file in the repo.
