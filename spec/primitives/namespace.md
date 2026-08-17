---
type: spec-doc
---

# Namespace

A namespace is a playbook's identity and its place in the composition:
declared once per playbook in `manifest.yaml`. Every other primitive in the
playbook belongs to the namespace `manifest.yaml` names.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `name` | yes | kebab-case string | The playbook's namespace. What another playbook's `builds-on` refers to. |
| `version` | yes | semver string | The playbook's own version, independent of the spec version it targets. |
| `spec` | yes | `v0` (v0 scope) | Which version of the KBF Ontology Spec this playbook targets. See `spec/versioning.md`. |
| `builds-on` | yes | list of playbook names (`[]` allowed) | The playbook(s) this one composes, resolved as a closure (`spec/playbook-format.md`), not a single parent. `[]` is legal: for `layer: core` that is what makes a playbook a root (root-ness is derived, never its own value); `layer: vertical` always needs at least one entry. Every name must resolve, and the closure must not cycle: both `KBF011`. |
| `layer` | yes | `core \| vertical` | Where this playbook sits in the taxonomy: `core` (a foundation playbook other playbooks compose) or `vertical` (a business-specific leaf). Must be consistent with both `builds-on` and `name`: `KBF013`. Empty fails `KBF010`; a value outside the two fails `KBF002`. |

`manifest.yaml` is a single flat document, not a `kind:`-discriminated one
like entity, relation, metric, action, or competency-question: a playbook
has exactly one namespace, so there is nothing to discriminate.

## Example

The root playbook, `playbooks/core-universal/manifest.yaml`:

```yaml
name: core-universal
version: 0.1.0
spec: v0
builds-on: []
layer: core
```

A core playbook composing the root, `playbooks/core-operations/manifest.yaml`:

```yaml
name: core-operations
version: 0.1.0
spec: v0
builds-on: [core-universal]
layer: core
```

A vertical two hops from the root, `examples/cafe-demo/manifest.yaml`:

```yaml
name: cafe-demo
version: 0.1.0
spec: v0
builds-on: [core-operations]
layer: vertical
```

The three examples above are also the two `layer` values: `core-universal`
and `core-operations` are both `core` (`core-universal`'s empty `builds-on`
is what makes it the root; `core-operations` composing it is legal since a
core playbook may build on other core playbooks), and `cafe-demo` is
`vertical`. Every playbook in this repository is one or the other; see
"Layers" in `spec/conventions.md` for the naming rule that goes with each.

`cafe-demo`'s full composition closure is `{core-operations, core-universal}`;
`kbf lint`/`coverage`/`compile` need every playbook in the closure passed on
the command line (`kbf lint examples/cafe-demo playbooks/core-operations
playbooks/core-universal`), not just the one it names directly, since
resolution isn't a filesystem lookup against this repo's own `playbooks/`
folder. A playbook can build on more than one other playbook at once (the
diamond case: two core playbooks that both compose the same root, then a
third playbook composing both of them); the closure dedupes a shared
ancestor to one instance rather than loading it twice.

## Common mistakes

- **Missing or unresolvable `builds-on` entry.** `kbf lint` loads every
  playbook path passed on the command line into one set, keyed by `name`,
  and resolves each entry of every playbook's `builds-on` against that set,
  walking the full closure. A playbook whose composition (at any depth, not
  just its own immediate entries) was not also passed on the command line
  fails `KBF011` with a fix hint to pass that playbook's path too.
- **A cycle.** If `builds-on` eventually points back to a playbook already
  on its own composition path, that is a cycle: `kbf` detects it and fails
  every playbook on the cycle with `KBF011`, rather than hanging or
  crashing. A shared ancestor reached through two different paths (the
  diamond case) is not a cycle; only a genuine loop is.
- **Assuming `builds-on` is a single parent.** It isn't: a playbook can
  build on more than one other playbook, and each of those can build on
  more, arbitrarily deep. `examples/cafe-demo` builds on `core-operations`,
  which builds on `core-universal`; `examples/studio-demo` is the same
  shape through `core-services`. A playbook's own vocabulary, overridable
  elements, and slot-mapping context all come from its whole closure, not
  just what it names directly: see `spec/conventions.md` and
  `spec/primitives/relation.md`.
- **Reusing a `name` across playbooks.** Namespace collision breaks
  builds-on resolution silently (two playbooks both claiming to be
  `core-universal`, say). Playbook names are a flat space; keep them
  distinct.
- **The same identity declared by two different playbooks in one closure.**
  Composition has no resolution order: unlike a single linear parent, where
  "nearest ancestor wins" was unambiguous, a closure with more than one
  immediate parent has no such tie-break, so if two composed playbooks both
  declare the same entity, relation, metric, or action, that is always an
  error, `KBF003`, reported against both files.
- **`layer` inconsistent with `builds-on`.** A `core` playbook may only
  build on other `core` playbooks (an empty `builds-on` is fine: that is
  what makes it a root); a `vertical` playbook must build on at least one
  playbook, core or vertical. Fails `KBF013`. Whether each `builds-on` name
  resolves at all is checked first, as `KBF011`; the layer comparison only
  runs once there is an actual playbook to compare against.
- **`layer` inconsistent with `name`.** Core playbooks (see
  `spec/conventions.md`) must have a name starting with `core-`; vertical
  playbooks must not. `core-widget` as a `vertical` fails `KBF013` exactly
  like `widget` as a `core` playbook does, in opposite directions.
