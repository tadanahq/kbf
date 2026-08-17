---
type: spec-doc
---

# Namespace

A namespace is a playbook's identity and its place in the extension chain,
declared once per playbook in `manifest.yaml`. Every other primitive in the
playbook belongs to the namespace `manifest.yaml` names.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `name` | yes | kebab-case string | The playbook's namespace. What an extending playbook's `extends` refers to. |
| `version` | yes | semver string | The playbook's own version, independent of the spec version it targets. |
| `spec` | yes | `v0` (v0 scope) | Which version of the KBF Ontology Spec this playbook targets. See `spec/versioning.md`. |
| `extends` | yes | playbook name, or `null` | `null` only for a root playbook (`core-universal`). Every other playbook names its immediate parent; the parent may itself extend something else, and `kbf` resolves the whole chain. An unresolved `extends` fails `KBF011`. |
| `layer` | yes | `root \| base \| vertical` | Where this playbook sits in the taxonomy: `root` (the single universal floor), `base` (a core playbook other playbooks build on), or `vertical` (a business-specific leaf). Must be consistent with both `extends` and `name`: `KBF013`. Empty fails `KBF010`; a value outside the three fails `KBF002`. |

`manifest.yaml` is a single flat document, not a `kind:`-discriminated one
like entity, relation, metric, action, or competency-question: a playbook
has exactly one namespace, so there is nothing to discriminate.

## Example

The root playbook, `playbooks/core-universal/manifest.yaml`:

```yaml
name: core-universal
version: 0.1.0
spec: v0
extends: null
layer: root
```

A core playbook extending the root, `playbooks/core-operations/manifest.yaml`:

```yaml
name: core-operations
version: 0.1.0
spec: v0
extends: core-universal
layer: base
```

A leaf two hops from the root, `examples/cafe-demo/manifest.yaml`:

```yaml
name: cafe-demo
version: 0.1.0
spec: v0
extends: core-operations
layer: vertical
```

The three examples above are also the three `layer` values: `core-universal`
is `root`, `core-operations` is `base`, `cafe-demo` is `vertical`. Every
playbook in this repository fits one of them; see "Layers" in
`spec/conventions.md` for the naming rule that goes with each.

`cafe-demo`'s full chain is `cafe-demo` → `core-operations` →
`core-universal`; `kbf lint`/`coverage`/`compile` need every playbook in
the chain passed on the command line (`kbf lint examples/cafe-demo
playbooks/core-operations playbooks/core-universal`), not just the
immediate parent, since resolution isn't a filesystem lookup against this
repo's own `playbooks/` folder.

## Common mistakes

- **Missing or unresolvable `extends`.** `kbf lint` loads every playbook
  path passed on the command line into one set, keyed by `name`, and
  resolves each playbook's `extends` against that set, walking the full
  chain to its root. A playbook whose parent (at any hop, not just the
  immediate one) was not also passed on the command line fails `KBF011`
  with a fix hint to pass that playbook's path too.
- **A non-null `extends` on a root playbook.** `core-universal` is the
  one playbook in this repo with no parent. Giving a root playbook a real
  `extends` value that eventually points back to itself is a cycle:
  `kbf` detects it and fails every playbook on the cycle with `KBF011`,
  rather than hanging or crashing.
- **Assuming the chain stops at one hop.** It doesn't: `extends` is fully
  recursive. `examples/cafe-demo` extends `core-operations`, which
  extends `core-universal`, two hops, and `examples/studio-demo` is the
  same shape through `core-services`. A playbook's own vocabulary,
  overridable elements, and slot-mapping context all come from its whole
  chain, not just its immediate parent: see `spec/conventions.md` and
  `spec/primitives/relation.md`.
- **Reusing a `name` across playbooks.** Namespace collision breaks
  extends-resolution silently (two playbooks both claiming to be
  `core-universal`, say). Playbook names are a flat space; keep them
  distinct.
- **`layer` inconsistent with `extends`.** A `root` playbook must have
  `extends: null`; a `base` or `vertical` playbook must extend a playbook
  whose own `layer` is `root` or `base` (never another `vertical`). Fails
  `KBF013`. Whether the `extends` name resolves at all is checked first,
  as `KBF011`; the layer comparison only runs once there is an actual
  parent to compare against.
- **`layer` inconsistent with `name`.** `root` and `base` playbooks
  (together, "core playbooks": see `spec/conventions.md`) must have a
  name starting with `core-`; `vertical` playbooks must not. `core-widget`
  as a `vertical` fails `KBF013` exactly like `widget` as a `base` does,
  in opposite directions.
