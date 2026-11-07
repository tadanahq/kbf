# AGENTS.md: kbf

The single source of truth for how work happens in this repo. Project intent,
architecture, standards, and decisions live in `.agents/steering/`
(project-overview, project-architecture, project-standards, project-decisions):
read the steering docs before planning anything non-trivial. Specs and skills
reference this file and the steering docs; they never restate either.
Maintainers decide what gets built and in which order; work enters this repo
only through spec capsules.

**This repo is public.** It is the home of KBF, the Kozmo Business Framework:
its specs (the KBF Ontology first, further modules as they land), reference
tooling, and core playbooks. The ontology is KBF's foundation module. Everything
here is written for a stranger: any business or person a file describes is
fictional, every reference is one a reader can follow, and no file carries a
private path or a price. That rule is absolute. `make boundaries` machine-checks
paths and prices; the rest is reviewed by hand.

## Workflow Orchestration

- Plan mode is the default for non-trivial work: design first, write code second.
- Use subagents liberally: fan out research, parallel edits, and wide reads.
- **Reuse implementers across related batches**: when the next batch touches the
  same area and the implementer's context is not near-full, resume that agent
  instead of spawning fresh: orientation (briefing + AGENTS.md + capsule) is paid
  once, not per batch. Spawn fresh when the work is a different area, when the
  batch reviews or contradicts that agent's own prior output (no self-review), or
  for parallel batches. Say which mode when handing the work over.
- Verification precedes "done": never mark work complete without proving it.
- The bar: would a staff engineer approve this? Demand elegance.
- Skip the ceremony for obvious fixes (typos, one-liners): match effort to stakes.

## Core Principles

- Simplicity first: the simplest design that fully solves the problem wins.
- No laziness: fix root causes, not symptoms. No band-aids, no silent fallbacks.
- Senior standards: code as a staff engineer would ship it.
- Minimal impact: the smallest change that is correct; do not gold-plate.
- Reporting: concise, fragments fine. (Code, docs, spec prose keep full, careful
  prose: they are the public product.)

## Feature workflow

- Every feature is a spec capsule under `.agents/specs/<name>/`.
- A capsule is three files: `overview.md` (what/why), `design.md` (how),
  `tasks.md` (plan).
- Run the `kozmo-feature-plan` skill before writing any code.
- `kozmo-feature-implement` builds strictly from the capsule and keeps
  `tasks.md` live.
- Future related work appends a task batch to the same capsule; never duplicate
  folders.
- Specs must comply with `.agents/steering/project-standards.md` and must not
  silently override it.

## Build loop

One repeating cycle: **pick → spec → build → review → log**.

- Work enters this repo only as a spec capsule. **No capsule, no code.**
- **Pick** (maintainers): which capsule is next. **Spec** (maintainers + agent):
  the plan interview, then the capsule; no code before a maintainer approves it.
- **Build** (agent, autonomous): execute `tasks.md` batch by batch, keep it
  live, `make check` green. Stop and ask only on decisions the spec cannot
  answer.
- **Review** (maintainers): acceptance criteria demonstrated on the real deliverable,
  never a proxy.
- **Log** (agent): commits; project-level decisions into
  `.agents/steering/project-decisions.md`; docs update inside the task that
  changed reality, never as a later sweep.

## Architecture (summary; full detail in steering)

- `spec/` prose spec · `schema/` **generated** meta-schemas · `playbooks/`
  ontology content · `examples/` teaching content · `tools/` the Go `kbf` CLI ·
  `conformance/` language-agnostic fixture suite · `rfcs/` public change process.
- **The spec is engine-agnostic.** Databases exist only behind explicit
  interfaces in tooling; never in spec or playbooks.
- **Zero runtime dependencies**: the repo is files-in, files-out. No services,
  no network, no DB.
- **Canonical structs, generated schema**: Go structs in `tools/` are the
  meta-model; `schema/*.yaml` is a committed build artifact; CI fails on drift.
- **Config-phase tooling only (v0)**: the CLI serves an agent authoring and
  validating ontologies. Runtime enforcement (the gate) is roadmap, not v0.

## Quality gate

`make check` is the gate: gofmt, golangci-lint, go test, `kbf lint` over
`playbooks/` and `examples/` (dogfood), conformance suite, spec doc-extraction,
schema-freshness (regenerate, diff must be empty), embed-freshness, boundaries
(path and price scan). Green before any task is marked done.

## Prose style

- Spec and docs: full prose, precise, written for a stranger. No em-dashes; use
  colons, commas, parentheses.
- The vocabulary is **semantic elements** (entities, relations, metrics) and
  **action elements**.
