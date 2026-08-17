# kozmo-bf – Project Architecture

## Repository layout

```
spec/            # the prose spec: index, primitives/, playbook-format, conventions, versioning
schema/          # GENERATED meta-schemas (ontology.schema.yaml, manifest.schema.yaml)
playbooks/       # ontology content: core-business (Playbook Zero); verticals later
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
| Relation | verb (controlled vocabulary), from/to, cardinality, join keys, origin, temporal-validity flag |
| Metric | formula, grain, additivity, unit, canonical filters, thresholds (glossary-editable) |
| Action | per-entity verbs agents may execute, approval (automatic/required), what it writes |
| Slot mapping | element → source system declaration; fill status (static in v0) |
| Competency question | question + answer oracle; the acceptance suite |
| Namespace | `core-business` vs `<playbook>`; composition rules |

Grouped in prose as **semantic elements** (entity, relation, metric),
**action elements** (action), and **operational elements** (slot mapping,
competency question, namespace).

## Tooling architecture (v0)

- `tools/` is one Go module, package layout:
  - `internal/model/`: the canonical structs (the meta-model). Single source of truth.
  - `internal/schemagen/`: emits `schema/*.yaml` from the model.
  - `internal/lint/`: loads YAML (position-aware, `fs.FS`-generic so a real disk
    directory and an embedded one share one loader), validates structure against
    the model, then applies semantic rules (composition-not-fork, verb vocabulary,
    grain/additivity presence, tier presence, cross-references resolve).
  - `internal/coverage/`: static slot-mapping completeness report.
  - `internal/compile/`: emitters; v0 ships `mermaid` only.
  - `internal/embedded/`: the public core playbooks, the `kbf-authoring` skill,
    and the prose spec, baked into the binary (see Boundaries below); mirrored
    from the repo root into `data/` by `scripts/embedsync`, never hand-edited.
  - `cmd/kbf/`: cobra wiring: `lint`, `coverage`, `compile`, `schema` (regen),
    `vendor`, `skill install`, `docs`, `init`.
- Lint pipeline: parse → structural validation → semantic rules → render
  (lipgloss table or `--format json` for agent consumption). **JSON output is a
  stable interface**: authoring agents parse it.
- Conformance runner is a thin Go test that walks `conformance/`, but fixtures
  never depend on Go: each case is `input/` + `expect.yaml` (ok | errors with
  rule ids).

## Playbook anatomy (what the linter understands)

```
<playbook>/
  manifest.yaml        # name, version, spec version, builds-on, layer
  ontology/            # *.yaml: entities, relations, metrics, actions
  evals/               # competency questions
  install/             # slot-mapping templates, defaults
```

Workflows, agents, surfaces folders are reserved in the playbook format spec
but not linted in v0 (config-phase scope).

## Boundaries

- `spec/`, `schema/`, `playbooks/`, `examples/` never reference engines,
  databases, vendors, clients, or tooling internals.
- `tools/` may embed the *public* core playbooks (`playbooks/core-*`), the
  `kbf-authoring` skill, and the prose spec (`spec/*.md`,
  `spec/primitives/*.md`), strictly as a composition-resolution fallback:
  a local path of the same manifest name always overrides the embedded
  copy, never the reverse, and embedded content is a convenience default,
  never a privileged or hidden source of truth (`kbf vendor` always
  materializes it to a real, inspectable, editable directory on request).
  `tools/` may not embed a vertical, an example, or anything not already
  public in this repository: the fallback mirrors what's already here, it
  never becomes a second, parallel source of content. Evolved
  2026-08-17 from "tools/ may not embed playbook content; playbooks are
  always inputs" (the batteries-included binary needed a local-first,
  never-privileged carve-out that rule didn't have room for); see
  `project-decisions.md`'s "Batteries-included binary" entry for the why
  and the reversal condition.
- Public hygiene scan covers every file in the repo.
