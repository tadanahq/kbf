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
```

A core playbook extending the root, `playbooks/core-operations/manifest.yaml`:

```yaml
name: core-operations
version: 0.1.0
spec: v0
extends: core-universal
```

A leaf two hops from the root, `examples/cafe-demo/manifest.yaml`:

```yaml
name: cafe-demo
version: 0.1.0
spec: v0
extends: core-operations
```

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
