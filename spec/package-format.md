---
type: spec-doc
---

# Package format

A KBF package is a folder with one manifest and three content folders. This
document describes the anatomy; `conventions.md` covers naming and
vocabulary, and each file in `spec/primitives/` covers one element's shape.

## Anatomy

```
<package>/
  manifest.yaml    # namespace: name, version, spec, extends
  ontology/        # *.yaml: entities, relations, metrics, actions
  evals/           # *.yaml: competency questions
  install/         # slots.yaml: the slot-mapping template or fill
```

`packages/universal-core` and `examples/cafe-demo` both follow this
anatomy exactly; they are the reference, not just an illustration of it.

### `manifest.yaml`

One flat document, four fields: `name`, `version`, `spec`, `extends`. See
`spec/primitives/namespace.md` for the full field table and both worked
examples (a root package and an extending one).

### `ontology/`

Entities, relations, metrics, and actions: the semantic and action
elements. The linter accepts two layouts and mixes of the two within one
package:

- **One file per entity**, named after the entity (`product.yaml`,
  `location.yaml`), optionally holding that entity's own relations or
  metrics alongside it as additional YAML documents in the same file.
- **Grouped files per kind** (`relations.yaml`, `metrics.yaml`,
  `actions.yaml`), each a multi-document YAML file: every relation (or
  metric, or action) is a complete, independent `kind:`-tagged document,
  separated by `---`.

`packages/universal-core` uses one file per entity for the nine entities,
and one grouped file each for relations, metrics, and actions, because a
dozen small relation documents in twelve separate files would be harder to
scan than one file sorted by verb. Either layout, or a mix, is valid; `kind:`
is what the linter keys on, never the file name or folder position.

### `evals/`

Competency questions: the acceptance suite for the whole package. See
`spec/primitives/competency-question.md`.

### `install/`

`slots.yaml`: the slot-mapping register, one row per attribute slot
declared anywhere in `ontology/`. See `spec/primitives/slot-mapping.md`
for the row shape and the difference between a template (empty `source`)
and a filled install.

## Reserved, not linted in v0

Three more folders are reserved by this format for later use:
`workflows/`, `agents/`, `surfaces/`. A v0 package may not have them at
all; `kbf lint` does not look inside them if it does. They exist in the
format now so that a package which starts using them later is not a
breaking layout change, only an addition. What each will hold is future
scope, not part of this spec.

## Extension rules

- **`extends` chain depth 1 in v0.** A package's `extends` names its
  parent, and only `universal-core` may have `extends: null`. A package
  whose parent itself has a non-null `extends` is a deeper chain than v0
  resolves.
- **Extension, never fork.** An extending package may add new elements
  freely (a new entity, a new relation on a new entity pair, a new
  metric). It may not redefine an element that already exists in its
  parent, with one narrow exception: layering only a glossary-eligible
  field (an entity's `synonyms`, a metric's `thresholds`) onto an
  existing element, in a fragment that leaves every other field
  zero-valued, is a glossary edit, not a fork. Everything else that
  repeats a parent element's identity fails `KBF008`. See "Common
  mistakes" in `spec/primitives/entity.md` and `spec/primitives/metric.md`
  for the fragment shape, and `spec/conventions.md` for the full
  governance-tier picture.
