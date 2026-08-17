---
type: spec-doc
---

# Namespace

A namespace is a package's identity and its place in the extension chain,
declared once per package in `manifest.yaml`. Every other primitive in the
package belongs to the namespace `manifest.yaml` names.

## Fields

| Field | Required | Shape | Notes |
|---|---|---|---|
| `name` | yes | kebab-case string | The package's namespace. What an extending package's `extends` refers to. |
| `version` | yes | semver string | The package's own version, independent of the spec version it targets. |
| `spec` | yes | `v0` (v0 scope) | Which version of the KBF Ontology Spec this package targets. See `spec/versioning.md`. |
| `extends` | yes | package name, or `null` | `null` only for a root package (`universal-core`). Every other package names its immediate parent; the parent may itself extend something else, and `kbf` resolves the whole chain. An unresolved `extends` fails `KBF011`. |

`manifest.yaml` is a single flat document, not a `kind:`-discriminated one
like entity, relation, metric, action, or competency-question: a package
has exactly one namespace, so there is nothing to discriminate.

## Example

The root package, `packages/universal-core/manifest.yaml`:

```yaml
name: universal-core
version: 0.1.0
spec: v0
extends: null
```

A base package extending the root, `packages/operations-core/manifest.yaml`:

```yaml
name: operations-core
version: 0.1.0
spec: v0
extends: universal-core
```

A leaf two hops from the root, `examples/cafe-demo/manifest.yaml`:

```yaml
name: cafe-demo
version: 0.1.0
spec: v0
extends: operations-core
```

`cafe-demo`'s full chain is `cafe-demo` → `operations-core` →
`universal-core`; `kbf lint`/`coverage`/`compile` need every package in
the chain passed on the command line (`kbf lint examples/cafe-demo
packages/operations-core packages/universal-core`), not just the
immediate parent, since resolution isn't a filesystem lookup against this
repo's own `packages/` folder.

## Common mistakes

- **Missing or unresolvable `extends`.** `kbf lint` loads every package
  path passed on the command line into one set, keyed by `name`, and
  resolves each package's `extends` against that set, walking the full
  chain to its root. A package whose parent (at any hop, not just the
  immediate one) was not also passed on the command line fails `KBF011`
  with a fix hint to pass that package's path too.
- **A non-null `extends` on a root package.** `universal-core` is the
  one package in this repo with no parent. Giving a root package a real
  `extends` value that eventually points back to itself is a cycle:
  `kbf` detects it and fails every package on the cycle with `KBF011`,
  rather than hanging or crashing.
- **Assuming the chain stops at one hop.** It doesn't: `extends` is fully
  recursive. `examples/cafe-demo` extends `operations-core`, which
  extends `universal-core`, two hops, and `examples/studio-demo` is the
  same shape through `services-core`. A package's own vocabulary,
  overridable elements, and slot-mapping context all come from its whole
  chain, not just its immediate parent: see `spec/conventions.md` and
  `spec/primitives/relation.md`.
- **Reusing a `name` across packages.** Namespace collision breaks
  extends-resolution silently (two packages both claiming to be
  `universal-core`, say). Package names are a flat space; keep them
  distinct.
