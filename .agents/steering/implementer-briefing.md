# Implementer Briefing

You are building inside `kozmo-bf`, the public home of the KBF Ontology Spec
and its reference tooling. Read, in order, before your first task:

1. `AGENTS.md` (workflow, principles, quality gate)
2. `.agents/steering/project-overview.md` (what this is, v0 definition of done)
3. `.agents/steering/project-architecture.md` (layout, meta-model, tooling design)
4. `.agents/steering/project-standards.md` (stack, non-negotiables)
5. Your capsule under `.agents/specs/<name>/` (overview, design, tasks)

Non-negotiables that bite implementers most often:

- **Public hygiene**: no client names, internal references, private paths, or
  prices anywhere. If an example needs a business, it is the fictional cafe.
- **Schema is generated**: never hand-edit `schema/`.
- **Error messages are product**: file, line, rule id, and the fix, every time.
- **Module bar**: ~150 logic lines per file; split at the natural seam.
- **Tasks are live**: update `tasks.md` after every task, never in a final batch.
- **Verification precedes done**: `make check` green, and demonstrate the
  acceptance criteria on the real deliverable (run the CLI on the real files).

Report tersely: what landed, what's proven, what's blocked.
