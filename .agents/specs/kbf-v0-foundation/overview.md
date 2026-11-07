---
type: spec-overview
---
# kbf-v0-foundation – Overview

## Goal

Stand up the public `kbf` repository to its v0 definition of done: the KBF
Ontology Spec readable end to end, the meta-schema generated and published, the
`kbf` CLI validating the repository's own content, and teaching content proving
every primitive. The capsule serves two readers: an authoring agent raising an
ontology through the CLI and a skill, and a person reading the repository for
the first time.

## Scope

- Included: repository scaffolding (Go module, Makefile, CI); spec prose (index,
  seven primitive docs, `spec/playbook-format.md`, conventions, versioning,
  onboarding, CLI reference); canonical model structs and schema generation;
  `kbf lint` (structural and semantic rules, human and `--format json` output);
  `kbf coverage` (static); `kbf compile --to mermaid`; `kbf schema`; `kbf init`,
  `kbf playbooks pin`, `kbf skill install` and `kbf docs`; the core playbooks
  (`playbooks/core-business`, `playbooks/core-operations`,
  `playbooks/core-services`) and their embedding in the binary; the
  `examples/cafe-demo`, `examples/studio-demo` and `examples/bistro-demo`
  teaching playbooks; the `kbf-authoring` skill; conformance fixtures and runner;
  the spec-examples test; the boundaries gate; README; Apache-2.0 LICENSE.
- Excluded (roadmap, not v0): the query gate and any runtime enforcement, SQL
  anything, `kbf compile` targets beyond mermaid, `kbf diff`, live coverage, a
  docs site, release publishing, RFC process content beyond the folder README.

## Observable outcome

In a fresh clone, `make check` is green. `kbf lint examples/cafe-demo` and
`kbf lint playbooks/core-business` pass; introducing a broken field yields a
file:line error naming the rule and the fix. `kbf compile --to mermaid
examples/cafe-demo` renders the demo's ontology map. `schema/*.yaml` exists,
regenerates identically, and teaches the format in an editor via
yaml-language-server. A binary from `go install` runs `kbf init` and `kbf lint`
in an empty directory, with no clone of this repository.

## Acceptance criteria

1. Spec: `spec/index.md` and one doc per primitive, each with a valid YAML
   example; `spec/playbook-format.md` matches what the linter enforces.
2. `kbf lint` catches at minimum: unknown fields, missing identity keys, missing
   grain or additivity on metrics, undeclared relation endpoints, verbs outside
   the vocabulary, a fork of an element in the composition closure, dangling
   cross-references, a missing governance field (`tier`, `origin`, `approval`),
   and a `layer` inconsistent with `builds-on` or `name`. Each finding carries
   rule id, file:line and fix hint.
3. `--format json` output is stable and documented (the agent interface).
4. `core-business`: the seven universal entities, relations within the controlled
   verb vocabulary, at least 6 universal metrics with grain and additivity, at
   least 4 actions with `approval`, competency questions per entity.
5. Demos: `cafe-demo` builds on `core-operations`, `studio-demo` on
   `core-services`, `bistro-demo` on both. Each exercises every primitive across
   its closure, including one glossary threshold and one `client-configured`
   relation of its own.
6. Conformance: at least 12 fixtures (at least 6 valid and 6 invalid, the invalid
   ones with expected rule ids); runner green; fixtures contain no Go.
7. Boundaries: the scan fails the build on private filesystem paths and prices;
   a self-test proves it.
8. README: the thesis in one screen, a quickstart with no clone, and a five-minute
   tour over `cafe-demo`.
9. Distribution: `go install` is the whole install. Core playbooks, the skill and
   the spec resolve from the binary; a local path of the same name always wins.
