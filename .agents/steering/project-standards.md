# kbf – Project Standards

Global technical standards and non-negotiable rules. All specs and
implementations comply. Workflow rules live in `AGENTS.md`; this file governs
what is built.

## Technology Stack

- **Language**: Go, current stable. Single static binary, no runtime deps.
- **YAML**: `goccy/go-yaml` (position-aware errors: every lint message carries file:line).
- **Meta-schema**: canonical Go structs; JSON Schema generated via `invopop/jsonschema`.
- **CLI**: `cobra`; output styling `charmbracelet/lipgloss`.
- **Tests**: stdlib `testing`, with golden files for rendered CLI output. Conformance
  runs each fixture through the linter (`lint.Run`) and compares its findings to
  `expect.yaml`.
- **Quality**: `gofmt`, `golangci-lint`. **Distribution**: `go install`.

Specs must not introduce other languages, frameworks, or storage unless approved
as a project-level change.

## Architecture Principles

- **Spec is engine-agnostic**: no database concept appears in `spec/`,
  `schema/`, `playbooks/`, or `examples/`. Engine specificity may exist only in
  tooling behind explicit interfaces (none in v0).
- **Zero runtime dependencies**: files in, files out. No network, no services,
  no DB, anywhere in this repo.
- **Canonical structs, generated schema**: `schema/*.yaml` is a committed build
  artifact of the Go meta-model. Editing `schema/` by hand is a violation; CI
  regenerates and fails on drift.
- **Conformance is data**: fixtures are plain YAML (valid/invalid + expected
  outcome), runnable by any implementation. The runner is thin.
- **Composition, never fork**: a playbook composes others through `builds-on`;
  the linter fails a playbook that redefines an element already declared in its
  composition closure.
- **Public hygiene (absolute)**: any business or person a file describes is
  fictional, every reference is one a reader can follow, and no file carries a
  private path or a price, fixtures and comments included. `make boundaries`
  scans for private filesystem paths and prices, and a hit fails the build; the
  rest is reviewed by hand.

## Format Standards (the spec's own rules)

- Authoring format is **YAML**; docs are Markdown with YAML frontmatter
  (`type:` required).
- Vocabulary: **semantic elements** (entities, relations, metrics) and
  **action elements**.
- Relation verbs come from a small controlled vocabulary (target 10-20) seeded
  by `playbooks/core-business`. A playbook reuses a verb from its closure, or
  mints one on a relation touching an entity it declares itself (`KBF007`).
  Adding a verb to a core playbook in this repository goes through an RFC.
- Every metric declares grain and additivity. Every entity declares identity
  keys. Every element carries a governance-equivalent field (`tier` for
  entity/metric, `origin` for relation, `approval` for action); the
  vocabulary is kind-specific, see `spec/conventions.md`.
- Spec versioning is independent of tool versioning: `spec-v0.x` tags with
  migration notes; the tool README states which spec versions it understands.

## Quality & Maintainability

- **`make check` is the gate**: gofmt, golangci-lint, go test, `kbf lint` over
  `playbooks/` + `examples/`, conformance suite, spec doc-extraction,
  schema-freshness, embed-freshness, boundaries. Green before any task is
  marked done.
- **Module bar**: one concern per file, ~150 lines of real logic; split past
  that along the natural seam. No allowlist, no per-file exemption.
- **Error messages are product**: a lint error names the file, line, rule, and
  the fix. Go tests assert message and fix content; fixtures in `conformance/`
  assert rule ids.
- **Doc comments state the why**: the seam a package owns, not what the code does.
- **Prose style**: no em-dashes in authored docs; colons, commas, parentheses.

## Change Policy

Changes to this document are project-level changes: intentional, explicit,
recorded in `project-decisions.md`. Spec changes beyond typos go through
`rfcs/`.
