# kozmo-bf – Project Standards

Global technical standards and non-negotiable rules. All specs and
implementations comply. Workflow rules live in `AGENTS.md`; this file governs
what is built.

## Technology Stack

- **Language**: Go, current stable. Single static binary, no runtime deps.
- **YAML**: `goccy/go-yaml` (position-aware errors: every lint message carries file:line).
- **Meta-schema**: canonical Go structs; JSON Schema generated via `invopop/jsonschema`;
  conformance validates against the *published* schema with `santhosh-tekuri/jsonschema`.
- **CLI**: `cobra`; output styling `charmbracelet/lipgloss`.
- **Templates**: `text/template` (stdlib) for emitters (mermaid in v0).
- **Tests**: stdlib `testing` + `rogpeppe/go-internal/testscript` for CLI golden tests.
- **Quality**: `gofmt`, `golangci-lint`. **Release**: `goreleaser` (when public).

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
- **Composition, never fork**: playbooks build on `core-business`,
  directly or transitively; the linter fails a playbook that redefines an
  element already declared in its composition closure.
- **Public hygiene (absolute)**: no client names, no internal project
  references, no private paths, no prices, in any file including fixtures and
  comments. `make boundaries` scans for a named blocklist plus heuristics; a
  violation fails the build. The blocklist lives in `scripts/boundaries.go` and
  is maintained deliberately, never by pattern alone.

## Format Standards (the spec's own rules)

- Authoring format is **YAML**; docs are Markdown with YAML frontmatter
  (`type:` required), OKF-conformant so bundles stay portable.
- Vocabulary: **semantic elements** (entities, relations, metrics) and
  **action elements**. Never "kinetic".
- Relation verbs come from a small controlled vocabulary (target 10-20) defined
  in `playbooks/core-business`; playbooks may propose additions via RFC, not ad hoc.
- Every metric declares grain and additivity. Every entity declares identity
  keys. Every element carries a governance-equivalent field (`tier` for
  entity/metric, `origin` for relation, `approval` for action); the
  vocabulary is kind-specific, see `spec/conventions.md`.
- Spec versioning is independent of tool versioning: `spec-v0.x` tags with
  migration notes; the tool README states which spec versions it understands.

## Quality & Maintainability

- **`make check` is the gate**: gofmt, golangci-lint, go test, `kbf lint` over
  `playbooks/` + `examples/`, conformance suite, schema-freshness,
  embed-freshness, boundaries. Green before any task is marked done.
- **Module bar**: one concern per file, ~150 lines of real logic; split past
  that along the natural seam. No allowlist, no per-file exemption.
- **Error messages are product**: a lint error names the file, line, rule, and
  the fix. Fixtures in `conformance/` assert error *content*, not just failure.
- **Doc comments state the why**: the seam a package owns, not what the code does.
- **Prose style**: no em-dashes in authored docs; colons, commas, parentheses.

## Change Policy

Changes to this document are project-level changes: intentional, explicit,
recorded in `project-decisions.md`. Spec changes beyond typos go through
`rfcs/` once the repo is public.
