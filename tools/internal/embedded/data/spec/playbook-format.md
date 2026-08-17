---
type: spec-doc
---

<!-- DO NOT EDIT: generated from spec/playbook-format.md by scripts/embedsync. Edit the source, then run `make embed-sync`. -->

# Playbook format

A playbook is a runnable package of business capability. On disk it is a
folder with one manifest and three content folders. This document
describes the anatomy; `conventions.md` covers naming and vocabulary, and
each file in `spec/primitives/` covers one element's shape. See
`spec/onboarding.md` for the methodology: the order a playbook actually
gets authored in, not just the shape it ends in.

## Anatomy

```
<playbook>/
  manifest.yaml    # namespace: name, version, spec, builds-on, layer
  ontology/        # *.yaml: entities, relations, metrics, actions
  evals/           # *.yaml: competency questions
  install/         # slots.yaml: the slot-mapping template or fill
```

Every playbook in this repository follows this anatomy exactly (three core
playbooks, `core-business`, `core-operations`, `core-services`, and three
teaching leaves, `cafe-demo`, `studio-demo`, `bistro-demo`); they are the
reference, not just an illustration of it.

### `manifest.yaml`

One flat document, five fields: `name`, `version`, `spec`, `builds-on`,
`layer`. See `spec/primitives/namespace.md` for the full field table and
worked examples across both layers.

### `ontology/`

Entities, relations, metrics, and actions: the semantic and action
elements. The linter accepts two layouts and mixes of the two within one
playbook:

- **One file per entity**, named after the entity (`offering.yaml`,
  `location.yaml`), optionally holding that entity's own relations or
  metrics alongside it as additional YAML documents in the same file.
- **Grouped files per kind** (`relations.yaml`, `metrics.yaml`,
  `actions.yaml`), each a multi-document YAML file: every relation (or
  metric, or action) is a complete, independent `kind:`-tagged document,
  separated by `---`.

`playbooks/core-business` uses one file per entity for its seven
entities, and one grouped file each for relations, metrics, and actions,
because a dozen small relation documents in a dozen separate files would
be harder to scan than one file sorted by verb. Either layout, or a mix,
is valid; `kind:` is what the linter keys on, never the file name or
folder position. A composing playbook's own additions to an inherited
entity (an `examples/cafe-demo/ontology/entities.yaml` synonym fragment,
say) live in whatever file makes sense for *that* playbook; nothing
requires it to share a file name with the playbook it layers onto.

### `evals/`

Competency questions: the acceptance suite for the whole playbook. See
`spec/primitives/competency-question.md`.

### `install/`

`slots.yaml`: the slot-mapping register, one row per attribute slot
declared anywhere in `ontology/`. See `spec/primitives/slot-mapping.md`
for the row shape and the difference between a template (empty `source`)
and a filled install.

## Reserved, not linted in v0

Three more folders are reserved by this format for later use:
`workflows/`, `agents/`, `surfaces/`. A v0 playbook may not have them at
all; `kbf lint` does not look inside them if it does. They exist in the
format now so that a playbook which starts using them later is not a
breaking layout change, only an addition. What each will hold is future
scope, not part of this spec.

## Composition rules

A playbook's `builds-on` names zero or more other playbooks; `kbf`
resolves the full set transitively into that playbook's **composition
closure** (every playbook reachable by following `builds-on`, deduped by
name). This is a DAG, not single-parent inheritance: more than one
immediate parent is normal, not an edge case.

- **`builds-on` is a set of parents, not one.** A playbook can compose more
  than one other playbook at once, and each of those can compose more,
  arbitrarily deep. `examples/bistro-demo` builds on both
  `playbooks/core-operations` and `playbooks/core-services` directly (the
  diamond: both of those build on `playbooks/core-business`, so
  `core-business` is reached through two paths and deduped to one
  instance, not loaded twice). `kbf lint`/`coverage`/`compile` need every
  playbook in the closure passed as an argument, not only the ones a
  playbook names directly.
- **Root-ness is derived, never declared.** Nothing in `kbf` privileges
  `core-business` by name, and there is no `root: true` field to set: a
  playbook is a root simply because its own `builds-on` is `[]`. Any team
  may publish its own core playbook (a `retail-core`, a `healthcare-core`)
  that builds on this repository's `core-business`, or a root of its own
  that builds on nothing here at all; `kbf` resolves whatever closure the
  manifests describe, this repository's own playbooks included only as
  one worked example of the pattern, not the only legal shape of one.
- **The layer taxonomy, and what each may build on.**

  | `layer` | May build on | Name |
  |---|---|---|
  | `core` | other `core` playbooks only (`[]` allowed: that is a root) | must match `^core-` |
  | `vertical` | `core` or `vertical` playbooks, at least one | must not match `^core-` |

  A core playbook's own vocabulary and elements belong only to its own
  composers. `playbooks/core-operations` and `playbooks/core-services`
  both build on `playbooks/core-business` directly; they are siblings, not
  ancestors of each other. A verb `core-operations` mints, or a relation
  it adds, is available to anything that composes `core-operations`
  (`examples/cafe-demo` among them), never to `core-services` or its own
  composers, unless something builds on both directly, the way
  `examples/bistro-demo` does. Both rules (the build-target table and the
  name prefix) are `KBF013`.
- **Composition, never fork.** A composing playbook may add new elements
  freely (a new entity, a new relation on a new entity pair, a new
  metric). It may not redefine an element that already exists anywhere in
  its closure, with one narrow exception: layering only a glossary-eligible
  field (an entity's `synonyms`, a metric's `thresholds`) onto an
  existing element, in a fragment that leaves every other field
  zero-valued, is a glossary edit, not a fork. The linter matches against
  whichever closure member declares the identity: `examples/cafe-demo`'s
  `offering` synonym fragment matches `playbooks/core-business`, reached
  through `core-operations`, because `core-operations` itself never
  redeclares `offering`. Everything else that repeats a closure member's
  element identity fails `KBF008`. Two *different* closure members
  declaring the same identity, with no composing playbook doing the
  redeclaring itself, is a separate case, `KBF003`'s cross-playbook
  variant: composition has no resolution order, so that is always an
  error, not something one silently wins. See "Common mistakes" in
  `spec/primitives/entity.md` and `spec/primitives/metric.md` for the
  fragment shape, and `spec/conventions.md` for the full governance-tier
  picture.
