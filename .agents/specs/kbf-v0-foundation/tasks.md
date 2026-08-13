---
type: spec-tasks
status: in_progress
---
# kbf-v0-foundation – Tasks

## Batch 1: Scaffold + model + schema

- [-] Go module under `tools/` (module `github.com/kozmo-hq/kozmo-bf/tools`), cobra skeleton with `kbf` root + `schema` command — Done means: `go run ./cmd/kbf --help` lists commands.
  - go.mod created, deps fetched (cobra, goccy/go-yaml, invopop/jsonschema, lipgloss). design.md gained an "Implementation clarifications" section resolving tier vocab, KBF007/008 mechanics, extends resolution, schema scope — read that before touching lint or schemagen.
- [ ] Canonical model structs in `internal/model/` for the seven primitives per design — Done means: structs compile, doc comments state each field's why.
- [ ] `internal/schemagen/` + `kbf schema [--check]` generating `schema/ontology.schema.yaml` + `schema/manifest.schema.yaml` — Done means: files generated, `--check` exits non-zero after a struct edit.
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
- [ ] Spec prose: `spec/index.md`, `spec/primitives/*.md` (7), `spec/package-format.md`, `spec/conventions.md`, `spec/versioning.md` — Done means: every doc's YAML examples lint green via a doc-extraction test.
- [ ] README (thesis + 5-minute tour) + CONTRIBUTING + rfcs/README — Done means: tour commands copy-paste clean on a fresh clone.
