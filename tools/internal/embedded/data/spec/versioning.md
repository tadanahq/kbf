---
type: spec-doc
---

<!-- DO NOT EDIT: generated from spec/versioning.md by scripts/embedsync. Edit the source, then run `make embed-sync`. -->

# Versioning

The KBF Ontology Spec, the `kbf` CLI, and any given playbook all version
independently. This document covers how the spec itself versions, what a
playbook's `spec:` field commits it to, and which parts of v0 are
deliberately left opaque for a later version to close.

## Spec versions

While the spec is pre-1.0, its tags are named `spec-v0.x`: each tag is a
snapshot of `spec/` at a point where the meta-model (the shapes in
`spec/primitives/`) is stable enough to build against, even though it may
still change before 1.0. No tag has been cut yet; until the first one,
`spec: v0` means the `spec/` directory on the main branch. A playbook
declares which spec version it targets in its manifest:

```yaml
spec: v0
```

This is independent of the playbook's own `version` field and of the `kbf`
tool's release version. A new primitive or a new convention layered on top
of the existing ones does not break a playbook built against `spec: v0`; a
breaking change to the meta-model does, and until 1.0 it can land in any
`spec-v0.x` tag (see "Migration notes").

## Tool independence

`kbf` states which spec versions it understands in the repository README,
not in this document. The spec does not assume a particular version of the
CLI; a conforming implementation only needs to read `spec/` and the
generated `schema/` files, both of which are plain, versioned artifacts.
`conformance/` exists so an implementation other than `kbf` can prove it
matches the spec without depending on `kbf`'s own code.

## Migration notes

A breaking change to the meta-model ships with a migration note in the
same pull request that makes the change: what broke, and the mechanical
fix for a playbook that hits it. The pull request is the record; there is
no separate changelog to keep in sync by hand. Breaking changes go through
the RFC process in `rfcs/README.md`. Until the spec reaches 1.0, they are
expected between `spec-v0.x` tags and carry no deprecation window and no
dual-field transition period, just the fix applied everywhere in the same
change.

## What v0 leaves opaque, and why

A handful of fields are deliberately unparsed strings in v0, not because
the shape does not matter but because parsing them is work that belongs
after the ontology itself has stabilized, not before:

| Field | Why it is opaque in v0 |
|---|---|
| `Entity.resolution` | Cross-source identity resolution strategies (deterministic key match, fuzzy match, manual bridge) are real, but naming and validating a closed set of them is runtime-adjacent work, out of v0's config-phase scope. |
| `Metric.formula` | Parsing and validating expressions is the same class of work as the query gate: deliberately excluded from v0 (see `README.md`, "Status: v0"). `kbf lint` does not read a formula. |
| `Metric.unit` | A closed unit vocabulary (with conversion rules, eventually) is worth having; v0 documents the common values in `spec/primitives/metric.md`, not an enforced enum. |
| `Entity.attributes[].type` | Same shape of gap as `unit`: a small common set (`text`, `number`, `boolean`, `date`, `timestamp`, `currency`) documented in `spec/primitives/entity.md`, not yet closed by the linter. |

None of these are silent gaps: each is named here and in the primitive doc
that owns the field, so "opaque in v0" is a stated scope boundary, not
something a playbook author has to discover by trial and error against
`kbf lint`. Closing any of them is a spec-version change, tracked the same
way as any other breaking change, once there is a real need behind it
rather than a hypothetical one.
