---
name: kozmo-feature-plan
description: Use when the user requests a new feature, a new workflow, or a substantial change to product functionality. Produces a spec capsule before any code is written.
---

## What you are doing

You must always activate **Plan Mode** to design a feature and produce a **spec capsule**. You do not write implementation code in this skill. You think, design, and break the work into tasks. Implementation happens later via the `kozmo-feature-implement` skill.

Before creating or editing the spec capsule files, you must ensure a clear, shared understanding of the requirements by actively conducting an interactive design interview with the user:

1. **Assess Ambiguity**: If the user's initial request is high-level, ambiguous, or lacks design details.
2. **Conduct the Interview**: Initiate the interview in the chat by asking the user targeted, high-impact clarifying questions. Focus on:
3. **Iterate in Chat**: Discuss and refine the design concepts interactively based on the user's feedback.
4. **Commit to Spec**: Only after achieving conceptual alignment in the conversation should you proceed to create and write the spec capsule files.

---

## Feature capsule structure

Every feature lives in its own folder:

```
.agents/specs/<spec-name>/
  overview.md
  design.md
  tasks.md
```

Use kebab-case for `<spec-name>`. Create the folder if it does not exist.

Start each file from its template in `.agents/templates/specs/`. Replace all placeholder content. Remove template comments and instructions. The finished files must read as a completed spec, not a draft.

Create or update **only** these three files (plus `reference/`, see below). Do not scatter feature material elsewhere. Do not reuse another feature's spec.

---

## External API / service features (mandatory pattern)

If the feature integrates an external API or service (a new data source, a third-party platform, etc.):

- Its specification and reference materials, OpenAPI/Swagger files, API docs, exported schemas, live under `.agents/specs/<spec-name>/reference/`.
- `design.md` must reference these files by relative path (e.g. `reference/openapi-bundled.yaml`).
- `design.md` is also where curated API knowledge goes: auth flow, pagination behavior, rate limits or cost constraints, response-shape quirks, and which endpoints this workflow uses and why. Do not create a separate notes file: this knowledge is part of the design.
- The spec folder is the **durable home** for this workflow. Future related work (new endpoints, additional sync stages) appends a new task batch to the same folder's `tasks.md` rather than creating a new spec folder. Reopen, don't duplicate.

---

## `overview.md`: What & Why

Write:

- The user goal / problem being solved
- Scope: what is included, what is explicitly excluded
- User-facing behavior (or, for backend/infra features, the observable outcome)
- Acceptance criteria: clear and testable

Do not include implementation details.

---

## `design.md`: How (architectural)

Describe the solution at an architectural level:

- Main components, modules, packages involved
- Data concepts and relationships (conceptual: no full schemas, no code)
- Key flows and states
- Edge cases and constraints
- For external-API features: the API knowledge described above, plus pointers to `reference/` files

Be precise but implementation-agnostic. Decisions and their rationale belong here. If you are choosing between approaches, state what you chose and why.

---

## `tasks.md`: Executable plan

Tasks are Markdown checkboxes, nothing else.

Task notation (mandatory):

- `[ ]` not started
- `[-]` in progress
- `[x]` done

Each task must:

- Be small and independently verifiable
- Represent a single unit of work
- Include a short "Done means…" line

Order tasks logically (dependencies first). Group into labelled batches if the feature has distinct phases, so future work can append a new batch cleanly.

---

## Exit

When the capsule is complete, summarize the plan briefly and stop. Do not begin implementing: that is the `kozmo-feature-implement` skill's job. Confirm with the user before proceeding to build.
