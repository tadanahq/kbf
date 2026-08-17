---
type: spec-doc
---

# Playbook format

A playbook is a runnable package of business capability. On disk it is a
folder with one manifest and three content folders. This document
describes the anatomy; `conventions.md` covers naming and vocabulary, and
each file in `spec/primitives/` covers one element's shape.

## Anatomy

```
<playbook>/
  manifest.yaml    # namespace: name, version, spec, extends
  ontology/        # *.yaml: entities, relations, metrics, actions
  evals/           # *.yaml: competency questions
  install/         # slots.yaml: the slot-mapping template or fill
```

Every playbook in this repository follows this anatomy exactly (three core
playbooks, `core-business`, `core-operations`, `core-services`, and two
teaching leaves, `cafe-demo`, `studio-demo`); they are the reference, not
just an illustration of it.

### `manifest.yaml`

One flat document, four fields: `name`, `version`, `spec`, `extends`. See
`spec/primitives/namespace.md` for the full field table and worked
examples (a root playbook and an extending one).

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
folder position. An extending playbook's own additions to an inherited
entity (an `examples/cafe-demo/ontology/entities.yaml` synonym fragment,
say) live in whatever file makes sense for *that* playbook; nothing
requires it to share a file name with the ancestor it layers onto.

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

## Extension rules

- **`extends` is a chain, not a single hop.** A playbook's `extends` names
  its immediate parent; that parent may itself extend something else, and
  `kbf` resolves the whole line back to a root (`extends: null`). This
  repository has two-hop chains today: `examples/cafe-demo` →
  `playbooks/core-operations` → `playbooks/core-business`, and
  `examples/studio-demo` → `playbooks/core-services` →
  `playbooks/core-business`. There is no depth limit; the linter walks
  however far the chain goes and fails with `KBF011` on a cycle rather
  than hanging. `kbf lint`/`coverage`/`compile` need every playbook in the
  chain passed as an argument, not only the immediate parent.
- **The root is a parameter, not a hardcoded playbook.** Nothing in `kbf`
  privileges `core-business` by name: it is a root only because its own
  `manifest.yaml` says `extends: null`, the same way any other
  organization's own root would be. Any team may publish its own core
  playbook (a `retail-core`, a `healthcare-core`) that extends this
  repository's `core-business`, or a root of its own that doesn't extend
  anything here at all; `kbf` resolves whatever chain the manifests
  describe, this repository's own playbooks included only as one worked
  example of the pattern, not as the only legal shape of one.
- **A core playbook's own vocabulary and elements belong only to its own
  descendants.** `playbooks/core-operations` and `playbooks/core-services`
  both extend `playbooks/core-business` directly; they are siblings, not
  ancestors of each other. A verb `core-operations` mints, or a relation
  it adds, is available to anything that extends `core-operations`
  (`examples/cafe-demo` among them), never to `core-services` or its own
  descendants, and vice versa.
- **Extension, never fork.** An extending playbook may add new elements
  freely (a new entity, a new relation on a new entity pair, a new
  metric). It may not redefine an element that already exists anywhere in
  its chain, with one narrow exception: layering only a glossary-eligible
  field (an entity's `synonyms`, a metric's `thresholds`) onto an
  existing element, in a fragment that leaves every other field
  zero-valued, is a glossary edit, not a fork. The linter matches against
  the *nearest* ancestor that declares the identity, wherever in the chain
  that is: `examples/cafe-demo`'s `offering` synonym fragment matches
  `playbooks/core-business`, two hops up, because `core-operations` (its
  immediate parent) never redeclares `offering` itself. Everything else
  that repeats an ancestor element's identity fails `KBF008`. See "Common
  mistakes" in `spec/primitives/entity.md` and `spec/primitives/metric.md`
  for the fragment shape, and `spec/conventions.md` for the full
  governance-tier picture.
