---
type: spec-tasks
status: in_progress
---
# kbf-v0-foundation – Tasks

## Batch 1: Scaffold + model + schema

- [x] Go module under `tools/` (module `github.com/kozmo-hq/kozmo-bf/tools`), cobra skeleton with `kbf` root + `schema` command — `go run ./cmd/kbf --help` lists `schema` (lint lands with Batch 2). Verified.
  - go.mod created, deps fetched (cobra, goccy/go-yaml, invopop/jsonschema, lipgloss). design.md gained an "Implementation clarifications" section resolving tier vocab, KBF007/008 mechanics, extends resolution, schema scope — read that before touching lint or schemagen. Content agent (parallel) has since appended a relation-identity clarification (KBF003/008 key on (name,from,to) for relations) — binding for Batch 2 rules.
- [x] Canonical model structs in `internal/model/` for the seven primitives per design — 7 files (model.go + entity/relation/metric/action/competency/manifest/slots.go), every field has a why-comment. Compiles, gofmt clean.
- [x] `internal/schemagen/` + `kbf schema [--check]` generating `schema/ontology.schema.yaml` + `schema/manifest.schema.yaml` — Doc comments flow into schema `description` via invopop's AddGoComments (had to compute base/dir dynamically so it works from both repo root and tools/; see comment on `modelImportComponents`). Verified: generates clean YAML, `--check` exits 0 when current, exits 1 with a clear "stale" message after a struct-comment edit, exits 0 again after revert.
- [ ] Makefile (`check` = fmt, lint, test, dogfood-lint, conformance, schema-freshness, boundaries) + GitHub Actions workflow mirroring it — Done means: `make check` runs end to end (dogfood steps may be empty-pass until Batch 3).
- [ ] Apache-2.0 LICENSE, .gitignore finalized — Done means: files present, license header policy noted in CONTRIBUTING.

## Batch 2: Linter

- [ ] Position-aware loader (goccy) + manifest/extends resolution — Done means: loader returns elements with file:line, bad YAML yields KBF-coded error not panic.
- [ ] Structural validation against model (KBF001-KBF004, KBF010-KBF011) — Done means: unit tests per rule, error content asserted.
- [ ] Semantic rules (KBF005-KBF009, KBF012) — Done means: unit tests per rule.
- [ ] Renderers: lipgloss human + stable `--format json` — Done means: golden tests for both on a fixture package.

## Batch 3: Content

- [-] `packages/universal-core` per design (entities, verbs, metrics, actions, questions, manifest, slots) — Content complete: manifest (`extends: null`), 9 entity files (organization, location, order, product, customer, employee, shift, supplier, purchase — each meaning/identity/resolution/tier/synonyms(en+es)/attributes; states omitted on customer and supplier to keep it visibly optional), `ontology/relations.yaml` (15 relations, all 12 controlled-vocabulary verbs present at least once per the KBF007 clarification, 3 verbs recur across distinct (name,from,to) pairs, see the new design.md note on relation identity), `ontology/metrics.yaml` (6 metrics, grain+additivity+unit+tier on all, thresholds on labor-cost-ratio and gross-margin), `ontology/actions.yaml` (4 actions with risk auto/confirm, no tier field per the clarification), `evals/competency-questions.yaml` (9, one per entity, `expects` avoids bare-verb ambiguity where a verb recurs), `install/slots.yaml` (26 template rows, one per declared attribute, empty sources). Hand-verified (uv+pyyaml, no Go linter yet): all YAML parses (21 files / 53 docs, 0 errors); every relation from/to resolves to a declared entity; every attribute slot has exactly one install/slots.yaml row, zero drift; no duplicate (name,from,to) relation triples. Done-means (`kbf lint` green) blocked on Batch 2 landing.
- [-] `examples/cafe-demo` per design — Content complete: manifest (`extends: universal-core`), `ontology/entities.yaml` (product synonym override: per the KBF008 clarification this is a minimal fragment — kind+name+synonyms only, every non-glossary field zero-valued, not a full copy — adds "menu-item"), `ontology/relations.yaml` (new client-configured self-relation `location belongs-to location` for future multi-location grouping; distinct (name,from,to) from core's belongs-to, so not a fork), `ontology/metrics.yaml` (labor-cost-ratio threshold-only override, same minimal-fragment shape, 0.35 → 0.32; net-new waste-ratio metric, declared in full), `evals/competency-questions.yaml` (2 questions exercising the new content), `install/slots.yaml` (all 26 slots, mapped to `demopos`/`demobooks` except the 3 `crm.customer-*` slots left unmapped on purpose so a future `kbf coverage` run has both mapped and unmapped rows). Hand-verified: override fragments asserted to carry only {kind,name,glossary-field}; slot set diffed 1:1 against universal-core's 26, zero drift. Done-means (`kbf lint` green, `kbf coverage` mapped+unmapped) blocked on Batch 2/3 tooling.
- [ ] `kbf coverage` command — tooling agent (Go; outside this agent's scope).
- [ ] `kbf compile --to mermaid` — tooling agent (Go; outside this agent's scope).

## Batch 4: Conformance + boundaries + docs

- [ ] Conformance fixtures (≥6 valid, ≥6 invalid with expected rule ids) + thin runner — Done means: runner green; breaking a fixture fails with the named rule.
- [ ] `scripts/boundaries.go` + make target + self-test — Done means: planted violation fails; clean tree passes.
- [-] Spec prose: `spec/index.md`, `spec/primitives/*.md` (7), `spec/package-format.md`, `spec/conventions.md`, `spec/versioning.md` — Content complete, 10 files. index.md frames the 7 primitives as semantic/action/operational elements with a reading order; each primitives doc has purpose, a fields table, one worked YAML example, and a common-mistakes list tied to real rule ids; package-format.md documents the manifest + folder anatomy including the reserved-not-linted `workflows/agents/surfaces`; conventions.md carries the full controlled-verb table (all 12, meaning + which recur), the kind-specific tier table, and the glossary-eligible-field table; versioning.md covers spec-v0.x tags, tool independence, migration-note policy, and names every field v0 leaves opaque and why. All frontmatter is `type: spec-doc`, one required key. Hand-verified: every YAML block across all 11 files (12 blocks total) parses; every canonical example programmatically diffed byte-for-byte against the real file in packages/ or examples/ it claims to be copied from (all match); the two illustrative "wrong" fork examples were removed from entity.md/metric.md after checking a future AST-based doc-extraction test would find them regardless of list-item indentation and could flag content that is deliberately invalid, so only real, lintable fragments remain in every yaml-tagged block. Done-means (doc-extraction test) is the tooling agent's to build; content is ready for it.
- [ ] README (thesis + 5-minute tour) + CONTRIBUTING + rfcs/README — in progress, same agent, next.
