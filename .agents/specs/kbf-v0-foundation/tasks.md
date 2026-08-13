---
type: spec-tasks
status: in_progress
---
# kbf-v0-foundation – Tasks

## Batch 1: Scaffold + model + schema

- [ ] Go module under `tools/` (module `github.com/kozmo-hq/kozmo-bf/tools`), cobra skeleton with `kbf` root + `schema` command — Done means: `go run ./cmd/kbf --help` lists commands.
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

- [ ] `packages/universal-core` per design (entities, verbs, metrics, actions, questions, manifest, slots) — Done means: `kbf lint packages/universal-core` green.
- [ ] `examples/cafe-demo` per design — Done means: lint green; every primitive used; `kbf coverage` shows mapped + unmapped slots.
- [ ] `kbf coverage` command — Done means: table output correct on cafe-demo, golden-tested.
- [ ] `kbf compile --to mermaid` — Done means: deterministic diagram for cafe-demo, golden-tested, renders on GitHub.

## Batch 4: Conformance + boundaries + docs

- [ ] Conformance fixtures (≥6 valid, ≥6 invalid with expected rule ids) + thin runner — Done means: runner green; breaking a fixture fails with the named rule.
- [ ] `scripts/boundaries.go` + make target + self-test — Done means: planted violation fails; clean tree passes.
- [ ] Spec prose: `spec/index.md`, `spec/primitives/*.md` (7), `spec/package-format.md`, `spec/conventions.md`, `spec/versioning.md` — Done means: every doc's YAML examples lint green via a doc-extraction test.
- [ ] README (thesis + 5-minute tour) + CONTRIBUTING + rfcs/README — Done means: tour commands copy-paste clean on a fresh clone.
