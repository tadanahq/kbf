---
type: spec-overview
---
# kbf-v0-foundation – Overview

## Goal

Stand up the public kozmo-bf repository to its v0 definition of done: the KBF
Ontology Spec readable end to end, the meta-schema generated and published, the
`kbf` CLI validating real content, and teaching content proving every primitive.
The primary user is an authoring agent (Claude Code/Codex via a skill) spinning
up an ontology; the secondary user is a stranger evaluating whether KBF is
serious.

## Scope

- Included: repo scaffolding (Go module, Makefile, CI), spec prose (index,
  seven primitive docs, package-format, conventions, versioning), canonical
  model structs + schema generation, `kbf lint` (structural + semantic rules,
  human and `--format json` output), `kbf coverage` (static), `kbf compile --to
  mermaid`, `packages/universal-core` v0, `examples/cafe-demo`, conformance
  fixtures + runner, boundaries gate, README, Apache-2.0 LICENSE.
- Excluded (roadmap, not v0): the query gate and any runtime enforcement, SQL
  anything, PGQ/dbt emitters, `kbf diff`, live coverage, docs site, goreleaser
  publishing, RFC process content beyond the folder README.

## Observable outcome

A fresh clone: `make check` is green. `kbf lint examples/cafe-demo` and
`kbf lint packages/universal-core` pass; introducing a broken field yields a
file:line error naming the rule and the fix. `kbf compile --to mermaid
examples/cafe-demo` renders the demo's ontology map. `schema/*.yaml` exists,
regenerates identically, and teaches the format in an editor via
yaml-language-server.

## Acceptance criteria

1. Spec: `spec/index.md` + one doc per primitive, each with a valid YAML
   example; package-format doc matches what the linter actually enforces.
2. `kbf lint`: catches at minimum: unknown fields, missing identity keys,
   missing grain/additivity on metrics, undeclared relation endpoints, verbs
   outside the vocabulary, fork of a core element, dangling cross-references,
   missing governance tier. Each with rule id + file:line + fix hint.
3. `--format json` output is stable and documented (the agent interface).
4. universal-core: the 8 base entities, base relations within the controlled
   verb vocabulary, at least 6 universal metrics with grain/additivity, at
   least 4 actions with risk tiers, competency questions per entity.
5. cafe-demo: extends universal-core, uses every primitive at least once,
   including one glossary-tier threshold and one client-configured relation.
6. Conformance: at least 12 fixtures (6 valid / 6 invalid with expected rule
   ids); runner green; fixtures contain no Go.
7. Boundaries: scan fails the build on blocklisted terms; self-test proves it.
8. README: the thesis in one screen, 5-minute tour ending at cafe-demo.
